package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/payments"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/apierror"
)

// AdminCountryHandler gère les pays de paiement (superadmin uniquement).
type AdminCountryHandler struct {
	svc *payments.Service
}

// NewAdminCountryHandler construit le handler d'administration des pays.
func NewAdminCountryHandler(svc *payments.Service) *AdminCountryHandler {
	return &AdminCountryHandler{svc: svc}
}

type paymentCountryDTO struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Enabled  bool   `json:"enabled"`
}

// List gère GET /admin/payment-countries.
func (h *AdminCountryHandler) List(w http.ResponseWriter, r *http.Request) {
	countries, err := h.svc.ListPaymentCountries(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]paymentCountryDTO, 0, len(countries))
	for _, c := range countries {
		out = append(out, paymentCountryDTO{Code: c.Code, Name: c.Name, Currency: c.Currency, Enabled: c.Enabled})
	}
	writeData(w, http.StatusOK, out)
}

// Update gère PUT /admin/payment-countries/{code}.
func (h *AdminCountryHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	if err := h.svc.SetPaymentCountryEnabled(r.Context(), chi.URLParam(r, "code"), in.Enabled); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeAPIError(w, r, apierror.NotFound("Pays de paiement introuvable."))
			return
		}
		writeAPIError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
