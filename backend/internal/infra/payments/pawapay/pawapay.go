// Package pawapay encapsule TOUTE la logique PawaPay (Mobile Money agrégé)
// derrière le port PaymentProvider : checkout hébergé, statut, webhooks.
// Docs : https://docs.pawapay.io/ (API v2).
package pawapay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const defaultAPIURL = "https://api.pawapay.io"

// PawaPay implémente port.PaymentProvider.
type PawaPay struct {
	apiToken string
	apiURL   string
	country  string // ISO 3166-1 alpha-3 (BEN, CIV…) — marchés proposés au client
	http     *http.Client
}

// New construit le provider PawaPay. country vide = "BEN".
func New(apiToken, apiURL, country string) *PawaPay {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	if country == "" {
		country = "BEN"
	}
	return &PawaPay{apiToken: apiToken, apiURL: apiURL, country: country, http: &http.Client{Timeout: 30 * time.Second}}
}

func (p *PawaPay) Provider() string { return domain.PaymentProviderPawaPay }

// CreateCheckout crée un checkout hébergé (l'utilisateur choisit son
// opérateur et paye sur la page PawaPay) puis renvoie l'URL de redirection.
func (p *PawaPay) CreateCheckout(ctx context.Context, in port.PaymentCheckoutInput) (port.PaymentCheckoutResult, error) {
	body := map[string]any{
		"checkoutId":        in.PaymentID, // UUID interne : idempotence côté PawaPay
		"returnUrl":         in.ReturnURL,
		"returnMethod":      "INSTANT",
		"defaultLanguage":   "fr",
		"countries":         []string{p.country},
		"amounts":           []map[string]string{{"country": p.country, "currency": in.Currency, "amount": formatAmount(in.AmountMinor, in.Currency)}},
		"clientReferenceId": in.PaymentID,
		"reason":            map[string]string{"fr": truncate(in.Description, 100)},
	}
	var out struct {
		Status      string `json:"status"`
		RedirectURL string `json:"redirectUrl"`
		CheckoutID  string `json:"checkoutId"`
		Failure     *struct {
			FailureCode    string `json:"failureCode"`
			FailureMessage string `json:"failureMessage"`
		} `json:"failureReason"`
	}
	if err := p.do(ctx, http.MethodPost, "/v2/checkouts", body, &out); err != nil {
		return port.PaymentCheckoutResult{}, err
	}
	if out.Failure != nil {
		return port.PaymentCheckoutResult{}, fmt.Errorf("pawapay: %s: %s", out.Failure.FailureCode, out.Failure.FailureMessage)
	}
	if out.RedirectURL == "" {
		return port.PaymentCheckoutResult{}, fmt.Errorf("pawapay: checkout sans redirectUrl (status %s)", out.Status)
	}
	return port.PaymentCheckoutResult{RedirectURL: out.RedirectURL, ProviderReference: out.CheckoutID}, nil
}

// VerifyStatus interroge PawaPay (source de vérité).
func (p *PawaPay) VerifyStatus(ctx context.Context, providerReference string) (port.PaymentStatusResult, error) {
	var out struct {
		Status string `json:"status"`
		Data   *struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := p.do(ctx, http.MethodGet, "/v2/checkouts/"+providerReference, nil, &out); err != nil {
		return port.PaymentStatusResult{}, err
	}
	status := out.Status
	if out.Data != nil && out.Data.Status != "" {
		status = out.Data.Status
	}
	return port.PaymentStatusResult{Status: normalizeStatus(status)}, nil
}

// HandleWebhook analyse le callback PawaPay (JSON : depositId/checkoutId +
// status). La décision finale est toujours reconfirmée via VerifyStatus.
func (p *PawaPay) HandleWebhook(_ context.Context, in port.PaymentWebhookInput) (port.PaymentWebhookResult, error) {
	var body struct {
		DepositID  string `json:"depositId"`
		CheckoutID string `json:"checkoutId"`
		ClientRef  string `json:"clientReferenceId"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(in.Body, &body); err != nil {
		return port.PaymentWebhookResult{}, fmt.Errorf("pawapay: webhook: %w", err)
	}
	ref := body.CheckoutID
	if ref == "" {
		ref = body.ClientRef
	}
	if ref == "" {
		ref = body.DepositID
	}
	if ref == "" {
		return port.PaymentWebhookResult{}, fmt.Errorf("pawapay: webhook sans référence")
	}
	accepted := body.Status != "" && normalizeStatus(body.Status) != domain.PaymentPending
	return port.PaymentWebhookResult{ProviderReference: ref, Accepted: accepted}, nil
}

// ---------- HTTP ----------

func (p *PawaPay) do(ctx context.Context, method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.apiURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	resp, err := p.http.Do(req)
	if err != nil {
		return fmt.Errorf("pawapay: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("pawapay: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("pawapay: status %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// formatAmount rend le montant au format PawaPay : string décimale.
// XOF (zéro décimale) : minor == unités majeures.
func formatAmount(minor int64, currency string) string {
	if currency == "XOF" || currency == "XAF" {
		return fmt.Sprintf("%d", minor)
	}
	return fmt.Sprintf("%d.%02d", minor/100, minor%100)
}

// normalizeStatus mappe les statuts PawaPay sur les statuts internes.
func normalizeStatus(status string) string {
	switch status {
	case "COMPLETED":
		return domain.PaymentSucceeded
	case "FAILED", "CANCELLED", "EXPIRED", "REJECTED":
		return domain.PaymentFailed
	default: // WAITING_PAYMENT, PROCESSING, IN_RECONCILIATION, ACCEPTED, PENDING…
		return domain.PaymentPending
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
