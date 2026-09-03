package server

import (
	"net/http"
	"strings"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/apierror"
	"afrilaunch/backend/internal/server/authctx"
	"afrilaunch/backend/internal/server/handler"
)

// RequireAuth protège une route : l'accès exige un JWT Neon Auth valide
// (en-tête Authorization: Bearer).
func RequireAuth(tv port.TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				handler.WriteError(w, r, apierror.Unauthorized("Authentification requise."))
				return
			}
			user, err := tv.Verify(r.Context(), token)
			if err != nil {
				handler.WriteError(w, r, apierror.Unauthorized("Session invalide ou expirée."))
				return
			}
			next.ServeHTTP(w, r.WithContext(authctx.WithUser(r.Context(), user)))
		})
	}
}

// RequireSuperadmin protège les routes de suivi global : exige un JWT valide
// ET le rôle superadmin dans la table `users` locale.
func RequireSuperadmin(users port.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := authctx.UserID(r.Context())
			if userID == "" {
				handler.WriteError(w, r, apierror.Unauthorized("Authentification requise."))
				return
			}
			user, err := users.GetByID(r.Context(), userID)
			if err != nil || user.Role != domain.RoleSuperadmin {
				handler.WriteError(w, r, apierror.Forbidden("Accès réservé aux administrateurs."))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// bearerToken extrait le token du header Authorization.
func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
