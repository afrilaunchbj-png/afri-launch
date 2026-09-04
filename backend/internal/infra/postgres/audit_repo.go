package postgres

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// auditRepo implémente port.AuditRepository (journal append-only).
type auditRepo struct {
	s *Store
}

// NewAuditRepository construit le repository d'audit.
func NewAuditRepository(s *Store) *auditRepo { return &auditRepo{s: s} }

func (r *auditRepo) Log(ctx context.Context, entry domain.AuditLog) error {
	metadata, err := json.Marshal(entry.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}
	entityID := entry.EntityID
	var entityIDPtr *string
	if entityID != "" {
		entityIDPtr = &entityID
	}
	return r.s.q.InsertAuditLog(ctx, db.InsertAuditLogParams{
		UserID:   uuidOrNull(entry.UserID),
		Action:   entry.Action,
		Entity:   entry.Entity,
		EntityID: entityIDPtr,
		Metadata: metadata,
	})
}
