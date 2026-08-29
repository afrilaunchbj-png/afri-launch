package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// jobRepo implémente port.JobRepository.
type jobRepo struct {
	s *Store
}

// NewJobRepository construit le repository de jobs.
func NewJobRepository(s *Store) *jobRepo { return &jobRepo{s: s} }

func (r *jobRepo) Create(ctx context.Context, userID string, projectID, opportunityID, researchID, ideaID *string, kind string, cost int64) (domain.GenerationJob, error) {
	row, err := r.s.q.CreateJob(ctx, db.CreateJobParams{
		UserID:        userID,
		ProjectID:     strPtrToUUID(projectID),
		OpportunityID: strPtrToUUID(opportunityID),
		ResearchID:    strPtrToUUID(researchID),
		IdeaID:        strPtrToUUID(ideaID),
		Kind:          kind,
		Cost:          int32(cost),
	})
	if err != nil {
		return domain.GenerationJob{}, err
	}
	return toJob(row), nil
}

func (r *jobRepo) Get(ctx context.Context, userID, id string) (domain.GenerationJob, error) {
	row, err := r.s.q.GetJob(ctx, db.GetJobParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.GenerationJob{}, domain.ErrNotFound
		}
		return domain.GenerationJob{}, err
	}
	return toJob(row), nil
}

func (r *jobRepo) UpdateStatus(ctx context.Context, id, status string) (domain.GenerationJob, error) {
	row, err := r.s.q.UpdateJobStatus(ctx, db.UpdateJobStatusParams{ID: id, Status: status})
	if err != nil {
		return domain.GenerationJob{}, err
	}
	return toJob(row), nil
}

func (r *jobRepo) Complete(ctx context.Context, id string, result []byte, cost int64) (domain.GenerationJob, error) {
	row, err := r.s.q.CompleteJob(ctx, db.CompleteJobParams{ID: id, Result: result, Cost: int32(cost)})
	if err != nil {
		return domain.GenerationJob{}, err
	}
	return toJob(row), nil
}

func (r *jobRepo) Fail(ctx context.Context, id, errMsg string) (domain.GenerationJob, error) {
	row, err := r.s.q.FailJob(ctx, db.FailJobParams{ID: id, Error: &errMsg})
	if err != nil {
		return domain.GenerationJob{}, err
	}
	return toJob(row), nil
}

func (r *jobRepo) ListByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error) {
	rows, err := r.s.q.ListJobsByProject(ctx, strPtrToUUID(&projectID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.GenerationJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, toJob(row))
	}
	return out, nil
}

func toJob(j db.GenerationJob) domain.GenerationJob {
	job := domain.GenerationJob{
		ID:            j.ID,
		UserID:        j.UserID,
		ProjectID:     uuidPtr(j.ProjectID),
		OpportunityID: uuidPtr(j.OpportunityID),
		ResearchID:    uuidPtr(j.ResearchID),
		IdeaID:        uuidPtr(j.IdeaID),
		Kind:          j.Kind,
		Status:        j.Status,
		Cost:          int64(j.Cost),
		Result:        j.Result,
		CreatedAt:     j.CreatedAt,
		UpdatedAt:     j.UpdatedAt,
	}
	if j.Error != nil {
		job.Error = *j.Error
	}
	job.CompletedAt = timestamptzPtr(j.CompletedAt)
	return job
}
