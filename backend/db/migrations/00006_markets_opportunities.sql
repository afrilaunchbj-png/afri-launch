-- +goose Up
-- Marchés (référentiel géo-économique) et opportunités.
-- Au MVP, `opportunities` est un catalogue global de référence (user_id NULL),
-- consultable/filtrable par tous ; les sauvegardes sont par utilisateur via
-- `saved_opportunities`. Les opportunités propres à un utilisateur (générées
-- par la recherche IA) arriveront avec le worker de recherche.
CREATE TABLE markets (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code      TEXT NOT NULL,
    name      TEXT NOT NULL,
    currency  TEXT NOT NULL,
    language  TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX uq_markets_code ON markets (code);

CREATE TABLE opportunities (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT NOT NULL,
    summary    TEXT NOT NULL DEFAULT '',
    country    TEXT NOT NULL DEFAULT '',
    sector     TEXT NOT NULL DEFAULT '',
    language   TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL DEFAULT 'medium' CHECK (difficulty IN ('low', 'medium', 'high')),
    signal     TEXT NOT NULL DEFAULT 'hypothesis'
               CHECK (signal IN ('verified', 'estimated', 'inferred', 'hypothesis')),
    score      INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),
    scores     JSONB NOT NULL DEFAULT '{}'::jsonb,
    evidence   JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_opportunities_filters ON opportunities (country, sector, difficulty);
CREATE INDEX idx_opportunities_score ON opportunities (score DESC);

CREATE TABLE saved_opportunities (
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, opportunity_id)
);

CREATE INDEX idx_saved_opportunities_user ON saved_opportunities (user_id, created_at DESC);

-- Audit minimal (append-only) — journal des opérations sensibles.
CREATE TABLE audit_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID,
    action     TEXT NOT NULL,
    entity     TEXT NOT NULL,
    entity_id  TEXT,
    metadata   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user ON audit_logs (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS saved_opportunities;
DROP TABLE IF EXISTS opportunities;
DROP TABLE IF EXISTS markets;
