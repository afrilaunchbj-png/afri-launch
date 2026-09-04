package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

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

type ticketMessageDTO struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	IsAdmin    bool      `json:"is_admin"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func toTicketMessageDTO(m domain.TicketMessageView) ticketMessageDTO {
	return ticketMessageDTO{
		ID: m.ID, AuthorID: m.AuthorID, AuthorName: m.AuthorName,
		IsAdmin: m.IsAdmin, Content: m.Content, CreatedAt: m.CreatedAt,
	}
}

type ticketDetailDTO struct {
	Ticket   ticketDTO          `json:"ticket"`
	Messages []ticketMessageDTO `json:"messages"`
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

// GetTicket gère GET /support/tickets/{id} : détail + fil de discussion.
func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ticket, messages, err := h.svc.Detail(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]ticketMessageDTO, 0, len(messages))
	for _, m := range messages {
		out = append(out, toTicketMessageDTO(m))
	}
	writeData(w, http.StatusOK, ticketDetailDTO{Ticket: toTicketDTO(ticket), Messages: out})
}

type replyRequest struct {
	Content string `json:"content"`
}

// Reply gère POST /support/tickets/{id}/messages : réponse de l'utilisateur.
func (h *SupportHandler) Reply(w http.ResponseWriter, r *http.Request) {
	var in replyRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticket, msg, err := h.svc.Reply(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"), in.Content)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, ticketDetailDTO{
		Ticket: toTicketDTO(ticket), Messages: []ticketMessageDTO{toTicketMessageDTO(msg)},
	})
}
