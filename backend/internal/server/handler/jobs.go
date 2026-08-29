package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// JobHandler expose le statut des jobs de génération.
type JobHandler struct {
	jobs port.JobRepository
}

// NewJobHandler construit le handler de jobs.
func NewJobHandler(jobs port.JobRepository) *JobHandler { return &JobHandler{jobs: jobs} }

// jobDTO sérialise un job de génération.
type jobDTO struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`
	Status      string          `json:"status"`
	Error       string          `json:"error,omitempty"`
	Cost        int64           `json:"cost"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

func jobDTOFrom(j domain.GenerationJob) jobDTO {
	return jobDTO{
		ID:          j.ID,
		Kind:        j.Kind,
		Status:      j.Status,
		Error:       j.Error,
		Cost:        j.Cost,
		Result:      j.Result,
		CreatedAt:   j.CreatedAt,
		CompletedAt: j.CompletedAt,
	}
}

// Get gère GET /jobs/{id}.
func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	id := chi.URLParam(r, "id")

	job, err := h.jobs.Get(r.Context(), userID, id)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, jobDTOFrom(job))
}
