// Package projects orchestre les projets de produit et les générations associées.
// Workflow cover-first : la couverture (identité visuelle) est générée et
// validée en premier ; ebook, affiches et page de vente en dépendent.
package projects

import (
	"context"
	"encoding/json"

	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
) // Service orchestre les projets.
type Service struct {
	jobs     *jobs.Worker
	projects port.ProjectRepository
	ideas    port.IdeaRepository
	assets   port.AssetRepository
}

// NewService construit le service de projets.
func NewService(jobs *jobs.Worker, projects port.ProjectRepository, ideas port.IdeaRepository, assets port.AssetRepository) *Service {
	return &Service{jobs: jobs, projects: projects, ideas: ideas, assets: assets}
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

// ConfigInput porte les champs modifiables de la configuration (pointeur =
// champ absent de la requête, inchangé).
type ConfigInput struct {
	Palette       *domain.ProjectPalette `json:"palette,omitempty"`       // nil = inchangée ; vide = l'IA re-proposera
	ClearPalette  bool                   `json:"clear_palette,omitempty"` // force la re-proposition par l'IA
	Style         *string                `json:"style,omitempty"`
	EbookMinPages *int                   `json:"ebook_min_pages,omitempty"`
	EbookMaxPages *int                   `json:"ebook_max_pages,omitempty"`
}

// UpdateConfig met à jour l'identité visuelle et les réglages du projet.
func (s *Service) UpdateConfig(ctx context.Context, userID, projectID string, in ConfigInput) (domain.Project, error) {
	current, err := s.projects.Get(ctx, userID, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	config := domain.ParseProjectConfig(current.Config)

	if in.ClearPalette {
		config.Palette = nil
		config.Style = ""
	}
	if in.Palette != nil {
		p := in.Palette.Normalize(domain.PaletteSourceUser)
		for _, value := range []string{p.Primary, p.Secondary, p.Accent, p.Background, p.Text} {
			if !domain.ValidHexColor(value) {
				return domain.Project{}, domain.ErrInvalidInput
			}
		}
		if p.Empty() {
			config.Palette = nil
		} else {
			config.Palette = &p
		}
	}
	if in.Style != nil {
		config.Style = *in.Style
	}
	if in.EbookMinPages != nil {
		config.EbookMinPages = *in.EbookMinPages
	}
	if in.EbookMaxPages != nil {
		config.EbookMaxPages = *in.EbookMaxPages
	}
	// Bornes : toute valeur hors limites est ramenée dans l'intervalle.
	config.EbookMinPages, config.EbookMaxPages = clampPages(config.EbookMinPages), clampPages(config.EbookMaxPages)

	return s.projects.UpdateConfig(ctx, userID, projectID, config.Marshal())
}

func clampPages(n int) int {
	if n <= 0 {
		return 0 // non défini : défaut appliqué à la lecture
	}
	if n < domain.EbookMinPagesFloor {
		return domain.EbookMinPagesFloor
	}
	if n > domain.EbookMaxPagesCeiling {
		return domain.EbookMaxPagesCeiling
	}
	return n
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

// requireCover applique le workflow cover-first : la couverture doit exister
// avant l'ebook (qui l'intègre) et les assets marketing (qui en héritent
// l'identité visuelle).
func (s *Service) requireCover(ctx context.Context, userID, projectID string) error {
	assets, err := s.assets.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}
	for _, a := range assets {
		if a.Kind == domain.AssetCover {
			return nil
		}
	}
	return domain.ErrCoverRequired
}

// GenerateEbook lance la génération de l'ebook (PDF + deck).
func (s *Service) GenerateEbook(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	if err := s.requireCover(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobEbook})
}

// GenerateCover lance la génération (ou régénération) de la couverture.
// Chaque génération consomme des crédits (opération image_generation) ;
// instructions est le feedback utilisateur optionnel pour la régénération.
func (s *Service) GenerateCover(ctx context.Context, userID, projectID, instructions string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	var params []byte
	if instructions != "" {
		params, _ = json.Marshal(map[string]string{"instructions": instructions})
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobCover, Params: params})
}

// GeneratePosters lance la génération des affiches publicitaires (x3).
func (s *Service) GeneratePosters(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	if err := s.requireCover(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobPosters})
}

// GenerateSalesPage lance la génération de la page de vente.
func (s *Service) GenerateSalesPage(ctx context.Context, userID, projectID string) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	if err := s.requireCover(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{UserID: userID, ProjectID: &projectID, Kind: domain.JobSalesPage})
}

// GenerateVideoAd lance la génération d'une vidéo publicitaire (job
// video_ad). Le mockup du produit est la cover du projet (workflow
// cover-first) ; les paramètres sont normalisés côté domaine.
func (s *Service) GenerateVideoAd(ctx context.Context, userID, projectID string, params domain.VideoAdParams) (domain.GenerationJob, error) {
	if err := s.requireConfirmed(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	if err := s.requireCover(ctx, userID, projectID); err != nil {
		return domain.GenerationJob{}, err
	}
	return s.jobs.Dispatch(ctx, jobs.DispatchParams{
		UserID:    userID,
		ProjectID: &projectID,
		Kind:      domain.JobVideoAd,
		Params:    params.Normalized().Marshal(),
	})
}
