package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// commerceRepo implémente port.CommerceRepository. Les clés/secrets Chariow
// sont chiffrés AVANT d'arriver ici (couche application) — jamais en clair.
type commerceRepo struct {
	s *Store
}

// NewCommerceRepository construit le repository du module commerce.
func NewCommerceRepository(s *Store) *commerceRepo { return &commerceRepo{s: s} }

func commerceConnFromRow(r db.CommerceConnection) domain.CommerceConnection {
	return domain.CommerceConnection{
		ID: r.ID, UserID: r.UserID, Provider: r.Provider, Status: r.Status,
		StoreID: r.ExternalStoreID, StoreName: r.StoreName, StoreURL: r.StoreUrl, StoreCurrency: r.StoreCurrency,
		ConnectedAt: timestamptzPtr(r.ConnectedAt), LastSyncAt: timestamptzPtr(r.LastSyncAt),
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (r *commerceRepo) UpsertConnection(ctx context.Context, c domain.CommerceConnection, apiKeyEnc, webhookSecretEnc string) error {
	return r.s.q.UpsertCommerceConnection(ctx, db.UpsertCommerceConnectionParams{
		UserID: c.UserID, Provider: c.Provider, Status: c.Status,
		ApiKeyEnc: apiKeyEnc, WebhookSecretEnc: webhookSecretEnc,
		ExternalStoreID: c.StoreID, StoreName: c.StoreName, StoreUrl: c.StoreURL, StoreCurrency: c.StoreCurrency,
		ConnectedAt: timePtrToTimestamptz(c.ConnectedAt),
	})
}

func (r *commerceRepo) GetConnectionByProvider(ctx context.Context, userID, provider string) (domain.CommerceConnection, string, string, error) {
	row, err := r.s.q.GetCommerceConnection(ctx, db.GetCommerceConnectionParams{UserID: userID, Provider: provider})
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceConnection{}, "", "", domain.ErrNotFound
		}
		return domain.CommerceConnection{}, "", "", err
	}
	return commerceConnFromRow(row), row.ApiKeyEnc, row.WebhookSecretEnc, nil
}

func (r *commerceRepo) GetConnectionByStoreID(ctx context.Context, storeID string) (domain.CommerceConnection, string, string, error) {
	row, err := r.s.q.GetCommerceConnectionByStoreID(ctx, storeID)
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceConnection{}, "", "", domain.ErrNotFound
		}
		return domain.CommerceConnection{}, "", "", err
	}
	return commerceConnFromRow(row), row.ApiKeyEnc, row.WebhookSecretEnc, nil
}

func (r *commerceRepo) ListConnections(ctx context.Context, userID string) ([]domain.CommerceConnection, error) {
	rows, err := r.s.q.ListCommerceConnections(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CommerceConnection, 0, len(rows))
	for _, row := range rows {
		out = append(out, commerceConnFromRow(row))
	}
	return out, nil
}

func (r *commerceRepo) UpdateConnectionState(ctx context.Context, id, userID, status, lastError string, lastSyncAt, connectedAt *time.Time) error {
	return r.s.q.UpdateCommerceConnectionState(ctx, db.UpdateCommerceConnectionStateParams{
		ID: id, UserID: userID, Status: status, LastError: lastError,
		LastSyncAt: timePtrToTimestamptz(lastSyncAt), ConnectedAt: timePtrToTimestamptz(connectedAt),
	})
}

func (r *commerceRepo) DeleteConnection(ctx context.Context, id, userID string) error {
	return r.s.q.DeleteCommerceConnection(ctx, db.DeleteCommerceConnectionParams{ID: id, UserID: userID})
}

