package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// paymentCountryRepo implémente port.PaymentCountryRepository.
type paymentCountryRepo struct {
	s *Store
}

// NewPaymentCountryRepository construit le repository des pays de paiement.
func NewPaymentCountryRepository(s *Store) *paymentCountryRepo {
	return &paymentCountryRepo{s: s}
}

func (r *paymentCountryRepo) List(ctx context.Context) ([]domain.PaymentCountry, error) {
	rows, err := r.s.q.ListPaymentCountries(ctx)
	if err != nil {
		return nil, err
	}
	return rowsToPaymentCountries(rows), nil
}

func (r *paymentCountryRepo) ListEnabled(ctx context.Context) ([]domain.PaymentCountry, error) {
	rows, err := r.s.q.ListEnabledPaymentCountries(ctx)
	if err != nil {
		return nil, err
	}
	return rowsToPaymentCountries(rows), nil
}

func (r *paymentCountryRepo) SetEnabled(ctx context.Context, code string, enabled bool) error {
	return r.s.q.SetPaymentCountryEnabled(ctx, db.SetPaymentCountryEnabledParams{Code: code, Enabled: enabled})
}

func rowsToPaymentCountries(rows []db.PaymentCountry) []domain.PaymentCountry {
	out := make([]domain.PaymentCountry, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.PaymentCountry{Code: r.Code, Name: r.Name, Currency: r.Currency, Enabled: r.Enabled})
	}
	return out
}
