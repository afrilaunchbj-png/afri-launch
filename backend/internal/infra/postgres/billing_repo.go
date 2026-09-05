package postgres

import (
	"context"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// textPtr convertit une chaîne en pointeur (colonnes TEXT nullables, sqlc
// génère *string avec emit_pointers_for_null_types).
func textPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func textValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// planRepo implémente port.PlanRepository.
type planRepo struct {
	s *Store
}

// NewPlanRepository construit le repository de packs de crédits.
func NewPlanRepository(s *Store) *planRepo { return &planRepo{s: s} }

func (r *planRepo) ListActive(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.s.q.ListActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Plan, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPlan(row))
	}
	return out, nil
}

func (r *planRepo) Get(ctx context.Context, id string) (domain.Plan, error) {
	row, err := r.s.q.GetPlan(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return domain.Plan{}, domain.ErrNotFound
		}
		return domain.Plan{}, err
	}
	return toPlan(row), nil
}

func toPlan(row db.Plan) domain.Plan {
	return domain.Plan{
		ID: row.ID, Name: row.Name, Credits: int64(row.Credits),
		PriceMinor: int64(row.PriceMinor), Currency: row.Currency,
		SortOrder: int(row.SortOrder), IsActive: row.IsActive, CreatedAt: row.CreatedAt,
	}
}

// paymentRepo implémente port.PaymentRepository.
type paymentRepo struct {
	s *Store
}

// NewPaymentRepository construit le repository de paiements.
func NewPaymentRepository(s *Store) *paymentRepo { return &paymentRepo{s: s} }

func (r *paymentRepo) Create(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	row, err := r.s.q.CreatePayment(ctx, db.CreatePaymentParams{
		UserID:         p.UserID,
		PlanID:         toUUIDPtr(&p.PlanID),
		AmountMinor:    int32(p.AmountMinor),
		Currency:       p.Currency,
		Provider:       p.Provider,
		Status:         p.Status,
		IdempotencyKey: p.IdempotencyKey,
	})
	if err != nil {
		return domain.Payment{}, err
	}
	return toPayment(row), nil
}

func (r *paymentRepo) UpdateCheckout(ctx context.Context, id, providerReference, checkoutURL string) error {
	return r.s.q.UpdatePaymentCheckout(ctx, db.UpdatePaymentCheckoutParams{
		ID: id, ProviderReference: textPtr(providerReference), CheckoutUrl: textPtr(checkoutURL),
	})
}

func (r *paymentRepo) MarkStatus(ctx context.Context, id, status string) (domain.Payment, error) {
	row, err := r.s.q.MarkPaymentStatus(ctx, db.MarkPaymentStatusParams{ID: id, Status: status})
	if err != nil {
		if isNoRows(err) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	return toPayment(row), nil
}

func (r *paymentRepo) Get(ctx context.Context, userID, id string) (domain.Payment, error) {
	row, err := r.s.q.GetPayment(ctx, db.GetPaymentParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	return toPayment(row), nil
}

func (r *paymentRepo) GetByProviderReference(ctx context.Context, providerReference string) (domain.Payment, error) {
	row, err := r.s.q.GetPaymentByProviderReference(ctx, textPtr(providerReference))
	if err != nil {
		if isNoRows(err) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	return toPayment(row), nil
}

func (r *paymentRepo) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.s.q.ListPaymentsByUser(ctx, db.ListPaymentsByUserParams{UserID: userID, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Payment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPayment(row))
	}
	return out, nil
}

func toPayment(row db.Payment) domain.Payment {
	var planID string
	if row.PlanID.Valid {
		planID = uuid.UUID(row.PlanID.Bytes).String()
	}
	reference := textValue(row.ProviderReference)
	checkoutURL := textValue(row.CheckoutUrl)
	return domain.Payment{
		ID: row.ID, UserID: row.UserID, PlanID: planID,
		AmountMinor: int64(row.AmountMinor), Currency: row.Currency,
		Provider: row.Provider, Status: row.Status,
		IdempotencyKey: row.IdempotencyKey, ProviderReference: reference,
		CheckoutURL: checkoutURL, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
