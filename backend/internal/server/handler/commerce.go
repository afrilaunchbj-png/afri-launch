package handler

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/commerce"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/apierror"
	"afrilaunch/backend/internal/server/authctx"
)

// CommerceHandler expose le module commerce (Chariow).
type CommerceHandler struct {
	svc *commerce.Service
}

// NewCommerceHandler construit le handler commerce.
func NewCommerceHandler(svc *commerce.Service) *CommerceHandler { return &CommerceHandler{svc: svc} }

type commerceConnectionDTO struct {
	Provider          string  `json:"provider"`
	Status            string  `json:"status"`
	StoreID           string  `json:"store_id"`
	StoreName         string  `json:"store_name"`
	StoreURL          string  `json:"store_url"`
	StoreCurrency     string  `json:"store_currency"`
	WebhookURL        string  `json:"webhook_url"`
	WebhookConfigured bool    `json:"webhook_configured"`
	ConnectedAt       *string `json:"connected_at"`
	LastSyncAt        *string `json:"last_sync_at"`
}

type commerceProductDTO struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsFree     bool   `json:"is_free"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type commerceProductLinkDTO struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	ExternalProductID string `json:"external_product_id"`
	ExternalName      string `json:"external_product_name"`
	PriceMinor        int64  `json:"price_minor"`
	Currency          string `json:"currency"`
	IsPublic          bool   `json:"is_public"`
	PublicURL         string `json:"public_url"`
}

type commerceSaleDTO struct {
	ID            string `json:"id"`
	ExternalID    string `json:"external_sale_id"`
	Status        string `json:"status"`
	AmountMinor   int64  `json:"amount_minor"`
	Currency      string `json:"currency"`
	BuyerEmail    string `json:"buyer_email"`
	BuyerName     string `json:"buyer_name"`
	ProductLinkID string `json:"product_link_id,omitempty"`
	CreatedAt     string `json:"created_at"`
}

func toConnectionDTO(c domain.CommerceConnection, webhookURL string, webhookConfigured bool) commerceConnectionDTO {
	return commerceConnectionDTO{
		Provider: c.Provider, Status: c.Status, StoreID: c.StoreID,
		StoreName: c.StoreName, StoreURL: c.StoreURL, StoreCurrency: c.StoreCurrency,
		WebhookURL: webhookURL, WebhookConfigured: webhookConfigured,
		ConnectedAt: formatTS(c.ConnectedAt), LastSyncAt: formatTS(c.LastSyncAt),
	}
}

func formatTS(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

// Status gère GET /commerce/status.
func (h *CommerceHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	conn, ok, hasWebhook, err := h.svc.Status(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	if conn.ID == "" {
		writeData(w, http.StatusOK, map[string]any{"connected": false, "connection": nil})
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"connected": ok, "connection": toConnectionDTO(conn, h.webhookURL(r), hasWebhook),
	})
}

// Connect gère POST /commerce/chariow/connect.
func (h *CommerceHandler) Connect(w http.ResponseWriter, r *http.Request) {
	var in struct {
		APIKey        string `json:"api_key"`
		WebhookSecret string `json:"webhook_secret"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	conn, err := h.svc.Connect(r.Context(), authctx.UserID(r.Context()), commerce.ConnectInput{
		APIKey: in.APIKey, WebhookSecret: in.WebhookSecret,
	})
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toConnectionDTO(conn, h.webhookURL(r), strings.TrimSpace(in.WebhookSecret) != ""))
}

// UpdateWebhookSecret gère POST /commerce/chariow/webhook-secret.
func (h *CommerceHandler) UpdateWebhookSecret(w http.ResponseWriter, r *http.Request) {
	var in struct {
		WebhookSecret string `json:"webhook_secret"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	if err := h.svc.UpdateWebhookSecret(r.Context(), authctx.UserID(r.Context()), in.WebhookSecret); err != nil {
		writeCommerceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Disconnect gère POST /commerce/chariow/disconnect.
func (h *CommerceHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Disconnect(r.Context(), authctx.UserID(r.Context())); err != nil {
		writeCommerceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete gère DELETE /commerce/chariow (suppression complète de la connexion).
func (h *CommerceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteConnection(r.Context(), authctx.UserID(r.Context())); err != nil {
		writeCommerceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Products gère GET /commerce/products (produits publiés de la boutique).
func (h *CommerceHandler) Products(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListRemoteProducts(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	out := make([]commerceProductDTO, 0, len(items))
	for _, p := range items {
		out = append(out, toProductDTO(p))
	}
	writeData(w, http.StatusOK, out)
}

// Links gère GET /commerce/links (associations projet ↔ produit).
func (h *CommerceHandler) Links(w http.ResponseWriter, r *http.Request) {
	links, err := h.svc.ListLinks(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toLinksDTO(links, h))
}

// Link gère POST /commerce/links.
func (h *CommerceHandler) Link(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProjectID         string `json:"project_id"`
		ExternalProductID string `json:"external_product_id"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	link, err := h.svc.LinkProduct(r.Context(), authctx.UserID(r.Context()), in.ProjectID, in.ExternalProductID)
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toLinkDTO(link, h))
}

