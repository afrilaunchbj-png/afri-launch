package handler

import (
	"net/http"
	"strconv"
	"time"

	"afrilaunch/backend/internal/application/credits"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/apierror"
	"afrilaunch/backend/internal/server/authctx"
)

// CreditHandler expose les endpoints du ledger de crédits.
type CreditHandler struct {
	svc *credits.Service
}

// NewCreditHandler construit le handler de crédits.
func NewCreditHandler(svc *credits.Service) *CreditHandler {
	return &CreditHandler{svc: svc}
}

type creditTransactionDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	Reference *string   `json:"reference,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Summary gère GET /credits.
func (h *CreditHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	summary, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	costs, err := h.svc.GenerationCosts(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"summary": summary,
		"costs":   costs,
	})
}

// Transactions gère GET /credits/transactions.
func (h *CreditHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.svc.Transactions(r.Context(), userID, q.Get("type"), q.Get("operation"), pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}

	out := make([]creditTransactionDTO, 0, len(items))
	for _, t := range items {
		out = append(out, toCreditTransactionDTO(t))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	writeList(w, http.StatusOK, out, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages,
	})
}

// Reserve gère POST /credits/reserve (usage interne/workers à venir).
func (h *CreditHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Amount    int64  `json:"amount"`
		Operation string `json:"operation"`
		Reference string `json:"reference"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	if in.Amount <= 0 || in.Operation == "" || in.Reference == "" {
		writeAPIError(w, r, apierror.Validation("amount, operation et reference sont requis."))
		return
	}
	userID := authctx.UserID(r.Context())
	res, err := h.svc.Reserve(r.Context(), userID, in.Amount, in.Operation, in.Reference)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"id": res.ID, "status": res.Status, "amount": res.Amount,
	})
}

func toCreditTransactionDTO(t domain.CreditTransaction) creditTransactionDTO {
	return creditTransactionDTO{
		ID:        t.ID,
		Type:      t.Type,
		Amount:    t.Amount,
		Operation: t.Operation,
		Status:    t.Status,
		Reference: t.Reference,
		CreatedAt: t.CreatedAt,
	}
}
