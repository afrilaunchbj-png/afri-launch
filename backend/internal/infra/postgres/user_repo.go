package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// userRepo implémente port.UserRepository.
type userRepo struct {
	s *Store
}

// NewUserRepository construit le repository utilisateur.
func NewUserRepository(s *Store) *userRepo { return &userRepo{s: s} }

func (r *userRepo) GetByID(ctx context.Context, id string) (domain.User, error) {
	u, err := r.s.q.GetUserByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toUser(u), nil
}

func (r *userRepo) Upsert(ctx context.Context, user domain.User) (domain.User, error) {
	u, err := r.s.q.UpsertUser(ctx, db.UpsertUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarUrl: user.AvatarURL,
	})
	if err != nil {
		return domain.User{}, err
	}
	return toUser(u), nil
}

func toUser(u db.User) domain.User {
	return domain.User{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarUrl,
		CreatedAt: u.CreatedAt,
	}
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
