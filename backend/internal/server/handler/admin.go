package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/admin"
	"afrilaunch/backend/internal/domain"
)

// AdminHandler expose le suivi global superadmin.
type AdminHandler struct {
	svc *admin.Service
}

// NewAdminHandler construit le handler admin.
func NewAdminHandler(svc *admin.Service) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// Stats gère GET /admin/stats.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}

type adminUserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Users gère GET /admin/users (pagination serveur).
func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.svc.ListUsers(r.Context(), pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminUserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, adminUserDTO{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role, CreatedAt: u.CreatedAt})
	}
	writeList(w, http.StatusOK, out, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

type adminTicketDTO struct {
	ID        string    `json:"id"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Tickets gère GET /admin/tickets (pagination serveur).
func (h *AdminHandler) Tickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tickets, total, err := h.svc.ListTickets(r.Context(), pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminTicketDTO, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, adminTicketDTO{
			ID: t.ID, UserEmail: t.UserEmail, UserName: t.UserName,
			Subject: t.Subject, Message: t.Message, Status: t.Status, CreatedAt: t.CreatedAt,
		})
	}
	writeList(w, http.StatusOK, out, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ResolveTicket gère POST /admin/tickets/{id}/resolve.
func (h *AdminHandler) ResolveTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := h.svc.ResolveTicket(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toTicketDTO(domain.SupportTicket{
		ID: ticket.ID, Subject: ticket.Subject, Message: ticket.Message,
		Status: ticket.Status, CreatedAt: ticket.CreatedAt,
	}))
}
