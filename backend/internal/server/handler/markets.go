package handler

import (
	"net/http"

	"afrilaunch/backend/internal/application/port"
)

// MarketHandler expose le référentiel des marchés (filtres).
type MarketHandler struct {
	markets port.MarketRepository
}

// NewMarketHandler construit le handler des marchés.
func NewMarketHandler(markets port.MarketRepository) *MarketHandler {
	return &MarketHandler{markets: markets}
}

type marketDTO struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Language string `json:"language"`
}

// List gère GET /markets.
func (h *MarketHandler) List(w http.ResponseWriter, r *http.Request) {
	markets, err := h.markets.List(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]marketDTO, 0, len(markets))
	for _, m := range markets {
		out = append(out, marketDTO{Code: m.Code, Name: m.Name, Currency: m.Currency, Language: m.Language})
	}
	writeData(w, http.StatusOK, out)
}
