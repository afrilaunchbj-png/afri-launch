// Package chariow encapsule l'API publique Chariow (commerce) derrière
// port.CommerceProvider. Docs : https://chariow.dev/api-reference
package chariow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

const defaultBaseURL = "https://api.chariow.com/v1"

// Provider implémente port.CommerceProvider.
type Provider struct {
	baseURL string
	http    *http.Client
}

// New construit le provider Chariow (baseURL vide = production).
func New(baseURL string) *Provider {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Provider{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 20 * time.Second}}
}

func (p *Provider) Provider() string { return domain.CommerceProviderChariow }

// GetStore valide une clé API et renvoie la boutique.
func (p *Provider) GetStore(ctx context.Context, apiKey string) (port.CommerceStore, error) {
	var out struct {
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			URL      string `json:"url"`
			Currency string `json:"currency"`
		} `json:"data"`
	}
	if err := p.do(ctx, http.MethodGet, "/store", apiKey, nil, &out); err != nil {
		return port.CommerceStore{}, err
	}
	return port.CommerceStore{ID: out.Data.ID, Name: out.Data.Name, URL: out.Data.URL, Currency: out.Data.Currency}, nil
}

// productJSON est la forme produit retournée par l'API Chariow.
type productJSON struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	IsFree  bool   `json:"is_free"`
	Pricing struct {
		CurrentPrice struct {
			Value    float64 `json:"value"`
			Currency string  `json:"currency"`
		} `json:"current_price"`
	} `json:"pricing"`
}

// ListProducts liste les produits publiés de la boutique.
// L'API renvoie `data` soit en tableau direct, soit enveloppé
// ({data, pagination}) selon la version — on accepte les deux.
func (p *Provider) ListProducts(ctx context.Context, apiKey, cursor string, perPage int) ([]port.CommerceProduct, string, error) {
	path := "/products?per_page=100"
	if perPage > 0 {
		path = fmt.Sprintf("/products?per_page=%d", perPage)
	}
	if cursor != "" {
		path += "&cursor=" + cursor
	}
	var out struct {
		Data       json.RawMessage `json:"data"`
		Pagination struct {
			NextCursor string `json:"next_cursor"`
		} `json:"pagination"`
	}
	if err := p.do(ctx, http.MethodGet, path, apiKey, nil, &out); err != nil {
		return nil, "", err
	}
	next := out.Pagination.NextCursor
	var list []productJSON
	if err := json.Unmarshal(out.Data, &list); err != nil {
		var wrapped struct {
			Data       []productJSON `json:"data"`
			Pagination struct {
				NextCursor string `json:"next_cursor"`
			} `json:"pagination"`
		}
		if err2 := json.Unmarshal(out.Data, &wrapped); err2 != nil {
			return nil, "", fmt.Errorf("chariow: décodage produits: %w", err)
		}
		list = wrapped.Data
		if wrapped.Pagination.NextCursor != "" {
			next = wrapped.Pagination.NextCursor
		}
	}
	products := make([]port.CommerceProduct, 0, len(list))
	for _, pr := range list {
		products = append(products, port.CommerceProduct{
			ID: pr.ID, Slug: pr.Slug, Name: pr.Name, Type: pr.Type, IsFree: pr.IsFree,
			Currency:   pr.Pricing.CurrentPrice.Currency,
			PriceMinor: toMinor(pr.Pricing.CurrentPrice.Value, pr.Pricing.CurrentPrice.Currency),
		})
	}
	return products, next, nil
}

