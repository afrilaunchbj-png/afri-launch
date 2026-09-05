package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/admin"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
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

// listParams extrait la pagination et les filtres communs des listes admin.
func listParams(r *http.Request) (page, pageSize int, filter port.AdminListFilter) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := r.URL.Query()
	filter = port.AdminListFilter{
		Search: query.Get("search"),
		Status: query.Get("status"),
		Role:   query.Get("role"),
	}
	return page, pageSize, filter
}

func writePaginated[T any](w http.ResponseWriter, data []T, total int64, page, pageSize int) {
	writeList(w, http.StatusOK, data, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

type adminUserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Users gère GET /admin/users (pagination + filtres search/role).
func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	users, total, err := h.svc.ListUsers(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminUserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, adminUserDTO{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role, CreatedAt: u.CreatedAt})
	}
	writePaginated(w, out, total, page, pageSize)
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

// Tickets gère GET /admin/tickets (pagination + filtres status/search).
func (h *AdminHandler) Tickets(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	tickets, total, err := h.svc.ListTickets(r.Context(), filter, pageSize, (page-1)*pageSize)
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
	writePaginated(w, out, total, page, pageSize)
}

type adminProjectDTO struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	CreditsConsumed int64     `json:"credits_consumed"`
	CreatedAt       time.Time `json:"created_at"`
	UserEmail       string    `json:"user_email"`
	UserName        string    `json:"user_name"`
}

// Projects gère GET /admin/projects (pagination + filtres status/search).
func (h *AdminHandler) Projects(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	projects, total, err := h.svc.ListProjects(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminProjectDTO, 0, len(projects))
	for _, p := range projects {
		out = append(out, adminProjectDTO{
			ID: p.ID, Title: p.Title, Status: p.Status, CreditsConsumed: p.CreditsConsumed,
			CreatedAt: p.CreatedAt, UserEmail: p.UserEmail, UserName: p.UserName,
		})
	}
	writePaginated(w, out, total, page, pageSize)
}

type adminConversationDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
}

// Conversations gère GET /admin/conversations (pagination + filtre search).
func (h *AdminHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	conversations, total, err := h.svc.ListConversations(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminConversationDTO, 0, len(conversations))
	for _, c := range conversations {
		out = append(out, adminConversationDTO{
			ID: c.ID, Title: c.Title, Status: c.Status, CreatedAt: c.CreatedAt,
			UserEmail: c.UserEmail, UserName: c.UserName,
		})
	}
	writePaginated(w, out, total, page, pageSize)
}

type adminAssetDTO struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	Filename     string    `json:"filename"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	CreatedAt    time.Time `json:"created_at"`
	ProjectTitle string    `json:"project_title"`
	UserEmail    string    `json:"user_email"`
}

// Assets gère GET /admin/assets (pagination + filtre search).
func (h *AdminHandler) Assets(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	assets, total, err := h.svc.ListAssets(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminAssetDTO, 0, len(assets))
	for _, a := range assets {
		out = append(out, adminAssetDTO{
			ID: a.ID, Kind: a.Kind, Filename: a.Filename, ContentType: a.ContentType,
			SizeBytes: a.SizeBytes, CreatedAt: a.CreatedAt, ProjectTitle: a.ProjectTitle, UserEmail: a.UserEmail,
		})
	}
	writePaginated(w, out, total, page, pageSize)
}

type adminJobDTO struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	Cost      int64     `json:"cost"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
}

// Jobs gère GET /admin/jobs (pagination + filtres status/search).
func (h *AdminHandler) Jobs(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	jobs, total, err := h.svc.ListJobs(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminJobDTO, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, adminJobDTO{
			ID: j.ID, Kind: j.Kind, Status: j.Status, Cost: j.Cost,
			CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, UserEmail: j.UserEmail, UserName: j.UserName,
		})
	}
	writePaginated(w, out, total, page, pageSize)
}

type adminCreditTransactionDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UserEmail string    `json:"user_email"`
}

// CreditTransactions gère GET /admin/credit-transactions (pagination + filtres type/search).
func (h *AdminHandler) CreditTransactions(w http.ResponseWriter, r *http.Request) {
	page, pageSize, filter := listParams(r)

	transactions, total, err := h.svc.ListCreditTransactions(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]adminCreditTransactionDTO, 0, len(transactions))
	for _, t := range transactions {
		out = append(out, adminCreditTransactionDTO{
			ID: t.ID, Type: t.Type, Amount: t.Amount, Operation: t.Operation,
			Status: t.Status, CreatedAt: t.CreatedAt, UserEmail: t.UserEmail,
		})
	}
	writePaginated(w, out, total, page, pageSize)
}

// TicketDetail gère GET /admin/tickets/{id} : détail + fil de discussion.
func (h *AdminHandler) TicketDetail(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "id")
	ticket, messages, err := h.svc.TicketDetail(r.Context(), ticketID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticketFiles, messageFiles, err := h.svc.TicketAttachments(r.Context(), ticketID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]ticketMessageDTO, 0, len(messages))
	for _, m := range messages {
		dto := toTicketMessageDTO(m)
		dto.Attachments = toAttachmentDTOs(messageFiles[m.ID])
		out = append(out, dto)
	}
	writeData(w, http.StatusOK, map[string]any{
		"ticket": adminTicketDTO{
			ID: ticket.ID, UserEmail: ticket.UserEmail, UserName: ticket.UserName,
			Subject: ticket.Subject, Message: ticket.Message, Status: ticket.Status, CreatedAt: ticket.CreatedAt,
		},
		"messages":    out,
		"attachments": toAttachmentDTOs(ticketFiles),
	})
}

// ReplyTicket gère POST /admin/tickets/{id}/messages : réponse du support.
func (h *AdminHandler) ReplyTicket(w http.ResponseWriter, r *http.Request) {
	var in replyRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	msg, err := h.svc.ReplyTicket(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"), in.Content)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toTicketMessageDTO(msg))
}

// ResolveTicket gère POST /admin/tickets/{id}/resolve.
func (h *AdminHandler) ResolveTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := h.svc.ResolveTicket(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toTicketDTO(domain.SupportTicket{
		ID: ticket.ID, Subject: ticket.Subject, Message: ticket.Message,
		Status: ticket.Status, CreatedAt: ticket.CreatedAt,
	}))
}

type auditLogDTO struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	EntityID  string         `json:"entity_id"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// AuditLogs gère GET /admin/audit-logs : journal d'activités filtrable
// (userId, action, entity) avec pagination serveur.
func (h *AdminHandler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := r.URL.Query()
	filter := port.AuditFilter{
		UserID: query.Get("userId"),
		Action: query.Get("action"),
		Entity: query.Get("entity"),
	}

	logs, total, err := h.svc.ListAuditLogs(r.Context(), filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]auditLogDTO, 0, len(logs))
	for _, l := range logs {
		out = append(out, auditLogDTO{
			ID: l.ID, UserID: l.UserID, Action: l.Action, Entity: l.Entity,
			EntityID: l.EntityID, Metadata: l.Metadata, CreatedAt: l.CreatedAt,
		})
	}
	writeList(w, http.StatusOK, out, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
