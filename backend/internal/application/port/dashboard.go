package port

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// DashboardRepository agrège les indicateurs personnels du tableau de bord.
type DashboardRepository interface {
	Stats(ctx context.Context, userID string) (domain.DashboardStats, error)
}
