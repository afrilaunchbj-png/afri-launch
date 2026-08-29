// Package credits implémente les cas d'usage du ledger de crédits.
package credits

import (
	"context"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service orchestre les opérations sur le ledger de crédits.
type Service struct {
	repo port.CreditRepository
}

// NewService construit le service de crédits.
func NewService(repo port.CreditRepository) *Service {
	return &Service{repo: repo}
}

// Summary renvoie le solde et les agrégats mensuels d'un utilisateur.
func (s *Service) Summary(ctx context.Context, userID string) (domain.CreditSummary, error) {
	account, err := s.repo.GetOrCreateAccount(ctx, userID, 0)
	if err != nil {
		return domain.CreditSummary{}, err
	}
	return s.repo.Summary(ctx, account.ID, startOfMonth())
}

// Transactions liste le journal comptable (paginé).
func (s *Service) Transactions(ctx context.Context, userID, typeFilter, operationFilter string, limit, offset int) ([]domain.CreditTransaction, int64, error) {
	account, err := s.repo.GetOrCreateAccount(ctx, userID, 0)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.ListTransactions(ctx, account.ID, typeFilter, operationFilter, limit, offset)
}

// Reserve bloque des crédits pour une génération (idempotent par reference).
func (s *Service) Reserve(ctx context.Context, userID string, amount int64, operation, reference string) (domain.CreditReservation, error) {
	if amount <= 0 {
		return domain.CreditReservation{}, domain.ErrInvalidInput
	}
	return s.repo.Reserve(ctx, userID, amount, operation, reference, 24*time.Hour)
}

// Consume transforme une réservation en consommation effective (idempotent).
func (s *Service) Consume(ctx context.Context, userID, reference string) (domain.CreditTransaction, error) {
	return s.repo.Consume(ctx, userID, reference)
}

// Release libère une réservation (idempotent).
func (s *Service) Release(ctx context.Context, userID, reference string) error {
	return s.repo.Release(ctx, userID, reference)
}

// GenerationCosts renvoie les coûts configurés des opérations.
func (s *Service) GenerationCosts(ctx context.Context) ([]domain.GenerationCost, error) {
	return s.repo.ListGenerationCosts(ctx)
}

func startOfMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}
