package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"afrilaunch/backend/internal/infra/postgres/db"
)

// Store encapsule le pool et les requêtes sqlc typées.
type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewStore construit un Store à partir d'un pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: db.New(pool)}
}

// Close ferme le pool sous-jacent.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping vérifie la connectivité (readiness).
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// inTx exécute fn dans une transaction (rollback automatique sur erreur).
func (s *Store) inTx(ctx context.Context, fn func(tx pgx.Tx, q *db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx, db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
