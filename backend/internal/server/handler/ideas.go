package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/ideas"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// IdeaHandler expose la génération et la lecture des idées.
type IdeaHandler struct {
	svc *ideas.Service
}

// NewIdeaHandler construit le handler d'idées.
func NewIdeaHandler(svc *ideas.Service) *IdeaHandler { return &IdeaHandler{svc: svc} }

type ideaDTO struct {
	ID               string  `json:"id"`
	OpportunityID    *string `json:"opportunity_id,omitempty"`
	Title            string  `json:"title"`
	Hook             string  `json:"hook"`
	Explanation      string  `json:"explanation"`
	Subtitle         string  `json:"subtitle"`
	Audience         string  `json:"audience"`
	Problem          string  `json:"problem"`
	Promise          string  `json:"promise"`
	Format           string  `json:"format"`
	EstimatedPrice   string  `json:"estimated_price"`
	Difficulty       string  `json:"difficulty"`
	MarketEvidence   string  `json:"market_evidence"`
	WhyNow           string  `json:"why_now"`
	CompetitiveAngle string  `json:"competitive_angle"`
	IsSelected       bool    `json:"is_selected"`
	Status           string  `json:"status"`
}

// Generate gère POST /opportunities/{id}/ideas.
func (h *IdeaHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	opportunityID := chi.URLParam(r, "id")

	job, err := h.svc.Generate(r.Context(), userID, opportunityID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, jobDTOFrom(job))
}

// List gère GET /ideas (et /opportunities/{id}/ideas).
func (h *IdeaHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	opportunityID := chi.URLParam(r, "id")

	var (
		items []domain.ProductIdea
		err   error
	)
	if opportunityID != "" {
		items, err = h.svc.ListByOpportunity(r.Context(), userID, opportunityID)
	} else {
		items, err = h.svc.List(r.Context(), userID)
	}
	if err != nil {
		writeAPIError(w, r, err)
		return
	}

	out := make([]ideaDTO, 0, len(items))
	for _, i := range items {
		out = append(out, toIdeaDTO(i))
	}
	writeData(w, http.StatusOK, out)
}

// messageDTO sérialise un message de conversation d'idée.
type messageDTO struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func toMessageDTO(m domain.IdeaMessage) messageDTO {
	return messageDTO{ID: m.ID, Role: m.Role, Content: m.Content, CreatedAt: m.CreatedAt}
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

// ListMessages gère GET /ideas/{id}/messages.
func (h *IdeaHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	ideaID := chi.URLParam(r, "id")

	items, err := h.svc.ListMessages(r.Context(), userID, ideaID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]messageDTO, 0, len(items))
	for _, m := range items {
		out = append(out, toMessageDTO(m))
	}
	writeData(w, http.StatusOK, out)
}

// SendMessage gère POST /ideas/{id}/messages.
func (h *IdeaHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	ideaID := chi.URLParam(r, "id")

	var in sendMessageRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}

	job, err := h.svc.SendMessage(r.Context(), userID, ideaID, in.Content)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, jobDTOFrom(job))
}

// Confirm gère POST /ideas/{id}/confirm.
func (h *IdeaHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	ideaID := chi.URLParam(r, "id")

	idea, err := h.svc.Confirm(r.Context(), userID, ideaID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toIdeaDTO(idea))
}

func toIdeaDTO(i domain.ProductIdea) ideaDTO {
	return ideaDTO{
		ID:               i.ID,
		OpportunityID:    i.OpportunityID,
		Title:            i.Title,
		Hook:             i.Hook,
		Explanation:      i.Explanation,
		Subtitle:         i.Subtitle,
		Audience:         i.Audience,
		Problem:          i.Problem,
		Promise:          i.Promise,
		Format:           i.Format,
		EstimatedPrice:   i.EstimatedPrice,
		Difficulty:       i.Difficulty,
		MarketEvidence:   i.MarketEvidence,
		WhyNow:           i.WhyNow,
		CompetitiveAngle: i.CompetitiveAngle,
		IsSelected:       i.IsSelected,
		Status:           i.Status,
	}
}
