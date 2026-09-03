// Package preferences expose la lecture et la mise à jour des préférences
// utilisateur (langue, thème). La langue pilote aussi la langue des réponses
// du copilote.
package preferences

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les préférences utilisateur.
type Service struct {
	repo port.PreferenceRepository
}

// NewService construit le service de préférences.
func NewService(repo port.PreferenceRepository) *Service { return &Service{repo: repo} }

// Get renvoie les préférences (créées avec les défauts si absentes).
func (s *Service) Get(ctx context.Context, userID string) (domain.UserPreference, error) {
	return s.repo.GetOrCreate(ctx, userID)
}

// UpdateInput porte les champs modifiables (nil = inchangé).
type UpdateInput struct {
	Language *string
	Theme    *string
}

// Update valide et persiste les préférences.
func (s *Service) Update(ctx context.Context, userID string, in UpdateInput) (domain.UserPreference, error) {
	current, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return domain.UserPreference{}, err
	}

	next := current
	if in.Language != nil {
		if !domain.Contains(domain.SupportedLanguages(), *in.Language) {
			return domain.UserPreference{}, domain.ErrInvalidInput
		}
		next.Language = *in.Language
	}
	if in.Theme != nil {
		if !domain.Contains(domain.SupportedThemes(), *in.Theme) {
			return domain.UserPreference{}, domain.ErrInvalidInput
		}
		next.Theme = *in.Theme
	}
	return s.repo.Upsert(ctx, next)
}
