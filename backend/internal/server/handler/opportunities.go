package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/opportunities"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// OpportunityHandler expose les endpoints de recherche d'opportunités.
type OpportunityHandler struct {
	svc *opportunities.Service
}

// NewOpportunityHandler construit le handler d'opportunités.
func NewOpportunityHandler(svc *opportunities.Service) *OpportunityHandler {
	return &OpportunityHandler{svc: svc}
}

type opportunityDTO struct {
	ID         string                   `json:"id"`
	Title      string                   `json:"title"`
	Summary    string                   `json:"summary"`
	Country    string                   `json:"country"`
	Sector     string                   `json:"sector"`
	Language   string                   `json:"language"`
	Difficulty string                   `json:"difficulty"`
	Signal     string                   `json:"signal"`
	Score      int                      `json:"score"`
	Scores     domain.OpportunityScores `json:"scores"`
	Evidence   []domain.Evidence        `json:"evidence"`
	IsSaved    bool                     `json:"is_saved"`
	CreatedAt  time.Time                `json:"created_at"`
}

// List gère GET /opportunities.
func (h *OpportunityHandler) List(w http.ResponseWriter, r *http.Request) {
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

	filter := port.OpportunityFilter{
		Country:    q.Get("country"),
		Sector:     q.Get("sector"),
		Difficulty: q.Get("difficulty"),
		Query:      q.Get("q"),
	}

	items, total, err := h.svc.List(r.Context(), userID, filter, pageSize, (page-1)*pageSize)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}

	out := make([]opportunityDTO, 0, len(items))
	for _, o := range items {
		out = append(out, toOpportunityDTO(o))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	writeList(w, http.StatusOK, out, Pagination{
		Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages,
	})
}

// Save gère POST /opportunities/{id}/save.
func (h *OpportunityHandler) Save(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Save(r.Context(), userID, id); err != nil {
		writeAPIError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unsave gère DELETE /opportunities/{id}/save.
func (h *OpportunityHandler) Unsave(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Unsave(r.Context(), userID, id); err != nil {
		writeAPIError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Filters gère GET /opportunities/filters : facettes disponibles.
func (h *OpportunityHandler) Filters(w http.ResponseWriter, r *http.Request) {
	countries, err := h.svc.Countries(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	sectors, err := h.svc.Sectors(r.Context())
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"countries":    countries,
		"sectors":      sectors,
		"difficulties": []string{domain.DifficultyLow, domain.DifficultyMedium, domain.DifficultyHigh},
	})
}

func toOpportunityDTO(o domain.Opportunity) opportunityDTO {
	return opportunityDTO{
		ID:         o.ID,
		Title:      o.Title,
		Summary:    o.Summary,
		Country:    o.Country,
		Sector:     o.Sector,
		Language:   o.Language,
		Difficulty: o.Difficulty,
		Signal:     o.Signal,
		Score:      o.Score,
		Scores:     o.Scores,
		Evidence:   o.Evidence,
		IsSaved:    o.IsSaved,
		CreatedAt:  o.CreatedAt,
	}
}
