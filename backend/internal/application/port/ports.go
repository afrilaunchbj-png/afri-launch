// Package port définit les interfaces (ports) entre la couche application
// et l'infrastructure. Les adaptateurs infra les implémentent.
package port

import (
	"context"
	"time"

	"afrilaunch/backend/internal/domain"
)

// AuthUser est l'identité extraite d'un token vérifié (Neon Auth).
type AuthUser struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

// UserRepository accède aux comptes (profils) utilisateurs.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	Upsert(ctx context.Context, u domain.User) (domain.User, error)
	SetRole(ctx context.Context, userID, role string) (domain.User, error)
	List(ctx context.Context, limit, offset int) ([]domain.User, int64, error)
}

// CreditRepository implémente le ledger de crédits idempotent.
// `reference` est la clé d'idempotence (unique par opération).
type CreditRepository interface {
	GetOrCreateAccount(ctx context.Context, userID string, initialBalance int64) (domain.CreditAccount, error)
	GetAccount(ctx context.Context, userID string) (domain.CreditAccount, error)

	Grant(ctx context.Context, userID string, amount int64, operation, reference string) (domain.CreditTransaction, error)
	Reserve(ctx context.Context, userID string, amount int64, operation, reference string, ttl time.Duration) (domain.CreditReservation, error)
	Consume(ctx context.Context, userID, reference string) (domain.CreditTransaction, error)
	Release(ctx context.Context, userID, reference string) error

	ListTransactions(ctx context.Context, accountID string, typeFilter, operationFilter string, limit, offset int) ([]domain.CreditTransaction, int64, error)
	Summary(ctx context.Context, accountID string, since time.Time) (domain.CreditSummary, error)

	GetGenerationCost(ctx context.Context, operation string) (domain.GenerationCost, error)
	ListGenerationCosts(ctx context.Context) ([]domain.GenerationCost, error)
}

// OpportunityFilter filtre la recherche d'opportunités.
type OpportunityFilter struct {
	UserID     string // scope : catalogue global + opportunités de l'utilisateur
	Country    string
	Sector     string
	Difficulty string
	Query      string
}

// OpportunityRepository accède au catalogue d'opportunités et aux sauvegardes.
type OpportunityRepository interface {
	List(ctx context.Context, f OpportunityFilter, limit, offset int) ([]domain.Opportunity, int64, error)
	Get(ctx context.Context, id string) (domain.Opportunity, error)
	Create(ctx context.Context, o domain.Opportunity) (domain.Opportunity, error)
	ListSavedIDs(ctx context.Context, userID string) ([]string, error)
	Save(ctx context.Context, userID, opportunityID string) error
	Unsave(ctx context.Context, userID, opportunityID string) error
	Countries(ctx context.Context) ([]string, error)
	Sectors(ctx context.Context) ([]string, error)
}

// MarketRepository accède au référentiel des marchés.
type MarketRepository interface {
	List(ctx context.Context) ([]domain.Market, error)
}

// TokenVerifier vérifie un token JWT (Neon Auth, EdDSA/JWKS) et renvoie
// l'identité utilisateur associée.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (AuthUser, error)
}
