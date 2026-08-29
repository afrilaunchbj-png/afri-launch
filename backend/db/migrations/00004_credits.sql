-- +goose Up
-- Ledger de crédits (double-entrée, idempotent).
--   credit_accounts      : solde par utilisateur (balance = crédits possédés,
--                          reserved = crédits bloqués par une réservation,
--                          available = balance - reserved).
--   credit_transactions  : journal comptable (append-only).
--   credit_reservations  : blocages temporaires (cycle reserve → consume|release).
--   generation_costs     : coûts configurables par opération.

CREATE TABLE credit_accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance    INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    reserved   INTEGER NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (balance >= reserved)
);

CREATE UNIQUE INDEX uq_credit_accounts_user ON credit_accounts (user_id);

CREATE TABLE credit_transactions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('credit', 'debit')),
    amount     INTEGER NOT NULL CHECK (amount > 0),
    operation  TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'pending', 'failed')),
    reference  TEXT,
    metadata   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_credit_transactions_account ON credit_transactions (account_id, created_at DESC);
CREATE UNIQUE INDEX uq_credit_transactions_reference ON credit_transactions (reference) WHERE reference IS NOT NULL;

CREATE TABLE credit_reservations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    amount     INTEGER NOT NULL CHECK (amount > 0),
    operation  TEXT NOT NULL,
    reference  TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'reserved' CHECK (status IN ('reserved', 'consumed', 'released')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_credit_reservations_reference ON credit_reservations (reference);
CREATE INDEX idx_credit_reservations_account ON credit_reservations (account_id, created_at DESC);

CREATE TABLE generation_costs (
    operation  TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    credits    INTEGER NOT NULL CHECK (credits > 0),
    is_active  BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS generation_costs;
DROP TABLE IF EXISTS credit_reservations;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credit_accounts;
