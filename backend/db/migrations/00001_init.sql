-- +goose Up
-- Extension pour générer des UUID côté base (v4) et le support pgcrypto.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Table des utilisateurs (base d'identité). Le schéma complet du modèle de
-- données est documenté dans docs/architecture.md et migré dans les étapes
-- suivantes.
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL,
    password_hash TEXT,
    full_name     TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Unicité sur l'email uniquement pour les utilisateurs actifs (soft delete).
CREATE UNIQUE INDEX uq_users_email_active ON users (email) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_users_email_active;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS pgcrypto;
