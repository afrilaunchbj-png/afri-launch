// Package auth orchestre l'identité utilisateur à partir de Neon Auth.
// La création de compte (email/mot de passe) et les sessions sont gérées par
// Neon ; ce service maintient le profil local (`users`) et accorde le bonus
// de bienvenue au premier login.
package auth

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service réconcilie l'identité Neon Auth avec la table `users`.
type Service struct {
	users   port.UserRepository
	credits port.CreditRepository
}

// NewService construit le service d'identité.
func NewService(users port.UserRepository, credits port.CreditRepository) *Service {
	return &Service{users: users, credits: credits}
}

// GetOrCreateUser crée ou met à jour le profil local à partir de l'identité
// vérifiée, et accorde le bonus de bienvenue au premier login.
func (s *Service) GetOrCreateUser(ctx context.Context, identity port.AuthUser, welcomeCredits int64) (domain.User, error) {
	user := domain.User{
		ID:        identity.ID,
		Email:     identity.Email,
		FullName:  identity.Name,
		AvatarURL: stringPtr(identity.Picture),
	}

	if _, err := s.users.GetByID(ctx, identity.ID); err == domain.ErrNotFound {
		// Premier login : création du profil + bonus de bienvenue.
		created, err := s.users.Upsert(ctx, user)
		if err != nil {
			return domain.User{}, err
		}
		if welcomeCredits > 0 {
			if _, err := s.credits.GetOrCreateAccount(ctx, user.ID, 0); err != nil {
				return domain.User{}, err
			}
			if _, err := s.credits.Grant(ctx, user.ID, welcomeCredits, domain.OperationWelcomeBonus, "welcome:"+user.ID); err != nil {
				return domain.User{}, err
			}
		}
		return created, nil
	}

	// Login suivant : on synchronise le profil (nom/email/avatar).
	return s.users.Upsert(ctx, user)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
