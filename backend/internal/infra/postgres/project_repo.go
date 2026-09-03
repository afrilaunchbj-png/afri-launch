package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// projectRepo implémente port.ProjectRepository.
type projectRepo struct {
	s *Store
}

// NewProjectRepository construit le repository de projets.
func NewProjectRepository(s *Store) *projectRepo { return &projectRepo{s: s} }

func (r *projectRepo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	row, err := r.s.q.CreateProject(ctx, db.CreateProjectParams{
		UserID:        p.UserID,
		OpportunityID: strPtrToUUID(p.OpportunityID),
		IdeaID:        strPtrToUUID(p.IdeaID),
		Title:         p.Title,
		Status:        p.Status,
	})
	if err != nil {
		return domain.Project{}, err
	}
	return toProject(row), nil
}

func (r *projectRepo) Get(ctx context.Context, userID, id string) (domain.Project, error) {
	row, err := r.s.q.GetProject(ctx, db.GetProjectParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.Project{}, domain.ErrNotFound
		}
		return domain.Project{}, err
	}
	return toProject(row), nil
}

func (r *projectRepo) ListByUser(ctx context.Context, userID string) ([]domain.Project, error) {
	rows, err := r.s.q.ListProjectsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProject(row))
	}
	return out, nil
}

func (r *projectRepo) UpdateStatus(ctx context.Context, userID, id, status string) (domain.Project, error) {
	row, err := r.s.q.UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{ID: id, Status: status, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.Project{}, domain.ErrNotFound
		}
		return domain.Project{}, err
	}
	return toProject(row), nil
}

func (r *projectRepo) AddCredits(ctx context.Context, id string, amount int64) (domain.Project, error) {
	row, err := r.s.q.AddProjectCredits(ctx, db.AddProjectCreditsParams{ID: id, CreditsConsumed: int32(amount)})
	if err != nil {
		return domain.Project{}, err
	}
	return toProject(row), nil
}

func (r *projectRepo) UpdateConfig(ctx context.Context, userID, id string, config []byte) (domain.Project, error) {
	row, err := r.s.q.UpdateProjectConfig(ctx, db.UpdateProjectConfigParams{ID: id, UserID: userID, Config: config})
	if err != nil {
		if isNoRows(err) {
			return domain.Project{}, domain.ErrNotFound
		}
		return domain.Project{}, err
	}
	return toProject(row), nil
}

func toProject(p db.Project) domain.Project {
	return domain.Project{
		ID:              p.ID,
		UserID:          p.UserID,
		OpportunityID:   uuidPtr(p.OpportunityID),
		IdeaID:          uuidPtr(p.IdeaID),
		Title:           p.Title,
		Status:          p.Status,
		CreditsConsumed: int64(p.CreditsConsumed),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Config:          p.Config,
	}
}
