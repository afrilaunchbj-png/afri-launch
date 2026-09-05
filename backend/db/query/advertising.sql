-- name: CreateAdConnection :one
INSERT INTO ad_platform_connections (user_id, provider, status, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateAdConnectionTokens :one
UPDATE ad_platform_connections
SET access_token_enc = $2,
    refresh_token_enc = $3,
    access_token_expires_at = $4,
    external_user_id = $5,
    scopes = $6,
    status = $7,
    last_error = '',
    last_error_at = NULL,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SelectAdAccount :one
UPDATE ad_platform_connections
SET external_account_id = $3,
    external_account_name = $4,
    metadata = metadata || $5,
    status = 'active',
    updated_at = now()
WHERE id = $2 AND user_id = $1
RETURNING *;

-- name: SetAdConnectionStatus :one
UPDATE ad_platform_connections
SET status = $3,
    last_error = $4,
    last_error_at = CASE WHEN $4 <> '' THEN now() ELSE last_error_at END,
    updated_at = now()
WHERE id = $2 AND user_id = $1
RETURNING *;

-- name: SetAdConnectionSynced :one
UPDATE ad_platform_connections
SET last_sync_at = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetAdConnection :one
SELECT * FROM ad_platform_connections WHERE id = $1 AND user_id = $2;

-- name: GetAdConnectionByProvider :one
SELECT * FROM ad_platform_connections WHERE user_id = $1 AND provider = $2;

-- name: ListAdConnections :many
SELECT * FROM ad_platform_connections WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListAdConnectionsByProvider :many
SELECT * FROM ad_platform_connections WHERE provider = $1 ORDER BY updated_at DESC;

-- name: CreateOAuthState :exec
INSERT INTO oauth_states (state, user_id, provider, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeOAuthState :one
UPDATE oauth_states
SET used_at = now()
WHERE state = $1 AND provider = $2
  AND used_at IS NULL AND expires_at > now()
RETURNING *;

-- name: PruneOAuthStates :exec
DELETE FROM oauth_states WHERE expires_at < now() - interval '1 day';

-- name: UpsertAdCampaign :one
INSERT INTO ad_campaigns (user_id, connection_id, external_campaign_id, name, objective, status, budget_minor, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (connection_id, external_campaign_id) DO UPDATE
SET name = EXCLUDED.name,
    objective = EXCLUDED.objective,
    status = EXCLUDED.status,
    budget_minor = EXCLUDED.budget_minor,
    currency = EXCLUDED.currency,
    updated_at = now()
RETURNING *;

-- name: UpdateAdCampaign :one
UPDATE ad_campaigns
SET name = COALESCE(NULLIF($3, ''), name),
    status = COALESCE(NULLIF($4, ''), status),
    budget_minor = COALESCE(NULLIF($5::bigint, 0), budget_minor),
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetAdCampaign :one
SELECT * FROM ad_campaigns WHERE id = $1 AND user_id = $2;

-- name: ListAdCampaigns :many
SELECT * FROM ad_campaigns WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListAdCampaignsByConnection :many
SELECT * FROM ad_campaigns WHERE connection_id = $1 ORDER BY created_at DESC;

-- name: CreateAdCreative :one
INSERT INTO ad_creatives (user_id, connection_id, campaign_id, type, asset_id, headline, primary_text, cta, status, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateAdCreativeExternal :one
UPDATE ad_creatives
SET external_creative_id = $3, status = $4, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetAdCreative :one
SELECT * FROM ad_creatives WHERE id = $1 AND user_id = $2;

-- name: ListAdCreatives :many
SELECT * FROM ad_creatives WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpsertAdInsight :exec
INSERT INTO ad_insights (campaign_id, user_id, date, impressions, reach, clicks, spend_minor, conversions, currency, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (campaign_id, date) DO UPDATE
SET impressions = EXCLUDED.impressions,
    reach = EXCLUDED.reach,
    clicks = EXCLUDED.clicks,
    spend_minor = EXCLUDED.spend_minor,
    conversions = EXCLUDED.conversions,
    currency = EXCLUDED.currency,
    metadata = EXCLUDED.metadata;

-- name: ListAdInsights :many
SELECT * FROM ad_insights
WHERE campaign_id = $1 AND user_id = $2
  AND date >= $3::date AND date <= $4::date
ORDER BY date ASC;

-- name: CreateProviderOperation :one
INSERT INTO provider_operations (user_id, connection_id, provider, operation_type, status, internal_resource_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: MarkProviderOperationProcessing :exec
UPDATE provider_operations
SET status = 'processing', started_at = now(), attempts = attempts + 1
WHERE id = $1;

-- name: CompleteProviderOperation :exec
UPDATE provider_operations
SET status = 'completed', external_resource_id = $2, completed_at = now()
WHERE id = $1;

-- name: FailProviderOperation :exec
UPDATE provider_operations
SET status = 'failed', error_code = $2, error_message = $3, completed_at = now()
WHERE id = $1;

-- name: GetProviderOperation :one
SELECT * FROM provider_operations WHERE id = $1 AND user_id = $2;

-- name: ListProviderOperations :many
SELECT * FROM provider_operations WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;