// CreateCheckout initie un checkout hébergé Chariow.
func (p *Provider) CreateCheckout(ctx context.Context, apiKey string, in port.CommerceCheckoutInput) (port.CommerceCheckoutResult, error) {
	body := map[string]any{
		"product_id": in.ProductID,
		"email":      in.BuyerEmail,
		"first_name": in.BuyerFirstName,
		"last_name":  in.BuyerLastName,
		"phone": map[string]string{
			"number":       in.BuyerPhoneNumber,
			"country_code": in.BuyerPhoneCountry,
		},
	}
	if in.RedirectURL != "" {
		body["redirect_url"] = in.RedirectURL
	}
	if in.CustomerIP != "" {
		body["customer_ip"] = in.CustomerIP
	}
	if len(in.Metadata) > 0 {
		body["custom_metadata"] = in.Metadata
	}
	var out struct {
		Data struct {
			Step     string `json:"step"`
			Message  string `json:"message"`
			Purchase *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					Value    float64 `json:"value"`
					Currency string  `json:"currency"`
				} `json:"amount"`
			} `json:"purchase"`
			Payment struct {
				CheckoutURL string `json:"checkout_url"`
			} `json:"payment"`
		} `json:"data"`
	}
	if err := p.do(ctx, http.MethodPost, "/checkout", apiKey, body, &out); err != nil {
		return port.CommerceCheckoutResult{}, err
	}
	res := port.CommerceCheckoutResult{Step: out.Data.Step, CheckoutURL: out.Data.Payment.CheckoutURL}
	if out.Data.Purchase != nil {
		res.SaleID = out.Data.Purchase.ID
		res.Status = out.Data.Purchase.Status
		res.AmountMinor = toMinor(out.Data.Purchase.Amount.Value, out.Data.Purchase.Amount.Currency)
		res.Currency = out.Data.Purchase.Amount.Currency
	}
	return res, nil
}

// ListSales liste les ventes de la boutique.
func (p *Provider) ListSales(ctx context.Context, apiKey, cursor string, perPage int) ([]port.CommerceSaleInfo, string, error) {
	path := "/sales?per_page=100"
	if perPage > 0 {
		path = fmt.Sprintf("/sales?per_page=%d", perPage)
	}
	if cursor != "" {
		path += "&cursor=" + cursor
	}
	var out struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Amount struct {
				Value    float64 `json:"value"`
				Currency string  `json:"currency"`
			} `json:"amount"`
			Customer struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"customer"`
			Metadata map[string]string `json:"custom_metadata"`
			Raw      json.RawMessage   `json:"-"`
		} `json:"data"`
		Pagination struct {
			NextCursor string `json:"next_cursor"`
		} `json:"pagination"`
	}
	if err := p.do(ctx, http.MethodGet, path, apiKey, nil, &out); err != nil {
		return nil, "", err
	}
	sales := make([]port.CommerceSaleInfo, 0, len(out.Data))
	for _, s := range out.Data {
		raw, _ := json.Marshal(s)
		sales = append(sales, port.CommerceSaleInfo{
			ID: s.ID, Status: s.Status,
			AmountMinor: toMinor(s.Amount.Value, s.Amount.Currency), Currency: s.Amount.Currency,
			BuyerEmail: s.Customer.Email, BuyerName: s.Customer.Name,
			Metadata: s.Metadata, Raw: raw,
		})
	}
	return sales, out.Pagination.NextCursor, nil
}

// ---------- HTTP ----------

func (p *Provider) do(ctx context.Context, method, path, apiKey string, body any, out any) error {
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = b
	}
	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("chariow: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return domain.ErrCommerceUnavailable
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("chariow: lecture: %w", err)
	}
	if resp.StatusCode >= 400 {
		return p.mapError(resp.StatusCode, data)
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("chariow: décodage: %w", err)
		}
	}
	return nil
}

func (p *Provider) mapError(status int, data []byte) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.ErrCommerceInvalidCredentials
	case http.StatusNotFound:
		return domain.ErrCommerceProductUnavailable
	case http.StatusTooManyRequests:
		return domain.ErrCommerceUnavailable
	default:
		if status >= 500 {
			return domain.ErrCommerceUnavailable
		}
		// Erreurs 4xx : on expose le message Chariow (sans donnée sensible).
		var env struct {
			Message string `json:"message"`
		}
		msg := ""
		if json.Unmarshal(data, &env) == nil {
			msg = env.Message
		}
		if msg == "" {
			msg = fmt.Sprintf("Chariow a refusé la requête (HTTP %d)", status)
		}
		return &domain.CommerceRemoteError{Status: status, Message: msg}
	}
}

// toMinor convertit un montant décimal Chariow en unités mineures entières.
// XOF/XAF (0 décimale) : la valeur est déjà une unité mineure.
func toMinor(value float64, currency string) int64 {
	if currency == "XOF" || currency == "XAF" {
		return int64(math.Round(value))
	}
	return int64(math.Round(value * 100))
}
