package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// adConnectionRepo implémente port.AdConnectionRepository. Les tokens sont
// chiffrés au repos via le SecretEncryptor injecté (jamais en clair en DB).
type adConnectionRepo struct {
	s   *Store
	enc portSecretEncryptor
}

// portSecretEncryptor évite un import cyclique : interface locale = port.SecretEncryptor.
type portSecretEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// NewAdConnectionRepository construit le repository de connexions.
func NewAdConnectionRepository(s *Store, enc portSecretEncryptor) *adConnectionRepo {
	return &adConnectionRepo{s: s, enc: enc}
}

// orEmptyJSON garantit un JSONB non-null pour les colonnes NOT NULL DEFAULT '{}'.
func orEmptyJSON(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func (r *adConnectionRepo) Create(ctx context.Context, c domain.AdPlatformConnection) (domain.AdPlatformConnection, error) {
	row, err := r.s.q.CreateAdConnection(ctx, db.CreateAdConnectionParams{
		UserID:   c.UserID,
		Provider: c.Provider,
		Status:   c.Status,
		Metadata: orEmptyJSON(c.Metadata),
	})
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return r.fromRow(row), nil
}

func (r *adConnectionRepo) UpdateTokens(ctx context.Context, id string, accessEnc, refreshEnc string, expiresAt *time.Time, externalUserID string, scopes []string, status string) error {
	row, err := r.s.q.UpdateAdConnectionTokens(ctx, db.UpdateAdConnectionTokensParams{
		ID:                   id,
		AccessTokenEnc:       accessEnc,
		RefreshTokenEnc:      refreshEnc,
		AccessTokenExpiresAt: toTimestamptz(expiresAt),
		ExternalUserID:       externalUserID,
		Scopes:               scopes,
		Status:               status,
	})
	if err != nil {
		return err
	}
	_ = row
	return nil
}

func (r *adConnectionRepo) SelectAccount(ctx context.Context, userID, id, externalAccountID, accountName string, metadata []byte) error {
	_, err := r.s.q.SelectAdAccount(ctx, db.SelectAdAccountParams{
		UserID:              userID,
		ID:                  id,
		ExternalAccountID:   externalAccountID,
		ExternalAccountName: accountName,
		Metadata:            metadata,
	})
	return err
}

func (r *adConnectionRepo) SetStatus(ctx context.Context, userID, id, status, lastError string) error {
	_, err := r.s.q.SetAdConnectionStatus(ctx, db.SetAdConnectionStatusParams{
		UserID:    userID,
		ID:        id,
		Status:    status,
		LastError: lastError,
	})
	return err
}

func (r *adConnectionRepo) SetSynced(ctx context.Context, id string, at time.Time) error {
	_, err := r.s.q.SetAdConnectionSynced(ctx, db.SetAdConnectionSyncedParams{
		ID:         id,
		LastSyncAt: pgtype.Timestamptz{Time: at, Valid: true},
	})
	return err
}

func (r *adConnectionRepo) Get(ctx context.Context, userID, id string) (domain.AdPlatformConnection, error) {
	row, err := r.s.q.GetAdConnection(ctx, db.GetAdConnectionParams{ID: id, UserID: userID})
	if isNoRows(err) {
		return domain.AdPlatformConnection{}, domain.ErrConnectionNotFound
	}
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return r.fromRow(row), nil
}

func (r *adConnectionRepo) GetByProvider(ctx context.Context, userID, provider string) (domain.AdPlatformConnection, error) {
	row, err := r.s.q.GetAdConnectionByProvider(ctx, db.GetAdConnectionByProviderParams{UserID: userID, Provider: provider})
	if isNoRows(err) {
		return domain.AdPlatformConnection{}, domain.ErrConnectionNotFound
	}
	if err != nil {
		return domain.AdPlatformConnection{}, err
	}
	return r.fromRow(row), nil
}

func (r *adConnectionRepo) List(ctx context.Context, userID string) ([]domain.AdPlatformConnection, error) {
	rows, err := r.s.q.ListAdConnections(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AdPlatformConnection, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.fromRow(row))
	}
	return out, nil
}

func (r *adConnectionRepo) ListByProvider(ctx context.Context, provider string) ([]domain.AdPlatformConnection, error) {
	rows, err := r.s.q.ListAdConnectionsByProvider(ctx, provider)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AdPlatformConnection, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.fromRow(row))
	}
	return out, nil
}

