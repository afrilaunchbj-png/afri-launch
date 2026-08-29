-- +goose Up
-- Facturation : packs de crédits (plans) et paiements.
-- Le paiement effectif (Mobile Money) est une étape ultérieure ; ces tables
-- posent le modèle pour les « recharges » sans coupler au provider.
CREATE TABLE plans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    credits     INTEGER NOT NULL CHECK (credits > 0),
    price_minor INTEGER NOT NULL CHECK (price_minor >= 0),
    currency    TEXT NOT NULL DEFAULT 'XOF',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id            UUID REFERENCES plans(id),
    amount_minor       INTEGER NOT NULL,
    currency           TEXT NOT NULL,
    provider           TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending', 'succeeded', 'failed', 'refunded')),
    idempotency_key    TEXT NOT NULL,
    provider_reference TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_payments_idempotency ON payments (idempotency_key);
CREATE INDEX idx_payments_user ON payments (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS plans;