func (r *commerceRepo) UpsertProductLink(ctx context.Context, link domain.CommerceProductLink) (domain.CommerceProductLink, error) {
	row, err := r.s.q.UpsertCommerceProductLink(ctx, db.UpsertCommerceProductLinkParams{
		UserID: link.UserID, ConnectionID: link.ConnectionID, ProjectID: link.ProjectID,
		ExternalProductID: link.ExternalProductID, ExternalProductSlug: link.ExternalSlug,
		ExternalProductName: link.ExternalName, PriceMinor: link.PriceMinor, Currency: link.Currency,
		Status: link.Status, IsPublic: link.IsPublic, PublicToken: link.PublicToken,
	})
	if err != nil {
		return domain.CommerceProductLink{}, err
	}
	return productLinkFromRow(row), nil
}

func (r *commerceRepo) GetProductLinkByProject(ctx context.Context, userID, projectID string) (domain.CommerceProductLink, error) {
	row, err := r.s.q.GetCommerceProductLinkByProject(ctx, db.GetCommerceProductLinkByProjectParams{UserID: userID, ProjectID: projectID})
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceProductLink{}, domain.ErrNotFound
		}
		return domain.CommerceProductLink{}, err
	}
	return productLinkFromRow(row), nil
}

func (r *commerceRepo) GetProductLinkByExternal(ctx context.Context, connectionID, externalProductID string) (domain.CommerceProductLink, error) {
	row, err := r.s.q.GetCommerceProductLinkByExternal(ctx, db.GetCommerceProductLinkByExternalParams{ConnectionID: connectionID, ExternalProductID: externalProductID})
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceProductLink{}, domain.ErrNotFound
		}
		return domain.CommerceProductLink{}, err
	}
	return productLinkFromRow(row), nil
}

func (r *commerceRepo) GetProductLink(ctx context.Context, id, userID string) (domain.CommerceProductLink, error) {
	row, err := r.s.q.GetCommerceProductLink(ctx, db.GetCommerceProductLinkParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceProductLink{}, domain.ErrNotFound
		}
		return domain.CommerceProductLink{}, err
	}
	return productLinkFromRow(row), nil
}

func (r *commerceRepo) GetProductLinkByToken(ctx context.Context, token string) (domain.CommerceProductLink, error) {
	row, err := r.s.q.GetCommerceProductLinkByToken(ctx, token)
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceProductLink{}, domain.ErrNotFound
		}
		return domain.CommerceProductLink{}, err
	}
	return productLinkFromRow(row), nil
}

func (r *commerceRepo) UpdateProductLinkPublic(ctx context.Context, id, userID string, isPublic bool, publicToken string) error {
	return r.s.q.UpdateCommerceProductLinkPublic(ctx, db.UpdateCommerceProductLinkPublicParams{
		ID: id, UserID: userID, IsPublic: isPublic, PublicToken: publicToken,
	})
}

func (r *commerceRepo) DeleteProductLink(ctx context.Context, id, userID string) error {
	return r.s.q.DeleteCommerceProductLink(ctx, db.DeleteCommerceProductLinkParams{ID: id, UserID: userID})
}

func (r *commerceRepo) DeleteProductLinksByProject(ctx context.Context, userID, projectID string) error {
	return r.s.q.DeleteCommerceProductLinksByProject(ctx, db.DeleteCommerceProductLinksByProjectParams{UserID: userID, ProjectID: projectID})
}

func (r *commerceRepo) ListProductLinks(ctx context.Context, userID string) ([]domain.CommerceProductLink, error) {
	rows, err := r.s.q.ListCommerceProductLinks(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CommerceProductLink, 0, len(rows))
	for _, row := range rows {
		out = append(out, productLinkFromRow(row))
	}
	return out, nil
}