// PublishLink gère PUT /commerce/links/{id}/public.
func (h *CommerceHandler) PublishLink(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Public bool `json:"public"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	link, err := h.svc.SetProductPublic(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"), in.Public)
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toLinkDTO(link, h))
}

// Unlink gère DELETE /commerce/links/{id}.
func (h *CommerceHandler) Unlink(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.UnlinkProduct(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id")); err != nil {
		writeCommerceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Sync gère POST /commerce/sync.
func (h *CommerceHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.SyncSales(r.Context(), authctx.UserID(r.Context())); err != nil {
		writeCommerceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Sales gère GET /commerce/sales.
func (h *CommerceHandler) Sales(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Sales(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]commerceSaleDTO, 0, len(items))
	for _, s := range items {
		out = append(out, commerceSaleDTO{
			ID: s.ID, ExternalID: s.ExternalID, Status: s.Status,
			AmountMinor: s.AmountMinor, Currency: s.Currency,
			BuyerEmail: s.BuyerEmail, BuyerName: s.BuyerName,
			CreatedAt: s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeData(w, http.StatusOK, out)
}

// ---------- Public (acheteur final, sans JWT) ----------

// PublicProductInfo gère GET /commerce/public/{token}.
func (h *CommerceHandler) PublicProductInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.PublicInfo(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"name": info.Name, "price_minor": info.PriceMinor,
		"currency": info.Currency, "is_free": info.IsFree, "store_name": info.StoreName,
	})
}

// PublicCheckout gère POST /commerce/public/{token}/checkout.
func (h *CommerceHandler) PublicCheckout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     struct {
			Number      string `json:"number"`
			CountryCode string `json:"country_code"`
		} `json:"phone"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	res, sale, err := h.svc.PublicCheckout(r.Context(), chi.URLParam(r, "token"), commerce.CheckoutBuyer{
		Email: in.Email, FirstName: in.FirstName, LastName: in.LastName,
		PhoneNumber: in.Phone.Number, PhoneCountry: in.Phone.CountryCode,
		CustomerIP: customerIP(r),
	})
	if err != nil {
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"step": res.Step, "checkout_url": res.CheckoutURL,
		"sale_id": res.SaleID, "sale_status": sale.Status,
	})
}

// Webhook gère POST /commerce/webhook/chariow (public, signé HMAC).
func (h *CommerceHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		writeAPIError(w, r, domain.ErrInvalidInput)
		return
	}
	err = h.svc.HandleWebhook(r.Context(), body, commerce.WebhookHeaders{
		Signature:  r.Header.Get("X-Chariow-Signature"),
		PulseID:    r.Header.Get("X-Pulse-Id"),
		DeliveryID: r.Header.Get("X-Pulse-Delivery-Id"),
		Event:      r.Header.Get("X-Pulse-Event"),
	})
	if err != nil {
		if errors.Is(err, domain.ErrCommerceWebhookInvalid) {
			writeAPIError(w, r, apierror.Unauthorized("Signature webhook invalide."))
			return
		}
		writeCommerceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"received": true})
}

// ---------- Helpers ----------

func customerIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if i := strings.Index(ip, ","); i >= 0 {
			ip = ip[:i]
		}
		return strings.TrimSpace(ip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func toProductDTO(p port.CommerceProduct) commerceProductDTO {
	return commerceProductDTO{
		ID: p.ID, Slug: p.Slug, Name: p.Name, Type: p.Type,
		IsFree: p.IsFree, PriceMinor: p.PriceMinor, Currency: p.Currency,
	}
}

func (h *CommerceHandler) webhookURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/api/v1/commerce/webhook/chariow"
}

func toLinkDTO(l domain.CommerceProductLink, h *CommerceHandler) commerceProductLinkDTO {
	dto := commerceProductLinkDTO{
		ID: l.ID, ProjectID: l.ProjectID, ExternalProductID: l.ExternalProductID,
		ExternalName: l.ExternalName, PriceMinor: l.PriceMinor, Currency: l.Currency,
		IsPublic: l.IsPublic,
	}
	if l.IsPublic && l.PublicToken != "" {
		dto.PublicURL = "/buy/" + l.PublicToken
	}
	return dto
}

func toLinksDTO(links []domain.CommerceProductLink, h *CommerceHandler) []commerceProductLinkDTO {
	out := make([]commerceProductLinkDTO, 0, len(links))
	for _, l := range links {
		out = append(out, toLinkDTO(l, h))
	}
	return out
}

func writeCommerceError(w http.ResponseWriter, r *http.Request, err error) {
	var remote *domain.CommerceRemoteError
	switch {
	case errors.As(err, &remote):
		writeAPIError(w, r, apierror.Business(remote.Message))
	case errors.Is(err, domain.ErrCommerceInvalidCredentials):
		writeAPIError(w, r, apierror.Unauthorized("Clé API Chariow invalide ou révoquée."))
	case errors.Is(err, domain.ErrCommerceSecretMissing):
		writeAPIError(w, r, apierror.Business("Configurez le secret du webhook Pulse Chariow."))
	case errors.Is(err, domain.ErrCommerceConfigMissing):
		writeAPIError(w, r, apierror.Business("Configuration de chiffrement manquante (ENCRYPTION_KEY) — intégration indisponible."))
	case errors.Is(err, domain.ErrCommerceProductUnavailable):
		writeAPIError(w, r, apierror.NotFound("Produit Chariow introuvable ou non publié."))
	case errors.Is(err, domain.ErrCommerceUnavailable):
		writeAPIError(w, r, apierror.Business("Chariow est temporairement indisponible, réessayez."))
	case errors.Is(err, domain.ErrCommerceNotConnected):
		writeAPIError(w, r, apierror.Business("Connectez d'abord votre boutique Chariow."))
	default:
		writeAPIError(w, r, err)
	}
}
