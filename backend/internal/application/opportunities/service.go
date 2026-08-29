// Package opportunities implémente les cas d'usage de la recherche
// d'opportunités.
package opportunities

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre la recherche et la sauvegarde d'opportunités.
type Service struct {
	repo port.OpportunityRepository
}

// NewService construit le service d'opportunités.
func NewService(repo port.OpportunityRepository) *Service {
	return &Service{repo: repo}
}

// List renvoie les opportunités filtrées (paginées) avec l'état « sauvegardée »
// pour l'utilisateur courant.
func (s *Service) List(ctx context.Context, userID string, f port.OpportunityFilter, limit, offset int) ([]domain.Opportunity, int64, error) {
	items, total, err := s.repo.List(ctx, f, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.markSaved(ctx, userID, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Get renvoie une opportunité avec son état de sauvegarde.
func (s *Service) Get(ctx context.Context, userID, id string) (domain.Opportunity, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Opportunity{}, err
	}
	saved, err := s.repo.ListSavedIDs(ctx, userID)
	if err != nil {
		return domain.Opportunity{}, err
	}
	for _, sid := range saved {
		if sid == item.ID {
			item.IsSaved = true
			break
		}
	}
	return item, nil
}

// Save met une opportunité de côté pour l'utilisateur.
func (s *Service) Save(ctx context.Context, userID, id string) error {
	if _, err := s.repo.Get(ctx, id); err != nil {
		return err
	}
	return s.repo.Save(ctx, userID, id)
}

// Unsave retire une opportunité des sauvegardes.
func (s *Service) Unsave(ctx context.Context, userID, id string) error {
	return s.repo.Unsave(ctx, userID, id)
}

// Countries renvoie les pays disponibles (facettes de filtre).
func (s *Service) Countries(ctx context.Context) ([]string, error) {
	return s.repo.Countries(ctx)
}

// Sectors renvoie les secteurs disponibles (facettes de filtre).
func (s *Service) Sectors(ctx context.Context) ([]string, error) {
	return s.repo.Sectors(ctx)
}

func (s *Service) markSaved(ctx context.Context, userID string, items []domain.Opportunity) error {
	if userID == "" {
		return nil
	}
	saved, err := s.repo.ListSavedIDs(ctx, userID)
	if err != nil {
		return err
	}
	set := make(map[string]struct{}, len(saved))
	for _, id := range saved {
		set[id] = struct{}{}
	}
	for i := range items {
		if _, ok := set[items[i].ID]; ok {
			items[i].IsSaved = true
		}
	}
	return nil
}
