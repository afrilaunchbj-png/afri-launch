package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/apierror"
)

// writeJSON sérialise body en JSON avec le bon Content-Type et le statut donné.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeData enveloppe une ressource unique dans `{ "data": ... }`.
func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

// Pagination décrit les métadonnées de pagination d'une liste.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int64 `json:"totalPages"`
}

// writeList enveloppe une liste et sa pagination.
func writeList(w http.ResponseWriter, status int, data any, p Pagination) {
	writeJSON(w, status, map[string]any{"data": data, "pagination": p})
}

// WriteError expose l'écriture d'erreur (RFC 9457) aux middlewares.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	writeAPIError(w, r, err)
}

// writeAPIError sérialise une erreur en Problem Details (RFC 9457).
// Les erreurs inattendues (statut ≥ 500) sont loggées avec leur cause :
// le client ne reçoit jamais le détail technique.
func writeAPIError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr, ok := err.(*apierror.APIError)
	if !ok {
		apiErr = mapError(err)
		if apiErr.Status >= 500 {
			slog.Error("unexpected handler error", "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()), "err", err)
		}
	}
	apiErr.Instance = r.URL.Path
	writeJSON(w, apiErr.Status, apiErr)
}

// mapError convertit une erreur métier en Problem Details.
func mapError(err error) *apierror.APIError {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apierror.NotFound("La ressource demandée n'existe pas.")
	case errors.Is(err, domain.ErrConflict):
		return apierror.Conflict("Un conflit est survenu avec l'état actuel de la ressource.")
	case errors.Is(err, domain.ErrUnauthorized):
		return apierror.Unauthorized("Identifiants invalides.")
	case errors.Is(err, domain.ErrInvalidToken):
		return apierror.Unauthorized("Session invalide ou expirée.")
	case errors.Is(err, domain.ErrForbidden):
		return apierror.Forbidden("Accès refusé.")
	case errors.Is(err, domain.ErrInvalidInput):
		return apierror.Validation("La requête contient des champs invalides.")
	case errors.Is(err, domain.ErrInsufficient):
		return apierror.Business("Crédits insuffisants pour cette opération.")
	case errors.Is(err, domain.ErrNotConfirmed):
		return apierror.Business("Confirmez l'idée avant de générer les assets.")
	case errors.Is(err, domain.ErrCoverRequired):
		return apierror.Business("Générez et validez d'abord la couverture : elle sert d'identité visuelle à tous les assets.")
	default:
		return apierror.Internal()
	}
}
