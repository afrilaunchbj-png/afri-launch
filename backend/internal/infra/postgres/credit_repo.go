package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// creditRepo implémente port.CreditRepository (ledger idempotent).
type creditRepo struct {
	s *Store
}

// NewCreditRepository construit le repository de crédits.
func NewCreditRepository(s *Store) *creditRepo { return &creditRepo{s: s} }

func (r *creditRepo) GetAccount(ctx context.Context, userID string) (domain.CreditAccount, error) {
	a, err := r.s.q.GetCreditAccountByUserID(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return domain.CreditAccount{}, domain.ErrNotFound
		}
		return domain.CreditAccount{}, err
	}
	return toCreditAccount(a), nil
}

func (r *creditRepo) GetOrCreateAccount(ctx context.Context, userID string, initialBalance int64) (domain.CreditAccount, error) {
	a, err := r.s.q.GetCreditAccountByUserID(ctx, userID)
	if err == nil {
		return toCreditAccount(a), nil
	}
	if !isNoRows(err) {
		return domain.CreditAccount{}, err
	}

	created, err := r.s.q.CreateCreditAccount(ctx, db.CreateCreditAccountParams{
		UserID:  userID,
		Balance: int32(initialBalance),
	})
	if err != nil {
		if isUniqueViolation(err) {
			a, err = r.s.q.GetCreditAccountByUserID(ctx, userID)
			if err != nil {
				return domain.CreditAccount{}, err
			}
			return toCreditAccount(a), nil
		}
		return domain.CreditAccount{}, err
	}
	return toCreditAccount(created), nil
}

func (r *creditRepo) Grant(ctx context.Context, userID string, amount int64, operation, reference string) (domain.CreditTransaction, error) {
	if existing, err := r.s.q.GetCreditTransactionByReference(ctx, &reference); err == nil {
		return toCreditTransaction(existing), nil
	}

	var result domain.CreditTransaction
	err := r.s.inTx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		acc, err := r.ensureAccountTx(ctx, q, userID)
		if err != nil {
			return err
		}
		if _, err := q.AddCredits(ctx, db.AddCreditsParams{ID: acc.ID, Balance: int32(amount)}); err != nil {
			return err
		}
		t, err := q.CreateCreditTransaction(ctx, db.CreateCreditTransactionParams{
			AccountID: acc.ID,
			Type:      domain.TransactionCredit,
			Amount:    int32(amount),
			Operation: operation,
			Reference: &reference,
			Metadata:  []byte("{}"),
		})
		if err != nil {
			return err
		}
		result = toCreditTransaction(t)
		return nil
	})
	if err != nil && isUniqueViolation(err) {
		existing, ferr := r.s.q.GetCreditTransactionByReference(ctx, &reference)
		if ferr != nil {
			return domain.CreditTransaction{}, ferr
		}
		return toCreditTransaction(existing), nil
	}
	return result, err
}

func (r *creditRepo) Reserve(ctx context.Context, userID string, amount int64, operation, reference string, ttl time.Duration) (domain.CreditReservation, error) {
	if existing, err := r.s.q.GetCreditReservationByReference(ctx, reference); err == nil {
		return toCreditReservation(existing), nil
	}

	var result domain.CreditReservation
	err := r.s.inTx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		acc, err := r.ensureAccountTx(ctx, q, userID)
		if err != nil {
			return err
		}
		if int64(acc.Balance-acc.Reserved) < amount {
			return domain.ErrInsufficient
		}
		if _, err := q.AddReserved(ctx, db.AddReservedParams{ID: acc.ID, Reserved: int32(amount)}); err != nil {
			return err
		}
		res, err := q.CreateCreditReservation(ctx, db.CreateCreditReservationParams{
			AccountID: acc.ID,
			Amount:    int32(amount),
			Operation: operation,
			Reference: reference,
			ExpiresAt: time.Now().Add(ttl),
		})
		if err != nil {
			return err
		}
		result = toCreditReservation(res)
		return nil
	})
	if err != nil && isUniqueViolation(err) {
		existing, ferr := r.s.q.GetCreditReservationByReference(ctx, reference)
		if ferr != nil {
			return domain.CreditReservation{}, ferr
		}
		return toCreditReservation(existing), nil
	}
	return result, err
}

