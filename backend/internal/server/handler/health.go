package handler

import (
	"context"
	"net/http"
	"time"

	"afrilaunch/backend/internal/server/apierror"
)

// Pinger vérifie la connectivité d'une dépendance (ex. PostgreSQL).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Health expose les endpoints de santé de l'application.
type Health struct {
	started time.Time
	db      Pinger
}

// NewHealth construit un Health horodaté (db optionnel pour la readiness).
func NewHealth(db Pinger) *Health {
	return &Health{started: time.Now(), db: db}
}

// Liveness indique que le process est vivant.
func (h *Health) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness indique que l'application est prête à servir (DB joignable).
func (h *Health) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := h.db.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Health renvoie un état détaillé (uptime, timestamp).
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"uptime":    time.Since(h.started).String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// NotFound gère les routes inconnues (format Problem Details).
func NotFound(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, r, apierror.NotFound("La ressource demandée n'existe pas."))
}

// MethodNotAllowed gère les méthodes non supportées.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, r, apierror.NotFound("Méthode non autorisée pour cette ressource."))
}