// fromRow déchiffre les tokens en mémoire (jamais exposés au frontend).
func (r *adConnectionRepo) fromRow(row db.AdPlatformConnection) domain.AdPlatformConnection {
	access, _ := r.enc.Decrypt(row.AccessTokenEnc)
	refresh, _ := r.enc.Decrypt(row.RefreshTokenEnc)
	return domain.AdPlatformConnection{
		ID:                   row.ID,
		UserID:               row.UserID,
		Provider:             row.Provider,
		Status:               row.Status,
		ExternalUserID:       row.ExternalUserID,
		ExternalAccountID:    row.ExternalAccountID,
		ExternalAccountName:  row.ExternalAccountName,
		AccessToken:          access,
		RefreshToken:         refresh,
		AccessTokenExpiresAt: fromTimestamptz(row.AccessTokenExpiresAt),
		Scopes:               row.Scopes,
		Metadata:             row.Metadata,
		LastError:            row.LastError,
		LastErrorAt:          fromTimestamptz(row.LastErrorAt),
		LastSyncAt:           fromTimestamptz(row.LastSyncAt),
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

// oauthStateStore implémente port.OAuthStateStore.
type oauthStateStore struct {
	s *Store
}

// NewOAuthStateStore construit le store d'états OAuth.
func NewOAuthStateStore(s *Store) *oauthStateStore { return &oauthStateStore{s: s} }

func (r *oauthStateStore) Create(ctx context.Context, s domain.OAuthState) error {
	// Ménage opportuniste des états expirés.
	_ = r.s.q.PruneOAuthStates(ctx)
	return r.s.q.CreateOAuthState(ctx, db.CreateOAuthStateParams{
		State:     s.State,
		UserID:    s.UserID,
		Provider:  s.Provider,
		ExpiresAt: s.ExpiresAt,
	})
}

func (r *oauthStateStore) Consume(ctx context.Context, state, provider string) (domain.OAuthState, error) {
	row, err := r.s.q.ConsumeOAuthState(ctx, db.ConsumeOAuthStateParams{State: state, Provider: provider})
	if isNoRows(err) {
		return domain.OAuthState{}, domain.ErrOAuthStateInvalid
	}
	if err != nil {
		return domain.OAuthState{}, err
	}
	return domain.OAuthState{
		State:     row.State,
		UserID:    row.UserID,
		Provider:  row.Provider,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    fromTimestamptz(row.UsedAt),
	}, nil
}

func (r *oauthStateStore) Prune(ctx context.Context) error { return r.s.q.PruneOAuthStates(ctx) }

// adCampaignRepo implémente port.AdCampaignRepository.
type adCampaignRepo struct {
	s *Store
}

// NewAdCampaignRepository construit le repository de campagnes.
func NewAdCampaignRepository(s *Store) *adCampaignRepo { return &adCampaignRepo{s: s} }

func (r *adCampaignRepo) Upsert(ctx context.Context, c domain.Campaign) (domain.Campaign, error) {
	row, err := r.s.q.UpsertAdCampaign(ctx, db.UpsertAdCampaignParams{
		UserID:             c.UserID,
		ConnectionID:       c.ConnectionID,
		ExternalCampaignID: c.ExternalCampaignID,
		Name:               c.Name,
		Objective:          c.Objective,
		Status:             c.Status,
		BudgetMinor:        c.BudgetMinor,
		Currency:           c.Currency,
	})
	if err != nil {
		return domain.Campaign{}, err
	}
	return toCampaign(row), nil
}

func (r *adCampaignRepo) Update(ctx context.Context, userID, id string, c domain.Campaign) (domain.Campaign, error) {
	row, err := r.s.q.UpdateAdCampaign(ctx, db.UpdateAdCampaignParams{
		ID:      id,
		UserID:  userID,
		Column3: c.Name,
		Column4: c.Status,
		Column5: c.BudgetMinor,
	})
	if isNoRows(err) {
		return domain.Campaign{}, domain.ErrConnectionNotFound
	}
	if err != nil {
		return domain.Campaign{}, err
	}
	return toCampaign(row), nil
}

func (r *adCampaignRepo) Get(ctx context.Context, userID, id string) (domain.Campaign, error) {
	row, err := r.s.q.GetAdCampaign(ctx, db.GetAdCampaignParams{ID: id, UserID: userID})
	if isNoRows(err) {
		return domain.Campaign{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Campaign{}, err
	}
	return toCampaign(row), nil
}

func (r *adCampaignRepo) List(ctx context.Context, userID string) ([]domain.Campaign, error) {
	rows, err := r.s.q.ListAdCampaigns(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Campaign, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCampaign(row))
	}
	return out, nil
}

func (r *adCampaignRepo) ListByConnection(ctx context.Context, connectionID string) ([]domain.Campaign, error) {
	rows, err := r.s.q.ListAdCampaignsByConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Campaign, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCampaign(row))
	}
	return out, nil
}

