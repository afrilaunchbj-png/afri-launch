// Package ideas expose la génération, la lecture et l'itération des idées.
package ideas

import (
	"context"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre la génération et l'itération des idées.
type Service struct {
	jobs          *jobs.Worker
	ideas         port.IdeaRepository
	ideaMessages  port.IdeaMessageRepository
	opportunities port.OpportunityRepository
	credits       port.CreditRepository
	ai            *ai.Service
}

// NewService construit le service d'idées.
func NewService(jobs *jobs.Worker, ideas port.IdeaRepository, ideaMessages port.IdeaMessageRepository, opportunities port.OpportunityRepository, credits port.CreditRepository, ai *ai.Service) *Service {
	return &Service{jobs: jobs, ideas: ideas, ideaMessages: ideaMessages, opportunities: opportunities, credits: credits, ai: ai}
}

// Generate lance la génération d'idées pour une opportunité (asynchrone).
func (s *Service) Generate(ctx context.Context, userID, opportunityID string) (domain.GenerationJob, error) {
	if _, err := s.opportunities.Get(ctx, opportunityID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, OpportunityID: &opportunityID, Kind: domain.JobIdeas})
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

// ListMessages renvoie l'historique de conversation d'une idée.
func (s *Service) ListMessages(ctx context.Context, userID, ideaID string) ([]domain.IdeaMessage, error) {
	if _, err := s.ideas.Get(ctx, userID, ideaID); err != nil {
		return nil, err
	}
	return s.ideaMessages.ListByIdea(ctx, ideaID)
}

// Confirm valide le résultat final d'une idée (draft → confirmed).
func (s *Service) Confirm(ctx context.Context, userID, ideaID string) (domain.ProductIdea, error) {
	if _, err := s.ideas.Get(ctx, userID, ideaID); err != nil {
		return domain.ProductIdea{}, err
	}
	return s.ideas.SetStatus(ctx, userID, ideaID, domain.IdeaConfirmed)
}
