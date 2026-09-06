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
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const defaultAPIURL = "https://api.pawapay.io"

// PawaPay implémente port.PaymentProvider.
type PawaPay struct {
	apiToken string
	apiURL   string
	http     *http.Client
}

// New construit le provider PawaPay.
func New(apiToken, apiURL string) *PawaPay {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &PawaPay{apiToken: apiToken, apiURL: apiURL, http: &http.Client{Timeout: 30 * time.Second}}
}

func (p *PawaPay) Provider() string { return domain.PaymentProviderPawaPay }

// CreateCheckout crée un checkout hébergé (l'utilisateur choisit son
// opérateur et paye sur la page PawaPay) puis renvoie l'URL de redirection.
func (p *PawaPay) CreateCheckout(ctx context.Context, in port.PaymentCheckoutInput) (port.PaymentCheckoutResult, error) {
	if len(in.CountryAmounts) == 0 {
		return port.PaymentCheckoutResult{}, fmt.Errorf("pawapay: aucun pays de paiement actif")
	}
	countries := make([]string, 0, len(in.CountryAmounts))
	amounts := make([]map[string]string, 0, len(in.CountryAmounts))
	for _, c := range in.CountryAmounts {
		countries = append(countries, c.Country)
		amounts = append(amounts, map[string]string{"country": c.Country, "currency": c.Currency, "amount": formatAmount(c.AmountMinor, c.Currency)})
	}

	body := map[string]any{
		"checkoutId":        in.PaymentID, // UUID interne : idempotence côté PawaPay
		"returnUrl":         in.ReturnURL,
		"returnMethod":      "INSTANT",
		"defaultLanguage":   "fr",
		"countries":         countries,
		"amounts":           amounts,
		"clientReferenceId": in.PaymentID,
		"reason":            map[string]string{"fr": pawaReason(in.Description)},
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

// pawaReason assainit la description métier pour l'exigence PawaPay :
// le champ `reason` n'accepte que des caractères ALPHANUMÉRIQUES (ASCII)
// et des ESPACES. Les accents/tirets/pontuations sont translittérés ou
// retirés, puis la chaîne est réduite à 100 caractères.
func pawaReason(s string) string {
	clean := strings.Map(func(r rune) rune {
		switch r {
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'É', 'È', 'Ê', 'Ë':
			return 'E'
		case 'à', 'â', 'ä':
			return 'a'
		case 'À', 'Â', 'Ä':
			return 'A'
		case 'î', 'ï':
			return 'i'
		case 'Î', 'Ï':
			return 'I'
		case 'ô', 'ö':
			return 'o'
		case 'Ô', 'Ö':
			return 'O'
		case 'ù', 'û', 'ü':
			return 'u'
		case 'Ù', 'Û', 'Ü':
			return 'U'
		case 'ç':
			return 'c'
		case 'Ç':
			return 'C'
		}
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == ' ':
			return r
		default:
			// Ponctuation, tirets, emojis… : remplacés par une espace pour
			// éviter de coller les mots (puis espaces multiples réduits).
			return ' '
		}
	}, s)
	clean = strings.Join(strings.Fields(clean), " ")
	if clean == "" {
		clean = "Credit purchase"
	}
	return truncate(clean, 100)
}
