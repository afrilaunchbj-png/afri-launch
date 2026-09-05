// Package paydunya encapsule TOUTE la logique PayDunya (checkout-invoice,
// confirmation, IPN) derrière le port PaymentProvider.
// Docs : https://developers.paydunya.com/doc/FR/
package paydunya

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Modes d'environnement (l'URL /sandbox-api/ correspond aux clés test_private_…).
const (
	modeTest = "test"
	modeLive = "live"
)

// PayDunya implémente port.PaymentProvider.
type PayDunya struct {
	masterKey  string
	privateKey string
	token      string
	mode       string // test | live
	storeName  string
	// apiBaseOverride est utilisé par les tests pour pointer vers httptest.
	apiBaseOverride string
	http            *http.Client
}

// New construit le provider PayDunya. mode vide = "test".
func New(masterKey, privateKey, token, mode, storeName string) *PayDunya {
	if mode == "" {
		mode = modeTest
	}
	if storeName == "" {
		storeName = "AfriLaunch"
	}
	return &PayDunya{
		masterKey: masterKey, privateKey: privateKey, token: token, mode: mode,
		storeName: storeName, http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *PayDunya) Provider() string { return domain.PaymentProviderPayDunya }

func (p *PayDunya) apiURL() string {
	if p.apiBaseOverride != "" {
		return p.apiBaseOverride
	}
	if p.mode == modeLive {
		return "https://app.paydunya.com/api/v1"
	}
	return "https://app.paydunya.com/sandbox-api/v1"
}

// CreateCheckout crée la checkout-invoice et renvoie l'URL hébergée.
func (p *PayDunya) CreateCheckout(ctx context.Context, in port.PaymentCheckoutInput) (port.PaymentCheckoutResult, error) {
	body := map[string]any{
		"invoice": map[string]any{
			"total_amount": majorAmount(in.AmountMinor, in.Currency),
			"description":  in.Description,
		},
		"store": map[string]any{"name": p.storeName},
		"custom_data": map[string]string{
			"payment_id": in.PaymentID,
		},
		"actions": map[string]string{
			"return_url":   in.ReturnURL,
			"cancel_url":   in.ReturnURL,
			"callback_url": in.WebhookURL,
		},
	}

	var out struct {
		ResponseCode string `json:"response_code"`
		ResponseText string `json:"response_text"`
		Token        string `json:"token"`
	}
	if err := p.do(ctx, http.MethodPost, "/checkout-invoice/create", body, &out); err != nil {
		return port.PaymentCheckoutResult{}, err
	}
	if out.ResponseCode != "00" || out.Token == "" {
		return port.PaymentCheckoutResult{}, fmt.Errorf("paydunya: création checkout: %s", out.ResponseText)
	}
	redirect := out.ResponseText
	if redirect == "" {
		redirect = "https://app.paydunya.com/checkout/invoice/" + out.Token
	}
	return port.PaymentCheckoutResult{RedirectURL: redirect, ProviderReference: out.Token}, nil
}

// VerifyStatus confirme la facture chez PayDunya (source de vérité) :
// response_code "00", hash = SHA-512(master key) et status "completed".
func (p *PayDunya) VerifyStatus(ctx context.Context, providerReference string) (port.PaymentStatusResult, error) {
	var out struct {
		ResponseCode string `json:"response_code"`
		Hash         string `json:"hash"`
		Status       string `json:"status"`
	}
	if err := p.do(ctx, http.MethodGet, "/checkout-invoice/confirm/"+providerReference, nil, &out); err != nil {
		return port.PaymentStatusResult{}, err
	}
	if out.Hash != "" && out.Hash != p.expectedHash() {
		return port.PaymentStatusResult{}, fmt.Errorf("paydunya: hash de confirmation invalide")
	}
	return port.PaymentStatusResult{Status: normalizeStatus(out.Status)}, nil
}

// HandleWebhook analyse l'IPN PayDunya : POST x-www-form-urlencoded avec
// les champs sous la clé data[…] (ou un JSON {token}). Le statut annoncé
// n'est jamais une preuve : le service reconfirmera via VerifyStatus.
func (p *PayDunya) HandleWebhook(_ context.Context, in port.PaymentWebhookInput) (port.PaymentWebhookResult, error) {
	// 1. JSON éventuel.
	var jsonBody struct {
		Token string `json:"token"`
		Data  *struct {
			Token  string `json:"token"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(in.Body, &jsonBody); err == nil {
		if ref := firstNonEmpty(jsonBody.Data.Token, jsonBody.Token); ref != "" {
			return port.PaymentWebhookResult{ProviderReference: ref, Accepted: true}, nil
		}
	}

	// 2. Formulaire urlencodé : data[token]=… / data[status]=…
	if values, err := url.ParseQuery(string(in.Body)); err == nil {
		token := values.Get("data[token]")
		if token == "" {
			token = values.Get("token")
		}
		if token != "" {
			return port.PaymentWebhookResult{ProviderReference: token, Accepted: true}, nil
		}
	}
	return port.PaymentWebhookResult{}, fmt.Errorf("paydunya: webhook sans token")
}

// expectedHash : SHA-512 hex de la master key (contrôle d'origine PayDunya).
func (p *PayDunya) expectedHash() string {
	sum := sha512.Sum512([]byte(p.masterKey))
	return hex.EncodeToString(sum[:])
}

// ---------- HTTP ----------

func (p *PayDunya) do(ctx context.Context, method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.apiURL()+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PAYDUNYA-MASTER-KEY", p.masterKey)
	req.Header.Set("PAYDUNYA-PRIVATE-KEY", p.privateKey)
	req.Header.Set("PAYDUNYA-TOKEN", p.token)
	resp, err := p.http.Do(req)
	if err != nil {
		return fmt.Errorf("paydunya: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("paydunya: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("paydunya: status %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// majorAmount : XOF/XAF sans décimale, sinon 2 décimales.
func majorAmount(minor int64, currency string) int64 {
	if currency == "XOF" || currency == "XAF" {
		return minor
	}
	return minor / 100
}

func normalizeStatus(status string) string {
	switch status {
	case "completed":
		return domain.PaymentSucceeded
	case "cancelled", "failed":
		return domain.PaymentFailed
	default: // pending…
		return domain.PaymentPending
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
