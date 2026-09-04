// Package dashboard agrège les indicateurs personnels du tableau de bord.
package dashboard

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service calcule les statistiques d'un utilisateur.
type Service struct {
	repo port.DashboardRepository
}

// NewService construit le service de tableau de bord.
func NewService(repo port.DashboardRepository) *Service { return &Service{repo: repo} }

// Stats renvoie les indicateurs et séries temporelles de l'utilisateur.
func (s *Service) Stats(ctx context.Context, userID string) (domain.DashboardStats, error) {
	return s.repo.Stats(ctx, userID)
}
