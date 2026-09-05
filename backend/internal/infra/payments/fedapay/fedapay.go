// Package fedapay encapsule TOUTE la logique FedaPay (transactions + token
// de checkout + webhooks) derrière le port PaymentProvider.
// Docs : https://docs.fedapay.com/ (API v1).
package fedapay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const defaultAPIURL = "https://api.fedapay.com/v1"

// FedaPay implémente port.PaymentProvider.
type FedaPay struct {
	secretKey string
	apiURL    string
	storeName string
	http      *http.Client
}

// New construit le provider FedaPay. apiURL vide = live
// (sandbox : https://sandbox-api.fedapay.com/v1).
func New(secretKey, apiURL, storeName string) *FedaPay {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	if storeName == "" {
		storeName = "AfriLaunch"
	}
	return &FedaPay{secretKey: secretKey, apiURL: apiURL, storeName: storeName, http: &http.Client{Timeout: 30 * time.Second}}
}

func (f *FedaPay) Provider() string { return domain.PaymentProviderFedaPay }

// CreateCheckout crée la transaction puis son token de paiement
// (page hébergée FedaPay).
func (f *FedaPay) CreateCheckout(ctx context.Context, in port.PaymentCheckoutInput) (port.PaymentCheckoutResult, error) {
	txBody := map[string]any{
		"description":        in.Description,
		"amount":             majorAmount(in.AmountMinor, in.Currency),
		"currency":           map[string]any{"iso": in.Currency},
		"merchant_reference": in.PaymentID,
		"callback_url":       in.ReturnURL,
		"custom_metadata":    map[string]string{"payment_id": in.PaymentID},
	}
	if in.CustomerName != "" || in.CustomerEmail != "" {
		customer := map[string]any{}
		if in.CustomerEmail != "" {
			customer["email"] = in.CustomerEmail
		}
		if in.CustomerName != "" {
			customer["firstname"] = in.CustomerName
		}
		txBody["customer"] = customer
	}

	var tx struct {
		Transaction struct {
			ID int64 `json:"id"`
		} `json:"transaction"`
	}
	if err := f.do(ctx, http.MethodPost, "/transactions", txBody, &tx); err != nil {
		return port.PaymentCheckoutResult{}, err
	}
	if tx.Transaction.ID == 0 {
		return port.PaymentCheckoutResult{}, fmt.Errorf("fedapay: transaction sans id")
	}

	var token struct {
		Token string `json:"token"`
		URL   string `json:"url"`
	}
	if err := f.do(ctx, http.MethodPost, fmt.Sprintf("/transactions/%d/token", tx.Transaction.ID), nil, &token); err != nil {
		return port.PaymentCheckoutResult{}, err
	}
	if token.URL == "" {
		return port.PaymentCheckoutResult{}, fmt.Errorf("fedapay: url de checkout vide")
	}
	return port.PaymentCheckoutResult{
		RedirectURL:       token.URL,
		ProviderReference: strconv.FormatInt(tx.Transaction.ID, 10),
	}, nil
}

// VerifyStatus interroge FedaPay (source de vérité) : seul `approved`
// débloque les crédits.
func (f *FedaPay) VerifyStatus(ctx context.Context, providerReference string) (port.PaymentStatusResult, error) {
	var out struct {
		Transaction struct {
			Status string `json:"status"`
		} `json:"transaction"`
	}
	if err := f.do(ctx, http.MethodGet, "/transactions/"+providerReference, nil, &out); err != nil {
		return port.PaymentStatusResult{}, err
	}
	return port.PaymentStatusResult{Status: normalizeStatus(out.Transaction.Status)}, nil
}

// HandleWebhook analyse l'objet Event FedaPay ({name, object{…}}) et
// renvoie l'ID de transaction concerné. La décision finale est
// reconfirmée via VerifyStatus (la doc impose la vérification par API).
func (f *FedaPay) HandleWebhook(_ context.Context, in port.PaymentWebhookInput) (port.PaymentWebhookResult, error) {
	var body struct {
		Name   string `json:"name"`
		Object struct {
			ID json.Number `json:"id"`
		} `json:"object"`
	}
	if err := json.Unmarshal(in.Body, &body); err != nil {
		return port.PaymentWebhookResult{}, fmt.Errorf("fedapay: webhook: %w", err)
	}
	if body.Object.ID == "" {
		return port.PaymentWebhookResult{}, fmt.Errorf("fedapay: webhook sans transaction")
	}
	accepted := body.Name == "transaction.approved" || body.Name == "transaction.declined" ||
		body.Name == "transaction.canceled" || body.Name == "transaction.updated"
	return port.PaymentWebhookResult{ProviderReference: body.Object.ID.String(), Accepted: accepted}, nil
}

// ---------- HTTP ----------

func (f *FedaPay) do(ctx context.Context, method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, f.apiURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.secretKey)
	resp, err := f.http.Do(req)
	if err != nil {
		return fmt.Errorf("fedapay: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("fedapay: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("fedapay: status %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// majorAmount convertit les unités mineures en unités majeures : XOF/XAF
// sont sans décimale, sinon 2 décimales.
func majorAmount(minor int64, currency string) int64 {
	if currency == "XOF" || currency == "XAF" {
		return minor
	}
	return minor / 100
}

func normalizeStatus(status string) string {
	switch status {
	case "approved", "transferred":
		return domain.PaymentSucceeded
	case "declined", "canceled", "expired":
		return domain.PaymentFailed
	default: // pending…
		return domain.PaymentPending
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
