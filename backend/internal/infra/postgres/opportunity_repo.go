package postgres

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// opportunityRepo implémente port.OpportunityRepository.
type opportunityRepo struct {
	s *Store
}

// NewOpportunityRepository construit le repository d'opportunités.
func NewOpportunityRepository(s *Store) *opportunityRepo { return &opportunityRepo{s: s} }

func (r *opportunityRepo) List(ctx context.Context, f port.OpportunityFilter, limit, offset int) ([]domain.Opportunity, int64, error) {
	total, err := r.s.q.CountOpportunities(ctx, db.CountOpportunitiesParams{
		UserID:     strPtrToUUID(&f.UserID),
		Country:    f.Country,
		Sector:     f.Sector,
		Difficulty: f.Difficulty,
		Query:      f.Query,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.q.ListOpportunities(ctx, db.ListOpportunitiesParams{
		UserID:     strPtrToUUID(&f.UserID),
		Country:    f.Country,
		Sector:     f.Sector,
		Difficulty: f.Difficulty,
		Query:      f.Query,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.Opportunity, 0, len(rows))
	for _, o := range rows {
		items = append(items, toOpportunity(o))
	}
	return items, total, nil
}

func (r *opportunityRepo) Get(ctx context.Context, id string) (domain.Opportunity, error) {
	o, err := r.s.q.GetOpportunity(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.Opportunity{}, domain.ErrNotFound
		}
		return domain.Opportunity{}, err
	}
	return toOpportunity(o), nil
}

func (r *opportunityRepo) Create(ctx context.Context, o domain.Opportunity) (domain.Opportunity, error) {
	scores, _ := json.Marshal(o.Scores)
	evidence, _ := json.Marshal(o.Evidence)
	row, err := r.s.q.CreateOpportunity(ctx, db.CreateOpportunityParams{
		UserID:     strPtrToUUID(o.UserID),
		ResearchID: strPtrToUUID(o.ResearchID),
		Title:      o.Title,
		Summary:    o.Summary,
		Country:    o.Country,
		Sector:     o.Sector,
		Language:   o.Language,
		Difficulty: o.Difficulty,
		Signal:     o.Signal,
		Score:      int32(o.Score),
		Scores:     scores,
		Evidence:   evidence,
	})
	if err != nil {
		return domain.Opportunity{}, err
	}
	return toOpportunity(row), nil
}

func (r *opportunityRepo) ListSavedIDs(ctx context.Context, userID string) ([]string, error) {
	ids, err := r.s.q.ListSavedOpportunityIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *opportunityRepo) Save(ctx context.Context, userID, opportunityID string) error {
	return r.s.q.SaveOpportunity(ctx, db.SaveOpportunityParams{UserID: userID, OpportunityID: opportunityID})
}

func (r *opportunityRepo) Unsave(ctx context.Context, userID, opportunityID string) error {
	return r.s.q.UnsaveOpportunity(ctx, db.UnsaveOpportunityParams{UserID: userID, OpportunityID: opportunityID})
}

func (r *opportunityRepo) Countries(ctx context.Context) ([]string, error) {
	return r.s.q.ListDistinctCountries(ctx)
}

func (r *opportunityRepo) Sectors(ctx context.Context) ([]string, error) {
	return r.s.q.ListDistinctSectors(ctx)
}

func toOpportunity(o db.Opportunity) domain.Opportunity {
	opp := domain.Opportunity{
		ID:         o.ID,
		UserID:     uuidPtr(o.UserID),
		ResearchID: uuidPtr(o.ResearchID),
		Title:      o.Title,
		Summary:    o.Summary,
		Country:    o.Country,
		Sector:     o.Sector,
		Language:   o.Language,
		Difficulty: o.Difficulty,
		Signal:     o.Signal,
		Score:      int(o.Score),
		CreatedAt:  o.CreatedAt,
	}
	_ = json.Unmarshal(o.Scores, &opp.Scores)
	_ = json.Unmarshal(o.Evidence, &opp.Evidence)
	return opp
}
