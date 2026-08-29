// Package ideas expose la génération et la lecture des idées de produit.
package ideas

import (
	"context"

	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre la génération d'idées.
type Service struct {
	jobs          *jobs.Worker
	ideas         port.IdeaRepository
	opportunities port.OpportunityRepository
}

// NewService construit le service d'idées.
func NewService(jobs *jobs.Worker, ideas port.IdeaRepository, opportunities port.OpportunityRepository) *Service {
	return &Service{jobs: jobs, ideas: ideas, opportunities: opportunities}
}

// Generate lance la génération d'idées pour une opportunité (asynchrone).
func (s *Service) Generate(ctx context.Context, userID, opportunityID string) (domain.GenerationJob, error) {
	if _, err := s.opportunities.Get(ctx, opportunityID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, userID, nil, &opportunityID, domain.JobIdeas)
}

// ListByOpportunity renvoie les idées d'une opportunité.
func (s *Service) ListByOpportunity(ctx context.Context, userID, opportunityID string) ([]domain.ProductIdea, error) {
	return s.ideas.ListByOpportunity(ctx, userID, opportunityID)
}

// List renvoie toutes les idées de l'utilisateur.
func (s *Service) List(ctx context.Context, userID string) ([]domain.ProductIdea, error) {
	return s.ideas.ListByUser(ctx, userID)
}

// Get renvoie une idée.
func (s *Service) Get(ctx context.Context, userID, id string) (domain.ProductIdea, error) {
	return s.ideas.Get(ctx, userID, id)
}
