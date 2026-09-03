package handler

import (
	"net/http"
	"time"

	"afrilaunch/backend/internal/application/support"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// SupportHandler expose les demandes d'assistance utilisateur.
type SupportHandler struct {
	svc *support.Service
}

// NewSupportHandler construit le handler de support.
func NewSupportHandler(svc *support.Service) *SupportHandler {
	return &SupportHandler{svc: svc}
}

type ticketDTO struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toTicketDTO(t domain.SupportTicket) ticketDTO {
	return ticketDTO{ID: t.ID, Subject: t.Subject, Message: t.Message, Status: t.Status, CreatedAt: t.CreatedAt}
}

type createTicketRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Create gère POST /support/tickets.
func (h *SupportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in createTicketRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticket, err := h.svc.Create(r.Context(), authctx.UserID(r.Context()), in.Subject, in.Message)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toTicketDTO(ticket))
}

// ListMine gère GET /support/tickets.
func (h *SupportHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListMine(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]ticketDTO, 0, len(items))
	for _, t := range items {
		out = append(out, toTicketDTO(t))
	}
	writeData(w, http.StatusOK, out)
}
