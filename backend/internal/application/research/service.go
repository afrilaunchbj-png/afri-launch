// Package research orchestre la recherche d'opportunités en ligne.
package research

import (
	"context"

	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les demandes de recherche en ligne.
type Service struct {
	jobs *jobs.Worker
	repo port.ResearchRepository
}

// NewService construit le service de recherche.
func NewService(jobs *jobs.Worker, repo port.ResearchRepository) *Service {
	return &Service{jobs: jobs, repo: repo}
}

// Start crée une demande de recherche et lance le job associé (asynchrone).
func (s *Service) Start(ctx context.Context, userID, query, sector string, markets []string, language string) (domain.GenerationJob, error) {
	if language == "" {
		language = "fr"
	}
	req, err := s.repo.Create(ctx, domain.ResearchRequest{
		UserID:   userID,
		Query:    query,
		Sector:   sector,
		Markets:  markets,
		Language: language,
		Status:   domain.ResearchPending,
	})
	if err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{
		UserID:     userID,
		ResearchID: &req.ID,
		Kind:       domain.JobResearch,
	})
}

// Get renvoie une demande de recherche.
func (s *Service) Get(ctx context.Context, userID, id string) (domain.ResearchRequest, error) {
	return s.repo.Get(ctx, userID, id)
}
