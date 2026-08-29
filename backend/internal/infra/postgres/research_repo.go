package postgres

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// researchRepo implémente port.ResearchRepository.
type researchRepo struct {
	s *Store
}

// NewResearchRepository construit le repository de recherche.
func NewResearchRepository(s *Store) *researchRepo { return &researchRepo{s: s} }

func (r *researchRepo) Create(ctx context.Context, req domain.ResearchRequest) (domain.ResearchRequest, error) {
	markets, _ := json.Marshal(req.Markets)
	row, err := r.s.q.CreateResearchRequest(ctx, db.CreateResearchRequestParams{
		UserID:   req.UserID,
		Query:    req.Query,
		Sector:   req.Sector,
		Markets:  markets,
		Language: req.Language,
	})
	if err != nil {
		return domain.ResearchRequest{}, err
	}
	return toResearchRequest(row), nil
}

func (r *researchRepo) Get(ctx context.Context, userID, id string) (domain.ResearchRequest, error) {
	row, err := r.s.q.GetResearchRequest(ctx, db.GetResearchRequestParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.ResearchRequest{}, domain.ErrNotFound
		}
		return domain.ResearchRequest{}, err
	}
	return toResearchRequest(row), nil
}

func (r *researchRepo) UpdateStatus(ctx context.Context, id, status string) (domain.ResearchRequest, error) {
	row, err := r.s.q.UpdateResearchRequestStatus(ctx, db.UpdateResearchRequestStatusParams{ID: id, Status: status})
	if err != nil {
		if isNoRows(err) {
			return domain.ResearchRequest{}, domain.ErrNotFound
		}
		return domain.ResearchRequest{}, err
	}
	return toResearchRequest(row), nil
}

func toResearchRequest(row db.ResearchRequest) domain.ResearchRequest {
	req := domain.ResearchRequest{
		ID:        row.ID,
		UserID:    row.UserID,
		Query:     row.Query,
		Sector:    row.Sector,
		Language:  row.Language,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	_ = json.Unmarshal(row.Markets, &req.Markets)
	if row.Error != nil {
		req.Error = *row.Error
	}
	return req
}
