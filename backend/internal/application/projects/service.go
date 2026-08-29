// Package projects orchestre les projets de produit et les générations associées.
package projects

import (
	"context"

	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les projets.
type Service struct {
	jobs    *jobs.Worker
	projects port.ProjectRepository
}

// NewService construit le service de projets.
func NewService(jobs *jobs.Worker, projects port.ProjectRepository) *Service {
	return &Service{jobs: jobs, projects: projects}
}

// Create crée un projet à partir d'une idée (et d'une opportunité).
func (s *Service) Create(ctx context.Context, userID string, opportunityID, ideaID *string, title string) (domain.Project, error) {
	if title == "" {
		title = "Projet"
	}
	return s.projects.Create(ctx, domain.Project{
		UserID:        userID,
		OpportunityID: opportunityID,
		IdeaID:        ideaID,
		Title:         title,
		Status:        domain.ProjectIdeaSelected,
	})
}

// Get renvoie un projet.
func (s *Service) Get(ctx context.Context, userID, id string) (domain.Project, error) {
	return s.projects.Get(ctx, userID, id)
}

// List renvoie les projets de l'utilisateur.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Project, error) {
	return s.projects.ListByUser(ctx, userID)
}

// GenerateEbook lance la génération de l'ebook (PDF).
func (s *Service) GenerateEbook(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if _, err := s.projects.Get(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, userID, &projectID, nil, domain.JobEbook)
}

// GenerateCover lance la génération de la couverture.
func (s *Service) GenerateCover(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if _, err := s.projects.Get(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, userID, &projectID, nil, domain.JobCover)
}

// GenerateSalesPage lance la génération de la page de vente.
func (s *Service) GenerateSalesPage(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if _, err := s.projects.Get(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, userID, &projectID, nil, domain.JobSalesPage)
}
