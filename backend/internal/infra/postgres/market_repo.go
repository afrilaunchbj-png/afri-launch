package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
)

// marketRepo implémente port.MarketRepository.
type marketRepo struct {
	s *Store
}

// NewMarketRepository construit le repository des marchés.
func NewMarketRepository(s *Store) *marketRepo { return &marketRepo{s: s} }

func (r *marketRepo) List(ctx context.Context) ([]domain.Market, error) {
	rows, err := r.s.q.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Market, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Market{
			ID:       m.ID,
			Code:     m.Code,
			Name:     m.Name,
			Currency: m.Currency,
			Language: m.Language,
		})
	}
	return out, nil
}
