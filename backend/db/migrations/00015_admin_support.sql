-- +goose Up
-- Sections Paramètres & Support + suivi global superadmin.

-- Rôle utilisateur : 'user' par défaut, 'superadmin' pour le suivi global
-- (promotion via SUPERADMIN_EMAILS au login ou manuellement).
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- Demandes de support (parcours utilisateur) + modération superadmin.
CREATE TABLE support_tickets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject    TEXT NOT NULL,
    message    TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'open'
               CHECK (status IN ('open','resolved')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_tickets_user ON support_tickets (user_id, created_at DESC);
CREATE INDEX idx_support_tickets_status ON support_tickets (status, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS support_tickets;
ALTER TABLE users DROP COLUMN IF EXISTS role;
