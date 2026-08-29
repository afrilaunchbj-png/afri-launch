package handler

import (
	"net/http"

	"afrilaunch/backend/internal/application/research"
	"afrilaunch/backend/internal/server/apierror"
	"afrilaunch/backend/internal/server/authctx"
)

// ResearchHandler expose le lancement de la recherche en ligne.
type ResearchHandler struct {
	svc *research.Service
}

// NewResearchHandler construit le handler de recherche.
func NewResearchHandler(svc *research.Service) *ResearchHandler { return &ResearchHandler{svc: svc} }

type startResearchRequest struct {
	Query    string   `json:"query"`
	Sector   string   `json:"sector"`
	Markets  []string `json:"markets"`
	Language string   `json:"language"`
}

// Start gère POST /research.
func (h *ResearchHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())

	var in startResearchRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	if in.Query == "" {
		writeAPIError(w, r, apierror.Validation("query est requis"))
		return
	}

	job, err := h.svc.Start(r.Context(), userID, in.Query, in.Sector, in.Markets, in.Language)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusAccepted, jobDTOFrom(job))
}
