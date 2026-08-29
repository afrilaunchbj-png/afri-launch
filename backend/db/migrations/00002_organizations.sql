-- +goose Up
-- Organisations (tenant optionnel, prévu pour le B2B2C futur). Au MVP,
-- l'isolation se fait par user_id ; organization_id est posé pour l'avenir.
CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Colonnes d'identité/auth supplémentaires pour les utilisateurs.
ALTER TABLE users
    ADD COLUMN organization_id    UUID REFERENCES organizations(id),
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN email_verified_at  TIMESTAMPTZ;

CREATE INDEX idx_users_organization_id ON users (organization_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_organization_id;
ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS organization_id;
DROP TABLE IF EXISTS organizations;
