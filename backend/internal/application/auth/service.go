// Package auth orchestre l'identité utilisateur à partir de Neon Auth.
// La création de compte (email/mot de passe) et les sessions sont gérées par
// Neon ; ce service maintient le profil local (`users`) et accorde le bonus
// de bienvenue au premier login.
package auth

import (
	"context"
	"strings"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service réconcilie l'identité Neon Auth avec la table `users`.
type Service struct {
	users       port.UserRepository
	credits     port.CreditRepository
	superadmins []string // emails promus superadmin au login
}

// NewService construit le service d'identité.
func NewService(users port.UserRepository, credits port.CreditRepository, superadminEmails []string) *Service {
	return &Service{users: users, credits: credits, superadmins: superadminEmails}
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
		// Promotion superadmin déclarative (SUPERADMIN_EMAILS).
		if s.isSuperadmin(created.Email) {
			return s.users.SetRole(ctx, created.ID, domain.RoleSuperadmin)
		}
		return created, nil
	}

	// Login suivant : on synchronise le profil (nom/email/avatar).
	user, err := s.users.Upsert(ctx, user)
	if err != nil {
		return domain.User{}, err
	}
	// Promotion superadmin déclarative (SUPERADMIN_EMAILS).
	if user.Role != domain.RoleSuperadmin && s.isSuperadmin(user.Email) {
		return s.users.SetRole(ctx, user.ID, domain.RoleSuperadmin)
	}
	return user, nil
}

func (s *Service) isSuperadmin(email string) bool {
	for _, e := range s.superadmins {
		if e == strings.ToLower(email) {
			return true
		}
	}
	return false
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
