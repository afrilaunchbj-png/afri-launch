package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/payments"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// PaymentHandler expose l'achat de crédits (plans, checkout, suivi) et les
// webhooks des providers (publics, sans JWT).
type PaymentHandler struct {
	svc *payments.Service
}

// NewPaymentHandler construit le handler de paiements.
func NewPaymentHandler(svc *payments.Service) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

type planDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Credits    int64  `json:"credits"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type paymentDTO struct {
	ID          string `json:"id"`
	PlanID      string `json:"plan_id,omitempty"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url,omitempty"`
}

func toPlanDTO(p domain.Plan) planDTO {
	return planDTO{ID: p.ID, Name: p.Name, Credits: p.Credits, PriceMinor: p.PriceMinor, Currency: p.Currency}
}

func toPaymentDTO(p domain.Payment) paymentDTO {
	return paymentDTO{
		ID: p.ID, PlanID: p.PlanID, AmountMinor: p.AmountMinor, Currency: p.Currency,
		Provider: p.Provider, Status: p.Status, CheckoutURL: p.CheckoutURL,
	}
}

// Plans gère GET /payments/plans.
func (h *PaymentHandler) Plans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]planDTO, 0, len(plans))
	for _, p := range plans {
		out = append(out, toPlanDTO(p))
	}
	writeData(w, http.StatusOK, map[string]any{"plans": out, "provider": h.svc.ActiveProvider(), "enabled": h.svc.Enabled()})
}

// Checkout gère POST /payments/checkout {plan_id} → URL de paiement hébergée.
func (h *PaymentHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	var in payments.CheckoutInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	payment, redirectURL, err := h.svc.Checkout(r.Context(), userID, in.PlanID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, map[string]any{
		"payment":      toPaymentDTO(payment),
		"redirect_url": redirectURL,
	})
}

// Get gère GET /payments/{id} (propriétaire) — suivi après redirection.
func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	payment, err := h.svc.GetPayment(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toPaymentDTO(payment))
}

// ListMine gère GET /payments.
func (h *PaymentHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	items, err := h.svc.ListMyPayments(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]paymentDTO, 0, len(items))
	for _, p := range items {
		out = append(out, toPaymentDTO(p))
	}
	writeData(w, http.StatusOK, out)
}

// Sync gère POST /payments/{id}/sync : reconfirmation du statut via le
// provider (retour navigateur avant réception du webhook).
func (h *PaymentHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	payment, err := h.svc.SyncStatus(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toPaymentDTO(payment))
}

// Webhook gère POST /payments/webhook/{provider} — route PUBLIC (sans JWT) :
// notifications serveur-à-serveur des providers. Le corps n'est jamais une
// preuve : le service reconfirme chaque statut par API avant d'agir.
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeAPIError(w, r, domain.ErrInvalidInput)
		return
	}
	headers := map[string]string{
		"content-type": r.Header.Get("Content-Type"),
	}
	for _, key := range []string{"X-Fedapay-Signature", "X-Pawapay-Signature", "Signature", "Signature-Input"} {
		if v := r.Header.Get(key); v != "" {
			headers[key] = v
		}
	}
	payment, err := h.svc.HandleWebhook(r.Context(), provider, port.PaymentWebhookInput{Headers: headers, Body: body})
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"received": true, "status": payment.Status})
}
