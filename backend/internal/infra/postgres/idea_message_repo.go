package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// ideaMessageRepo implémente port.IdeaMessageRepository.
type ideaMessageRepo struct {
	s *Store
}

// NewIdeaMessageRepository construit le repository de messages d'idées.
func NewIdeaMessageRepository(s *Store) *ideaMessageRepo { return &ideaMessageRepo{s: s} }

func (r *ideaMessageRepo) Create(ctx context.Context, m domain.IdeaMessage) (domain.IdeaMessage, error) {
	row, err := r.s.q.CreateIdeaMessage(ctx, db.CreateIdeaMessageParams{
		IdeaID:  m.IdeaID,
		UserID:  m.UserID,
		Role:    m.Role,
		Content: m.Content,
	})
	if err != nil {
		return domain.IdeaMessage{}, err
	}
	return toIdeaMessage(row), nil
}

func (r *ideaMessageRepo) ListByIdea(ctx context.Context, ideaID string) ([]domain.IdeaMessage, error) {
	rows, err := r.s.q.ListIdeaMessages(ctx, ideaID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.IdeaMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, toIdeaMessage(row))
	}
	return out, nil
}

func toIdeaMessage(row db.IdeaMessage) domain.IdeaMessage {
	return domain.IdeaMessage{
		ID:        row.ID,
		IdeaID:    row.IdeaID,
		UserID:    row.UserID,
		Role:      row.Role,
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
	}
}
