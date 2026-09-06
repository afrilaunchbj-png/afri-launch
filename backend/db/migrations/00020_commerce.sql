-- +goose Up
-- Module commerce (Chariow, ADR-019) : chaque utilisateur connecte sa propre
-- boutique via une API key (chiffrée AES-GCM au repos). Ventes + webhooks.

-- Connexion d'un utilisateur à une plateforme de commerce (1 par provider).
CREATE TABLE commerce_connections (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider             TEXT NOT NULL DEFAULT 'chariow'
                         CHECK (provider IN ('chariow')),
    status               TEXT NOT NULL DEFAULT 'connected'
                         CHECK (status IN ('connected','disconnected','error','reconnect_required')),
    api_key_enc          TEXT NOT NULL DEFAULT '',
    webhook_secret_enc   TEXT NOT NULL DEFAULT '',
    external_store_id    TEXT NOT NULL DEFAULT '',
    store_name           TEXT NOT NULL DEFAULT '',
    store_url            TEXT NOT NULL DEFAULT '',
    store_currency       TEXT NOT NULL DEFAULT '',
    last_error           TEXT NOT NULL DEFAULT '',
    last_error_at        TIMESTAMPTZ,
    last_sync_at         TIMESTAMPTZ,
    connected_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, provider)
);

CREATE INDEX idx_commerce_connections_user ON commerce_connections (user_id);
CREATE INDEX idx_commerce_connections_store ON commerce_connections (external_store_id)
    WHERE external_store_id <> '';

-- Association d'un projet SaaS à un produit Chariow existant (l'API publique
-- ne permet pas de créer des produits — mapping créé manuellement).
CREATE TABLE commerce_product_links (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id        UUID NOT NULL REFERENCES commerce_connections(id) ON DELETE CASCADE,
    project_id           UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    external_product_id  TEXT NOT NULL,
    external_product_slug TEXT NOT NULL DEFAULT '',
    external_product_name TEXT NOT NULL DEFAULT '',
    price_minor          BIGINT NOT NULL DEFAULT 0,
    currency             TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'linked'
                         CHECK (status IN ('linked','unlinked','unavailable')),
    is_public            BOOLEAN NOT NULL DEFAULT false,
    public_token         TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, external_product_id)
);

CREATE INDEX idx_commerce_links_user ON commerce_product_links (user_id, created_at DESC);
CREATE INDEX idx_commerce_links_project ON commerce_product_links (user_id, project_id);
CREATE UNIQUE INDEX idx_commerce_links_public_token ON commerce_product_links (public_token)
    WHERE public_token <> '';

-- Ventes enregistrées (init checkout ou webhooks/sync).
CREATE TABLE commerce_sales (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id        UUID NOT NULL REFERENCES commerce_connections(id) ON DELETE CASCADE,
    product_link_id      UUID REFERENCES commerce_product_links(id) ON DELETE SET NULL,
    external_sale_id     TEXT NOT NULL,
    status               TEXT NOT NULL DEFAULT 'awaiting_payment',
    amount_minor         BIGINT NOT NULL DEFAULT 0,
    currency             TEXT NOT NULL DEFAULT '',
    buyer_email          TEXT NOT NULL DEFAULT '',
    buyer_name           TEXT NOT NULL DEFAULT '',
    checkout_url         TEXT NOT NULL DEFAULT '',
    custom_metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    payload              JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at         TIMESTAMPTZ,
    UNIQUE (connection_id, external_sale_id)
);

CREATE INDEX idx_commerce_sales_user ON commerce_sales (user_id, created_at DESC);
CREATE INDEX idx_commerce_sales_link ON commerce_sales (product_link_id, created_at DESC);

-- Idempotence des webhooks (une livraison = un traitement).
CREATE TABLE commerce_webhook_events (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider             TEXT NOT NULL DEFAULT 'chariow',
    external_pulse_id    TEXT NOT NULL DEFAULT '',
    external_delivery_id TEXT NOT NULL DEFAULT '',
    event_type           TEXT NOT NULL DEFAULT '',
    payload              JSONB NOT NULL DEFAULT '{}'::jsonb,
    status               TEXT NOT NULL DEFAULT 'processed'
                         CHECK (status IN ('processed','failed','ignored')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, external_delivery_id)
);

CREATE INDEX idx_commerce_webhook_events_user ON commerce_webhook_events (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS commerce_webhook_events;
DROP TABLE IF EXISTS commerce_sales;
DROP TABLE IF EXISTS commerce_product_links;
DROP TABLE IF EXISTS commerce_connections;
