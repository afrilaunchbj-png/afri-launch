package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/assets"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// AssetHandler expose les assets générés et leur téléchargement.
type AssetHandler struct {
	svc *assets.Service
}

// NewAssetHandler construit le handler d'assets.
func NewAssetHandler(svc *assets.Service) *AssetHandler { return &AssetHandler{svc: svc} }

type assetDTO struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

// List gère GET /projects/{id}/assets.
func (h *AssetHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	items, err := h.svc.List(r.Context(), projectID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]assetDTO, 0, len(items))
	for _, a := range items {
		out = append(out, toAssetDTO(a))
	}
	writeData(w, http.StatusOK, out)
}

// Download gère GET /assets/{id}/download.
func (h *AssetHandler) Download(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	asset, data, err := h.svc.Download(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+asset.Filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func toAssetDTO(a domain.Asset) assetDTO {
	return assetDTO{
		ID:          a.ID,
		Kind:        a.Kind,
		Filename:    a.Filename,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
		CreatedAt:   a.CreatedAt,
	}
}
