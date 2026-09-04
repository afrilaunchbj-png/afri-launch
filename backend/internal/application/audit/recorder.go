// Package audit enregistre le journal des opérations sensibles.
// Une erreur d'écriture est journalisée sans bloquer l'opération métier
// (l'audit ne doit jamais faire échouer la requête qu'il observe).
package audit

import (
	"context"
	"log/slog"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Recorder écrit dans le journal d'activités.
type Recorder struct {
	repo port.AuditRepository
}

// NewRecorder construit l'enregistreur d'audit.
func NewRecorder(repo port.AuditRepository) *Recorder { return &Recorder{repo: repo} }

// Log enregistre une action sensible (best effort).
func (r *Recorder) Log(ctx context.Context, userID, action, entity, entityID string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	entry := domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Metadata: metadata,
	}
	if err := r.repo.Log(ctx, entry); err != nil {
		slog.Error("audit log write failed", "action", action, "entity", entity, "entity_id", entityID, "err", err)
	}
}
