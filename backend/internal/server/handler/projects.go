package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/projects"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// ProjectHandler expose les projets et leurs générations.
type ProjectHandler struct {
	svc *projects.Service
}

// NewProjectHandler construit le handler de projets.
func NewProjectHandler(svc *projects.Service) *ProjectHandler { return &ProjectHandler{svc: svc} }

type projectDTO struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Status          string          `json:"status"`
	CreditsConsumed int64           `json:"credits_consumed"`
	OpportunityID   *string         `json:"opportunity_id,omitempty"`
	IdeaID          *string         `json:"idea_id,omitempty"`
	Config          json.RawMessage `json:"config,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type createProjectRequest struct {
	OpportunityID *string `json:"opportunity_id"`
	IdeaID        *string `json:"idea_id"`
	Title         string  `json:"title"`
}

// Create gère POST /projects.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in createProjectRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	userID := authctx.UserID(r.Context())

	project, err := h.svc.Create(r.Context(), userID, in.OpportunityID, in.IdeaID, in.Title)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toProjectDTO(project))
}

// Get gère GET /projects/{id}.
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	project, err := h.svc.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toProjectDTO(project))
}

// List gère GET /projects.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]projectDTO, 0, len(items))
	for _, p := range items {
		out = append(out, toProjectDTO(p))
	}
	writeData(w, http.StatusOK, out)
}

// GenerateEbook gère POST /projects/{id}/ebook.
func (h *ProjectHandler) GenerateEbook(w http.ResponseWriter, r *http.Request) {
	h.dispatch(w, r, h.svc.GenerateEbook)
}

// GenerateCover gère POST /projects/{id}/cover (body optionnel : instructions).
func (h *ProjectHandler) GenerateCover(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Instructions string `json:"instructions"`
	}
	// Body optionnel : absent = première génération sans consignes.
	if r.ContentLength > 0 {
		if err := decodeJSON(w, r, &in); err != nil {
			writeAPIError(w, r, err)
			return
		}
	}
	userID := authctx.UserID(r.Context())
	projectID := chi.URLParam(r, "id")

	job, err := h.svc.GenerateCover(r.Context(), userID, projectID, in.Instructions)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, jobDTOFrom(job))
}

// GeneratePosters gère POST /projects/{id}/posters.
func (h *ProjectHandler) GeneratePosters(w http.ResponseWriter, r *http.Request) {
	h.dispatch(w, r, h.svc.GeneratePosters)
}

// GenerateSalesPage gère POST /projects/{id}/sales-page.
func (h *ProjectHandler) GenerateSalesPage(w http.ResponseWriter, r *http.Request) {
	h.dispatch(w, r, h.svc.GenerateSalesPage)
}

// UpdateConfig gère PUT /projects/{id}/config (identité visuelle + réglages).
func (h *ProjectHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	projectID := chi.URLParam(r, "id")

	var in projects.ConfigInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}

	project, err := h.svc.UpdateConfig(r.Context(), userID, projectID, in)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toProjectDTO(project))
}

func (h *ProjectHandler) dispatch(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, userID, projectID string) (domain.GenerationJob, error)) {
	userID := authctx.UserID(r.Context())
	projectID := chi.URLParam(r, "id")

	job, err := fn(r.Context(), userID, projectID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, jobDTOFrom(job))
}

func toProjectDTO(p domain.Project) projectDTO {
	return projectDTO{
		ID:              p.ID,
		Title:           p.Title,
		Status:          p.Status,
		CreditsConsumed: p.CreditsConsumed,
		OpportunityID:   p.OpportunityID,
		IdeaID:          p.IdeaID,
		Config:          json.RawMessage(p.Config),
		CreatedAt:       p.CreatedAt,
	}
}
