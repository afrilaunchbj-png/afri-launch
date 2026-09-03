package handler

import (
	"net/http"
	"time"

	"afrilaunch/backend/internal/application/preferences"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// PreferenceHandler expose les préférences utilisateur (langue, thème).
type PreferenceHandler struct {
	svc *preferences.Service
}

// NewPreferenceHandler construit le handler de préférences.
func NewPreferenceHandler(svc *preferences.Service) *PreferenceHandler {
	return &PreferenceHandler{svc: svc}
}

type preferenceDTO struct {
	Language  string    `json:"language"`
	Theme     string    `json:"theme"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toPreferenceDTO(p domain.UserPreference) preferenceDTO {
	return preferenceDTO{Language: p.Language, Theme: p.Theme, UpdatedAt: p.UpdatedAt}
}

// Get gère GET /preferences.
func (h *PreferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	pref, err := h.svc.Get(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toPreferenceDTO(pref))
}

type updatePreferenceRequest struct {
	Language *string `json:"language,omitempty"`
	Theme    *string `json:"theme,omitempty"`
}

// Update gère PUT /preferences (champs optionnels : absent = inchangé).
func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in updatePreferenceRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}

	pref, err := h.svc.Update(r.Context(), authctx.UserID(r.Context()), preferences.UpdateInput{
		Language: in.Language,
		Theme:    in.Theme,
	})
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toPreferenceDTO(pref))
}
