package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// AuditRepository écrit le journal des opérations sensibles (append-only).
type AuditRepository interface {
	Log(ctx context.Context, entry domain.AuditLog) error
}

// AuditFilter filtre la lecture du journal d'activités (chaîne vide = pas de filtre).
type AuditFilter struct {
	UserID string
	Action string
	Entity string
}
