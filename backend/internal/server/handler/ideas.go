package handler

import (
	"net/http"

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

func toIdeaDTO(i domain.ProductIdea) ideaDTO {
	return ideaDTO{
		ID:               i.ID,
		OpportunityID:    i.OpportunityID,
		Title:            i.Title,
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
	}
}
