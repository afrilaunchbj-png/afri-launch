package handler

import (
	"net/http"

	"afrilaunch/backend/internal/application/auth"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// AuthHandler expose les endpoints d'identité (Neon Auth côté backend).
type AuthHandler struct {
	svc            *auth.Service
	welcomeCredits int64
}

// NewAuthHandler construit le handler d'identité.
func NewAuthHandler(svc *auth.Service, welcomeCredits int64) *AuthHandler {
	return &AuthHandler{svc: svc, welcomeCredits: welcomeCredits}
}

type userDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Me gère GET /auth/me : renvoie (et crée si besoin) le profil local.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	identity, _ := authctx.User(r.Context())
	user, err := h.svc.GetOrCreateUser(r.Context(), identity, h.welcomeCredits)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toUserDTO(user))
}

func toUserDTO(u domain.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
	}
}
