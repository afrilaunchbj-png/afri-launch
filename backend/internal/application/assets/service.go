// Package assets expose la lecture et le téléchargement des assets générés.
package assets

import (
	"context"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// Service expose les assets générés.
type Service struct {
	assets  port.AssetRepository
	storage port.Storage
}

// NewService construit le service d'assets.
func NewService(assets port.AssetRepository, storage port.Storage) *Service {
	return &Service{assets: assets, storage: storage}
}

// List renvoie les assets d'un projet.
func (s *Service) List(ctx context.Context, projectID string) ([]domain.Asset, error) {
	return s.assets.ListByProject(ctx, projectID)
}

// Download renvoie l'asset et son contenu (bytes).
func (s *Service) Download(ctx context.Context, userID, assetID string) (domain.Asset, []byte, error) {
	asset, err := s.assets.Get(ctx, userID, assetID)
	if err != nil {
		return domain.Asset{}, nil, err
	}
	data, err := s.storage.Get(ctx, asset.StorageKey)
	if err != nil {
		return domain.Asset{}, nil, err
	}
	return asset, data, nil
}
