// Package postgres fournit le pool de connexions et les adaptateurs
// (repositories) implémentant les ports de l'application via sqlc.
package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Open crée un pool de connexions PostgreSQL et vérifie la connectivité.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	// Compatibilité PgBouncer (endpoint "pooler" de Neon, mode transaction) :
	// le mode par défaut de pgx prépare des statements nommés
	// (stmtcache_1, stmtcache_2…) qui entrent en collision entre connexions
	// clients multiplexées sur la même connexion serveur →
	// "prepared statement name is already in use" (SQLSTATE 08P01).
	// CacheDescribe exécute via le statement SANS NOM (invisible pour
	// PgBouncer) tout en gardant les OIDs de paramètres en cache — les
	// colonnes jsonb ([]byte) continuent de fonctionner.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