func productLinkFromRow(r db.CommerceProductLink) domain.CommerceProductLink {
	return domain.CommerceProductLink{
		ID: r.ID, UserID: r.UserID, ConnectionID: r.ConnectionID, ProjectID: r.ProjectID,
		ExternalProductID: r.ExternalProductID, ExternalSlug: r.ExternalProductSlug,
		ExternalName: r.ExternalProductName, PriceMinor: r.PriceMinor, Currency: r.Currency,
		Status: r.Status, IsPublic: r.IsPublic, PublicToken: r.PublicToken,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (r *commerceRepo) UpsertSale(ctx context.Context, sale domain.CommerceSale) (domain.CommerceSale, error) {
	row, err := r.s.q.UpsertCommerceSale(ctx, db.UpsertCommerceSaleParams{
		UserID: sale.UserID, ConnectionID: sale.ConnectionID,
		ProductLinkID:  uuidOrNull(valueOrEmpty(sale.ProductLinkID)),
		ExternalSaleID: sale.ExternalID, Status: sale.Status,
		AmountMinor: sale.AmountMinor, Currency: sale.Currency,
		BuyerEmail: sale.BuyerEmail, BuyerName: sale.BuyerName,
		CheckoutUrl:    sale.CheckoutURL,
		CustomMetadata: orEmptyJSON(sale.Metadata),
		Payload:        orEmptyJSON(sale.Payload),
		CompletedAt:    timePtrToTimestamptz(sale.CompletedAt),
	})
	if err != nil {
		return domain.CommerceSale{}, err
	}
	return commerceSaleFromRow(row), nil
}

func (r *commerceRepo) GetSaleByExternalID(ctx context.Context, connectionID, externalID string) (domain.CommerceSale, error) {
	row, err := r.s.q.GetCommerceSaleByExternalID(ctx, db.GetCommerceSaleByExternalIDParams{ConnectionID: connectionID, ExternalSaleID: externalID})
	if err != nil {
		if isNoRows(err) {
			return domain.CommerceSale{}, domain.ErrNotFound
		}
		return domain.CommerceSale{}, err
	}
	return commerceSaleFromRow(row), nil
}

func (r *commerceRepo) ListSalesByProductLink(ctx context.Context, linkID, userID string, limit int) ([]domain.CommerceSale, error) {
	rows, err := r.s.q.ListCommerceSalesByProductLink(ctx, db.ListCommerceSalesByProductLinkParams{
		ProductLinkID: uuidOrNull(linkID), UserID: userID, Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CommerceSale, 0, len(rows))
	for _, row := range rows {
		out = append(out, commerceSaleFromRow(row))
	}
	return out, nil
}

func (r *commerceRepo) ListSalesByUser(ctx context.Context, userID string, limit int) ([]domain.CommerceSale, error) {
	rows, err := r.s.q.ListCommerceSalesByUser(ctx, db.ListCommerceSalesByUserParams{UserID: userID, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CommerceSale, 0, len(rows))
	for _, row := range rows {
		out = append(out, commerceSaleFromRow(row))
	}
	return out, nil
}

func (r *commerceRepo) CreateWebhookEvent(ctx context.Context, userID, provider, pulseID, deliveryID, eventType string, payload []byte, status string) (bool, error) {
	_, err := r.s.q.InsertCommerceWebhookEvent(ctx, db.InsertCommerceWebhookEventParams{
		UserID: userID, Provider: provider, ExternalPulseID: pulseID,
		ExternalDeliveryID: deliveryID, EventType: eventType, Payload: orEmptyJSON(payload), Status: status,
	})
	if err != nil {
		if isNoRows(err) {
			return false, nil // livraison déjà traitée
		}
		return false, err
	}
	return true, nil
}

func commerceSaleFromRow(r db.CommerceSale) domain.CommerceSale {
	return domain.CommerceSale{
		ID: r.ID, UserID: r.UserID, ConnectionID: r.ConnectionID,
		ProductLinkID: productLinkIDPtr(r.ProductLinkID),
		ExternalID:    r.ExternalSaleID, Status: r.Status,
		AmountMinor: r.AmountMinor, Currency: r.Currency,
		BuyerEmail: r.BuyerEmail, BuyerName: r.BuyerName,
		CheckoutURL: r.CheckoutUrl, Metadata: r.CustomMetadata, Payload: r.Payload,
		CompletedAt: timestamptzPtr(r.CompletedAt), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func valueOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func productLinkIDPtr(u pgtype.UUID) *string {
	s := uuidString(u)
	if s == "" {
		return nil
	}
	return &s
}