func (r *creditRepo) Consume(ctx context.Context, userID, reference string) (domain.CreditTransaction, error) {
	var result domain.CreditTransaction
	err := r.s.inTx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		acc, err := q.GetCreditAccountForUpdate(ctx, userID)
		if err != nil {
			if isNoRows(err) {
				return domain.ErrNotFound
			}
			return err
		}
		res, err := q.GetCreditReservationByReference(ctx, reference)
		if err != nil {
			if isNoRows(err) {
				return domain.ErrNotFound
			}
			return err
		}
		if res.AccountID != acc.ID {
			return domain.ErrNotFound
		}
		switch res.Status {
		case domain.ReservationConsumed:
			t, err := q.GetCreditTransactionByReference(ctx, &reference)
			if err != nil {
				return err
			}
			result = toCreditTransaction(t)
			return nil
		case domain.ReservationReleased:
			return domain.ErrConflict
		}

		if _, err := q.SubtractReserved(ctx, db.SubtractReservedParams{ID: acc.ID, Reserved: res.Amount}); err != nil {
			return err
		}
		if _, err := q.SubtractCredits(ctx, db.SubtractCreditsParams{ID: acc.ID, Balance: res.Amount}); err != nil {
			return err
		}
		t, err := q.CreateCreditTransaction(ctx, db.CreateCreditTransactionParams{
			AccountID: acc.ID,
			Type:      domain.TransactionDebit,
			Amount:    res.Amount,
			Operation: res.Operation,
			Reference: &reference,
			Metadata:  []byte("{}"),
		})
		if err != nil {
			return err
		}
		if _, err := q.UpdateCreditReservationStatus(ctx, db.UpdateCreditReservationStatusParams{ID: res.ID, Status: domain.ReservationConsumed}); err != nil {
			return err
		}
		result = toCreditTransaction(t)
		return nil
	})
	return result, err
}

func (r *creditRepo) Release(ctx context.Context, userID, reference string) error {
	return r.s.inTx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		acc, err := q.GetCreditAccountForUpdate(ctx, userID)
		if err != nil {
			if isNoRows(err) {
				return domain.ErrNotFound
			}
			return err
		}
		res, err := q.GetCreditReservationByReference(ctx, reference)
		if err != nil {
			if isNoRows(err) {
				return domain.ErrNotFound
			}
			return err
		}
		if res.AccountID != acc.ID {
			return domain.ErrNotFound
		}
		switch res.Status {
		case domain.ReservationReleased:
			return nil
		case domain.ReservationConsumed:
			return domain.ErrConflict
		}
		if _, err := q.SubtractReserved(ctx, db.SubtractReservedParams{ID: acc.ID, Reserved: res.Amount}); err != nil {
			return err
		}
		_, err = q.UpdateCreditReservationStatus(ctx, db.UpdateCreditReservationStatusParams{ID: res.ID, Status: domain.ReservationReleased})
		return err
	})
}

