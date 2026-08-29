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
	jobs     *jobs.Worker
	projects port.ProjectRepository
	ideas    port.IdeaRepository
}

// NewService construit le service de projets.
func NewService(jobs *jobs.Worker, projects port.ProjectRepository, ideas port.IdeaRepository) *Service {
	return &Service{jobs: jobs, projects: projects, ideas: ideas}
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

// requireConfirmed vérifie que l'idée du projet est confirmée avant de générer
// les assets. Le parcours impose de confirmer le résultat final d'abord.
func (s *Service) requireConfirmed(ctx context.Context, userID, projectID string) error {
	project, err := s.projects.Get(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if project.IdeaID == nil {
		return nil
	}
	idea, err := s.ideas.Get(ctx, userID, *project.IdeaID)
	if err != nil {
		return err
	}
	if idea.Status != domain.IdeaConfirmed {
		return domain.ErrNotConfirmed
	}
	return nil
}

// GenerateEbook lance la génération de l'ebook (PDF).
func (s *Service) GenerateEbook(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobEbook})
}

// GenerateCover lance la génération de la couverture.
func (s *Service) GenerateCover(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobCover})
}

// GeneratePosters lance la génération des affiches publicitaires (x3).
func (s *Service) GeneratePosters(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobPosters})
}

// GenerateSalesPage lance la génération de la page de vente.
func (s *Service) GenerateSalesPage(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobSalesPage})
}
