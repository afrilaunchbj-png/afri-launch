package postgres

import (
	"context"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// assetRepo implémente port.AssetRepository.
type assetRepo struct {
	s *Store
}

// NewAssetRepository construit le repository d'assets.
func NewAssetRepository(s *Store) *assetRepo { return &assetRepo{s: s} }

func (r *assetRepo) Create(ctx context.Context, a domain.Asset) (domain.Asset, error) {
	row, err := r.s.q.CreateAsset(ctx, db.CreateAssetParams{
		UserID:      a.UserID,
		ProjectID:   a.ProjectID,
		Kind:        a.Kind,
		StorageKey:  a.StorageKey,
		Filename:    a.Filename,
		ContentType: a.ContentType,
		SizeBytes:   int32(a.SizeBytes),
	})
	if err != nil {
		return domain.Asset{}, err
	}
	return toAsset(row), nil
}

func (r *assetRepo) Get(ctx context.Context, userID, id string) (domain.Asset, error) {
	row, err := r.s.q.GetAsset(ctx, db.GetAssetParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.Asset{}, domain.ErrNotFound
		}
		return domain.Asset{}, err
	}
	return toAsset(row), nil
}

func (r *assetRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Asset, error) {
	rows, err := r.s.q.ListAssetsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Asset, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAsset(row))
	}
	return out, nil
}

func toAsset(a db.Asset) domain.Asset {
	return domain.Asset{
		ID:          a.ID,
		UserID:      a.UserID,
		ProjectID:   a.ProjectID,
		Kind:        a.Kind,
		StorageKey:  a.StorageKey,
		Filename:    a.Filename,
		ContentType: a.ContentType,
		SizeBytes:   int64(a.SizeBytes),
		CreatedAt:   a.CreatedAt,
	}
}
