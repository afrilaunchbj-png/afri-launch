package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// ideaRepo implémente port.IdeaRepository.
type ideaRepo struct {
	s *Store
}

// NewIdeaRepository construit le repository d'idées.
func NewIdeaRepository(s *Store) *ideaRepo { return &ideaRepo{s: s} }

func (r *ideaRepo) Create(ctx context.Context, idea domain.ProductIdea) (domain.ProductIdea, error) {
	row, err := r.s.q.CreateIdea(ctx, db.CreateIdeaParams{
		UserID:           idea.UserID,
		OpportunityID:    strPtrToUUID(idea.OpportunityID),
		ConversationID:   strPtrToUUID(idea.ConversationID),
		Title:            idea.Title,
		Hook:             idea.Hook,
		Explanation:      idea.Explanation,
		Subtitle:         idea.Subtitle,
		Audience:         idea.Audience,
		Problem:          idea.Problem,
		Promise:          idea.Promise,
		Format:           idea.Format,
		EstimatedPrice:   idea.EstimatedPrice,
		Difficulty:       idea.Difficulty,
		MarketEvidence:   idea.MarketEvidence,
		WhyNow:           idea.WhyNow,
		CompetitiveAngle: idea.CompetitiveAngle,
	})
	if err != nil {
		return domain.ProductIdea{}, err
	}
	return toIdea(row), nil
}

func (r *ideaRepo) Get(ctx context.Context, userID, id string) (domain.ProductIdea, error) {
	row, err := r.s.q.GetIdea(ctx, db.GetIdeaParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.ProductIdea{}, domain.ErrNotFound
		}
		return domain.ProductIdea{}, err
	}
	return toIdea(row), nil
}

func (r *ideaRepo) ListByUser(ctx context.Context, userID string) ([]domain.ProductIdea, error) {
	rows, err := r.s.q.ListIdeasByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductIdea, 0, len(rows))
	for _, row := range rows {
		out = append(out, toIdea(row))
	}
	return out, nil
}

func (r *ideaRepo) ListByOpportunity(ctx context.Context, userID, opportunityID string) ([]domain.ProductIdea, error) {
	rows, err := r.s.q.ListIdeasByOpportunity(ctx, db.ListIdeasByOpportunityParams{UserID: userID, OpportunityID: strPtrToUUID(&opportunityID)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductIdea, 0, len(rows))
	for _, row := range rows {
		out = append(out, toIdea(row))
	}
	return out, nil
}

func (r *ideaRepo) ListByConversation(ctx context.Context, userID, conversationID string) ([]domain.ProductIdea, error) {
	rows, err := r.s.q.ListIdeasByConversation(ctx, db.ListIdeasByConversationParams{UserID: userID, ConversationID: strPtrToUUID(&conversationID)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductIdea, 0, len(rows))
	for _, row := range rows {
		out = append(out, toIdea(row))
	}
	return out, nil
}

func (r *ideaRepo) Select(ctx context.Context, userID, id string) (domain.ProductIdea, error) {
	row, err := r.s.q.SelectIdea(ctx, db.SelectIdeaParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.ProductIdea{}, domain.ErrNotFound
		}
		return domain.ProductIdea{}, err
	}
	return toIdea(row), nil
}

func (r *ideaRepo) Unselect(ctx context.Context, userID, id string) error {
	return r.s.q.UnselectIdea(ctx, db.UnselectIdeaParams{ID: id, UserID: userID})
}

func (r *ideaRepo) UpdateContent(ctx context.Context, idea domain.ProductIdea) (domain.ProductIdea, error) {
	row, err := r.s.q.UpdateIdeaContent(ctx, db.UpdateIdeaContentParams{
		ID:          idea.ID,
		Title:       idea.Title,
		Hook:        idea.Hook,
		Explanation: idea.Explanation,
	})
	if err != nil {
		if isNoRows(err) {
			return domain.ProductIdea{}, domain.ErrNotFound
		}
		return domain.ProductIdea{}, err
	}
	return toIdea(row), nil
}

func (r *ideaRepo) SetStatus(ctx context.Context, userID, id, status string) (domain.ProductIdea, error) {
	row, err := r.s.q.SetIdeaStatus(ctx, db.SetIdeaStatusParams{ID: id, UserID: userID, Status: status})
	if err != nil {
		if isNoRows(err) {
			return domain.ProductIdea{}, domain.ErrNotFound
		}
		return domain.ProductIdea{}, err
	}
	return toIdea(row), nil
}

func toIdea(i db.ProductIdea) domain.ProductIdea {
	return domain.ProductIdea{
		ID:               i.ID,
		UserID:           i.UserID,
		OpportunityID:    uuidPtr(i.OpportunityID),
		ConversationID:   uuidPtr(i.ConversationID),
		Title:            i.Title,
		Hook:             i.Hook,
		Explanation:      i.Explanation,
		Subtitle:         i.Subtitle,
		Audience:         i.Audience,
		Problem:          i.Problem,
		Promise:          i.Promise,
		Format:           i.Format,
		EstimatedPrice:   i.EstimatedPrice,
		Difficulty:       i.Difficulty,
		MarketEvidence:   i.MarketEvidence,
		WhyNow:           i.WhyNow,
		CompetitiveAngle: i.CompetitiveAngle,
		IsSelected:       i.IsSelected,
		Status:           i.Status,
		CreatedAt:        i.CreatedAt,
	}
}