func (r *creditRepo) ListTransactions(ctx context.Context, accountID, typeFilter, operationFilter string, limit, offset int) ([]domain.CreditTransaction, int64, error) {
	total, err := r.s.q.CountCreditTransactions(ctx, db.CountCreditTransactionsParams{
		AccountID:       accountID,
		TypeFilter:      typeFilter,
		OperationFilter: operationFilter,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.q.ListCreditTransactions(ctx, db.ListCreditTransactionsParams{
		AccountID:       accountID,
		TypeFilter:      typeFilter,
		OperationFilter: operationFilter,
		Limit:           int32(limit),
		Offset:          int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.CreditTransaction, 0, len(rows))
	for _, t := range rows {
		items = append(items, toCreditTransaction(t))
	}
	return items, total, nil
}

func (r *creditRepo) Summary(ctx context.Context, accountID string, since time.Time) (domain.CreditSummary, error) {
	acc, err := r.s.q.GetCreditAccountByID(ctx, accountID)
	if err != nil {
		return domain.CreditSummary{}, err
	}
	added, err := r.s.q.SumCreditTransactionsByTypeSince(ctx, db.SumCreditTransactionsByTypeSinceParams{
		AccountID: accountID, Type: domain.TransactionCredit, CreatedAt: since,
	})
	if err != nil {
		return domain.CreditSummary{}, err
	}
	used, err := r.s.q.SumCreditTransactionsByTypeSince(ctx, db.SumCreditTransactionsByTypeSinceParams{
		AccountID: accountID, Type: domain.TransactionDebit, CreatedAt: since,
	})
	if err != nil {
		return domain.CreditSummary{}, err
	}
	return domain.CreditSummary{
		Balance:    int64(acc.Balance),
		Reserved:   int64(acc.Reserved),
		Available:  int64(acc.Balance - acc.Reserved),
		AddedMonth: added,
		UsedMonth:  used,
	}, nil
}

func (r *creditRepo) GetGenerationCost(ctx context.Context, operation string) (domain.GenerationCost, error) {
	c, err := r.s.q.GetGenerationCost(ctx, operation)
	if err != nil {
		if isNoRows(err) {
			return domain.GenerationCost{}, domain.ErrNotFound
		}
		return domain.GenerationCost{}, err
	}
	return toGenerationCost(c), nil
}

func (r *creditRepo) ListGenerationCosts(ctx context.Context) ([]domain.GenerationCost, error) {
	rows, err := r.s.q.ListGenerationCosts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.GenerationCost, 0, len(rows))
	for _, c := range rows {
		out = append(out, toGenerationCost(c))
	}
	return out, nil
}

// ensureAccountTx verrouille (FOR UPDATE) le compte de l'utilisateur, en le
// créant s'il n'existe pas encore.
func (r *creditRepo) ensureAccountTx(ctx context.Context, q *db.Queries, userID string) (db.CreditAccount, error) {
	acc, err := q.GetCreditAccountForUpdate(ctx, userID)
	if err == nil {
		return acc, nil
	}
	if !isNoRows(err) {
		return acc, err
	}
	if _, err := q.CreateCreditAccount(ctx, db.CreateCreditAccountParams{UserID: userID}); err != nil {
		if !isUniqueViolation(err) {
			return acc, err
		}
	}
	return q.GetCreditAccountForUpdate(ctx, userID)
}

func toCreditAccount(a db.CreditAccount) domain.CreditAccount {
	return domain.CreditAccount{
		ID:        a.ID,
		UserID:    a.UserID,
		Balance:   int64(a.Balance),
		Reserved:  int64(a.Reserved),
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func toCreditTransaction(t db.CreditTransaction) domain.CreditTransaction {
	return domain.CreditTransaction{
		ID:        t.ID,
		AccountID: t.AccountID,
		Type:      t.Type,
		Amount:    int64(t.Amount),
		Operation: t.Operation,
		Status:    t.Status,
		Reference: t.Reference,
		CreatedAt: t.CreatedAt,
	}
}

func toCreditReservation(res db.CreditReservation) domain.CreditReservation {
	return domain.CreditReservation{
		ID:        res.ID,
		AccountID: res.AccountID,
		Amount:    int64(res.Amount),
		Operation: res.Operation,
		Reference: res.Reference,
		Status:    res.Status,
		ExpiresAt: res.ExpiresAt,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}
}

func toGenerationCost(c db.GenerationCost) domain.GenerationCost {
	return domain.GenerationCost{
		Operation: c.Operation,
		Name:      c.Name,
		Credits:   int64(c.Credits),
		IsActive:  c.IsActive,
	}
}