func toCampaign(row db.AdCampaign) domain.Campaign {
	return domain.Campaign{
		ID:                 row.ID,
		UserID:             row.UserID,
		ConnectionID:       row.ConnectionID,
		ExternalCampaignID: row.ExternalCampaignID,
		Name:               row.Name,
		Objective:          row.Objective,
		Status:             row.Status,
		BudgetMinor:        row.BudgetMinor,
		Currency:           row.Currency,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// adCreativeRepo implémente port.AdCreativeRepository.
type adCreativeRepo struct {
	s *Store
}

// NewAdCreativeRepository construit le repository de creatives.
func NewAdCreativeRepository(s *Store) *adCreativeRepo { return &adCreativeRepo{s: s} }

func (r *adCreativeRepo) Create(ctx context.Context, c domain.Creative) (domain.Creative, error) {
	row, err := r.s.q.CreateAdCreative(ctx, db.CreateAdCreativeParams{
		UserID:       c.UserID,
		ConnectionID: c.ConnectionID,
		CampaignID:   toUUIDPtr(c.CampaignID),
		Type:         c.Type,
		AssetID:      toUUIDPtr(c.AssetID),
		Headline:     c.Headline,
		PrimaryText:  c.PrimaryText,
		Cta:          c.CTA,
		Status:       c.Status,
		Metadata:     orEmptyJSON(c.Metadata),
	})
	if err != nil {
		return domain.Creative{}, err
	}
	return toCreative(row), nil
}

func (r *adCreativeRepo) UpdateExternal(ctx context.Context, userID, id, externalCreativeID, status string) error {
	_, err := r.s.q.UpdateAdCreativeExternal(ctx, db.UpdateAdCreativeExternalParams{
		ID: id, UserID: userID, ExternalCreativeID: externalCreativeID, Status: status,
	})
	return err
}

func (r *adCreativeRepo) Get(ctx context.Context, userID, id string) (domain.Creative, error) {
	row, err := r.s.q.GetAdCreative(ctx, db.GetAdCreativeParams{ID: id, UserID: userID})
	if isNoRows(err) {
		return domain.Creative{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Creative{}, err
	}
	return toCreative(row), nil
}

func (r *adCreativeRepo) List(ctx context.Context, userID string) ([]domain.Creative, error) {
	rows, err := r.s.q.ListAdCreatives(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Creative, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCreative(row))
	}
	return out, nil
}

func toCreative(row db.AdCreative) domain.Creative {
	return domain.Creative{
		ID:                 row.ID,
		UserID:             row.UserID,
		CampaignID:         fromUUIDPtr(row.CampaignID),
		ConnectionID:       row.ConnectionID,
		Type:               row.Type,
		AssetID:            fromUUIDPtr(row.AssetID),
		ExternalCreativeID: row.ExternalCreativeID,
		Headline:           row.Headline,
		PrimaryText:        row.PrimaryText,
		CTA:                row.Cta,
		Status:             row.Status,
		Metadata:           row.Metadata,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// adInsightRepo implémente port.AdInsightRepository.
type adInsightRepo struct {
	s *Store
}

// NewAdInsightRepository construit le repository d'insights.
func NewAdInsightRepository(s *Store) *adInsightRepo { return &adInsightRepo{s: s} }

func (r *adInsightRepo) Upsert(ctx context.Context, campaignID, userID string, insights []domain.Insight) error {
	for _, in := range insights {
		if err := r.s.q.UpsertAdInsight(ctx, db.UpsertAdInsightParams{
			CampaignID:  campaignID,
			UserID:      userID,
			Date:        toDate(in.Date),
			Impressions: in.Impressions,
			Reach:       in.Reach,
			Clicks:      in.Clicks,
			SpendMinor:  in.SpendMinor,
			Conversions: in.Conversions,
			Currency:    in.Currency,
			Metadata:    orEmptyJSON(in.Metadata),
		}); err != nil {
			return err
		}
	}
	return nil
}

// toDate convertit "YYYY-MM-DD" en pgtype.Date.
func toDate(s string) pgtype.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: t.UTC(), Valid: true}
}

func (r *adInsightRepo) ListByCampaign(ctx context.Context, userID, campaignID string, since, until string) ([]domain.Insight, error) {
	rows, err := r.s.q.ListAdInsights(ctx, db.ListAdInsightsParams{
		CampaignID: campaignID,
		UserID:     userID,
		Column3:    toDate(since),
		Column4:    toDate(until),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Insight, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Insight{
			CampaignID:  row.CampaignID,
			Date:        row.Date.Time.Format("2006-01-02"),
			Impressions: row.Impressions,
			Reach:       row.Reach,
			Clicks:      row.Clicks,
			SpendMinor:  row.SpendMinor,
			Conversions: row.Conversions,
			Currency:    row.Currency,
			Metadata:    row.Metadata,
		})
	}
	return out, nil
}

// providerOperationRepo implémente port.ProviderOperationRepository.
type providerOperationRepo struct {
	s *Store
}

// NewProviderOperationRepository construit le repository d'opérations provider.
func NewProviderOperationRepository(s *Store) *providerOperationRepo {
	return &providerOperationRepo{s: s}
}

func (r *providerOperationRepo) Create(ctx context.Context, op domain.ProviderOperation) (domain.ProviderOperation, error) {
	row, err := r.s.q.CreateProviderOperation(ctx, db.CreateProviderOperationParams{
		UserID:             op.UserID,
		ConnectionID:       toUUIDPtr(&op.ConnectionID),
		Provider:           op.Provider,
		OperationType:      op.OperationType,
		Status:             op.Status,
		InternalResourceID: op.InternalResourceID,
	})
	if err != nil {
		return domain.ProviderOperation{}, err
	}
	return toProviderOperation(row), nil
}

func (r *providerOperationRepo) MarkProcessing(ctx context.Context, id string) error {
	return r.s.q.MarkProviderOperationProcessing(ctx, id)
}

func (r *providerOperationRepo) Complete(ctx context.Context, id, externalResourceID string) error {
	return r.s.q.CompleteProviderOperation(ctx, db.CompleteProviderOperationParams{
		ID:                 id,
		ExternalResourceID: externalResourceID,
	})
}

func (r *providerOperationRepo) Fail(ctx context.Context, id, code, message string) error {
	return r.s.q.FailProviderOperation(ctx, db.FailProviderOperationParams{
		ID:           id,
		ErrorCode:    code,
		ErrorMessage: message,
	})
}

func (r *providerOperationRepo) IncrementAttempts(ctx context.Context, id string) error {
	return r.s.q.MarkProviderOperationProcessing(ctx, id)
}

func (r *providerOperationRepo) Get(ctx context.Context, userID, id string) (domain.ProviderOperation, error) {
	row, err := r.s.q.GetProviderOperation(ctx, db.GetProviderOperationParams{ID: id, UserID: userID})
	if isNoRows(err) {
		return domain.ProviderOperation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ProviderOperation{}, err
	}
	return toProviderOperation(row), nil
}

func (r *providerOperationRepo) List(ctx context.Context, userID string, limit int) ([]domain.ProviderOperation, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.s.q.ListProviderOperations(ctx, db.ListProviderOperationsParams{UserID: userID, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProviderOperation, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProviderOperation(row))
	}
	return out, nil
}

func toProviderOperation(row db.ProviderOperation) domain.ProviderOperation {
	connID := ""
	if row.ConnectionID.Valid {
		connID = uuid.UUID(row.ConnectionID.Bytes).String()
	}
	return domain.ProviderOperation{
		ID:                 row.ID,
		UserID:             row.UserID,
		ConnectionID:       connID,
		Provider:           row.Provider,
		OperationType:      row.OperationType,
		Status:             row.Status,
		Attempts:           int(row.Attempts),
		InternalResourceID: row.InternalResourceID,
		ExternalResourceID: row.ExternalResourceID,
		ErrorCode:          row.ErrorCode,
		ErrorMessage:       row.ErrorMessage,
		CreatedAt:          row.CreatedAt,
		StartedAt:          fromTimestamptz(row.StartedAt),
		CompletedAt:        fromTimestamptz(row.CompletedAt),
	}
}

// toUUIDPtr / fromUUIDPtr convertissent *string ↔ pgtype.UUID pour les FK nullables.
func toUUIDPtr(s *string) pgtype.UUID {
	if s == nil || *s == "" {
		return pgtype.UUID{}
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

func fromUUIDPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

// toTimestamptz / fromTimestamptz : helpers *time.Time ↔ pgtype.
func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// optionalUUID convertit un filtre texte (peut être vide) en UUID nullable —
// une chaîne vide ne doit jamais produire d'erreur de cast côté Postgres.
func optionalUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}
