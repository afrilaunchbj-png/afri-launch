-- +goose Up
-- Module advertising (ADR-017) : connexions OAuth multi-tenant aux
-- plateformes publicitaires, campagnes, creatives, insights, opérations.

-- Connexion d'un utilisateur à une plateforme (1 par plateforme, extensible).
-- Les tokens OAuth sont chiffrés au repos (AES-GCM, cf. infra/crypto) —
-- jamais en clair, jamais exposés au frontend.
CREATE TABLE ad_platform_connections (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider                TEXT NOT NULL
                            CHECK (provider IN ('meta','google_ads','tiktok_ads')),
    status                  TEXT NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending','active','expired','revoked','error','disconnected')),
    external_user_id        TEXT NOT NULL DEFAULT '',
    external_account_id     TEXT NOT NULL DEFAULT '',
    external_account_name   TEXT NOT NULL DEFAULT '',
    access_token_enc        TEXT NOT NULL DEFAULT '',
    refresh_token_enc       TEXT NOT NULL DEFAULT '',
    access_token_expires_at TIMESTAMPTZ,
    scopes                  TEXT[] NOT NULL DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',
    last_error              TEXT NOT NULL DEFAULT '',
    last_error_at           TIMESTAMPTZ,
    last_sync_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, provider)
);

CREATE INDEX idx_ad_connections_user ON ad_platform_connections (user_id);
CREATE INDEX idx_ad_connections_provider_status ON ad_platform_connections (provider, status);

-- États CSRF OAuth (liés user+provider, à usage unique, TTL court).
CREATE TABLE oauth_states (
    state      TEXT PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ
);

-- Campagnes : ID interne UUID, mapping vers l'ID externe provider.
CREATE TABLE ad_campaigns (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id        UUID NOT NULL REFERENCES ad_platform_connections(id) ON DELETE CASCADE,
    external_campaign_id TEXT NOT NULL DEFAULT '',
    name                 TEXT NOT NULL,
    objective            TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft','active','paused','deleted')),
    budget_minor         BIGINT NOT NULL DEFAULT 0 CHECK (budget_minor >= 0),
    currency             TEXT NOT NULL DEFAULT 'XOF',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, external_campaign_id)
);

CREATE INDEX idx_ad_campaigns_user ON ad_campaigns (user_id, created_at DESC);

-- Creatives : visuels liés aux assets internes (vidéos du pipeline ADR-016).
CREATE TABLE ad_creatives (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id        UUID NOT NULL REFERENCES ad_platform_connections(id) ON DELETE CASCADE,
    campaign_id          UUID REFERENCES ad_campaigns(id) ON DELETE SET NULL,
    type                 TEXT NOT NULL CHECK (type IN ('video','image','carousel','text')),
    asset_id             UUID REFERENCES assets(id) ON DELETE SET NULL,
    external_creative_id TEXT NOT NULL DEFAULT '',
    headline             TEXT NOT NULL DEFAULT '',
    primary_text         TEXT NOT NULL DEFAULT '',
    cta                  TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft','uploading','active','error')),
    metadata             JSONB NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ad_creatives_user ON ad_creatives (user_id, created_at DESC);

-- Insights de performance (metrics normalisées + metadata brute provider).
-- Montants en unités mineures entières (jamais de float).
CREATE TABLE ad_insights (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    impressions  BIGINT NOT NULL DEFAULT 0,
    reach        BIGINT NOT NULL DEFAULT 0,
    clicks       BIGINT NOT NULL DEFAULT 0,
    spend_minor  BIGINT NOT NULL DEFAULT 0,
    conversions  NUMERIC NOT NULL DEFAULT 0,
    currency     TEXT NOT NULL DEFAULT 'XOF',
    metadata     JSONB NOT NULL DEFAULT '{}',
    UNIQUE (campaign_id, date)
);

-- Traçabilité des opérations mutatives (idempotence, retry, reconciliation).
CREATE TABLE provider_operations (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id        UUID REFERENCES ad_platform_connections(id) ON DELETE SET NULL,
    provider             TEXT NOT NULL,
    operation_type       TEXT NOT NULL,
    status               TEXT NOT NULL DEFAULT 'pending'
                         CHECK (status IN ('pending','processing','completed','failed')),
    attempts             INT NOT NULL DEFAULT 0,
    internal_resource_id TEXT NOT NULL DEFAULT '',
    external_resource_id TEXT NOT NULL DEFAULT '',
    error_code           TEXT NOT NULL DEFAULT '',
    error_message        TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at           TIMESTAMPTZ,
    completed_at         TIMESTAMPTZ
);

CREATE INDEX idx_provider_operations_user ON provider_operations (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS provider_operations;
DROP TABLE IF EXISTS ad_insights;
DROP TABLE IF EXISTS ad_creatives;
DROP TABLE IF EXISTS ad_campaigns;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS ad_platform_connections;
