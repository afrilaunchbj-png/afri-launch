package server

import (
	"net/http"
	"strings"

	"afrilaunch/backend/internal/application/port"
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

// bearerToken extrait le token du header Authorization.
func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
