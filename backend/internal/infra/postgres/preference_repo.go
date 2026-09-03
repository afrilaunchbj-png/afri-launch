package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// preferenceRepo implémente port.PreferenceRepository.
type preferenceRepo struct {
	s *Store
}

// NewPreferenceRepository construit le repository de préférences.
func NewPreferenceRepository(s *Store) *preferenceRepo { return &preferenceRepo{s: s} }

func (r *preferenceRepo) GetOrCreate(ctx context.Context, userID string) (domain.UserPreference, error) {
	row, err := r.s.q.GetUserPreferences(ctx, userID)
	if isNoRows(err) {
		// Première connexion : valeurs par défaut (langue fr, thème system).
		row, err = r.s.q.InsertUserPreferences(ctx, db.InsertUserPreferencesParams{
			UserID:   userID,
			Language: domain.LanguageFr,
			Theme:    domain.ThemeSystem,
		})
	}
	if err != nil {
		return domain.UserPreference{}, err
	}
	return toPreference(row), nil
}

func (r *preferenceRepo) Upsert(ctx context.Context, p domain.UserPreference) (domain.UserPreference, error) {
	// Assure l'existence de la ligne puis met à jour les champs.
	if _, err := r.GetOrCreate(ctx, p.UserID); err != nil {
		return domain.UserPreference{}, err
	}
	row, err := r.s.q.UpdateUserPreferences(ctx, db.UpdateUserPreferencesParams{
		UserID:   p.UserID,
		Language: p.Language,
		Theme:    p.Theme,
	})
	if err != nil {
		return domain.UserPreference{}, err
	}
	return toPreference(row), nil
}

func toPreference(row db.UserPreference) domain.UserPreference {
	return domain.UserPreference{
		UserID:    row.UserID,
		Language:  row.Language,
		Theme:     row.Theme,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
