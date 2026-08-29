-- +goose Up
-- Recherche en ligne à la demande + conversation d'idées.

-- Demande de recherche (une niche, un secteur, plusieurs marchés).
CREATE TABLE research_requests (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query      TEXT NOT NULL,
    sector     TEXT NOT NULL DEFAULT '',
    markets    JSONB NOT NULL DEFAULT '[]'::jsonb,
    language   TEXT NOT NULL DEFAULT 'fr',
    status     TEXT NOT NULL DEFAULT 'pending'
               CHECK (status IN ('pending','processing','completed','failed')),
    error      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_research_requests_user ON research_requests (user_id, created_at DESC);

-- Les opportunités peuvent être propres à un utilisateur (générées par la recherche).
ALTER TABLE opportunities
    ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN research_id UUID REFERENCES research_requests(id) ON DELETE SET NULL;

CREATE INDEX idx_opportunities_user ON opportunities (user_id) WHERE user_id IS NOT NULL;

-- Idées : phrase d'accroche + explication + statut (draft → confirmed).
ALTER TABLE product_ideas
    ADD COLUMN hook TEXT NOT NULL DEFAULT '',
    ADD COLUMN explanation TEXT NOT NULL DEFAULT '',
    ADD COLUMN status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','confirmed')),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Lien job → demande de recherche / idée (jobs de recherche + révision d'idée).
ALTER TABLE generation_jobs
    ADD COLUMN research_id UUID REFERENCES research_requests(id) ON DELETE CASCADE,
    ADD COLUMN idea_id UUID REFERENCES product_ideas(id) ON DELETE CASCADE;

-- Historique de conversation d'une idée (aller-retours avec le LLM).
CREATE TABLE idea_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idea_id    UUID NOT NULL REFERENCES product_ideas(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL CHECK (role IN ('user','assistant')),
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_idea_messages_idea ON idea_messages (idea_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS idea_messages;
ALTER TABLE generation_jobs
    DROP COLUMN IF EXISTS idea_id,
    DROP COLUMN IF EXISTS research_id;
ALTER TABLE product_ideas
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS explanation,
    DROP COLUMN IF EXISTS hook;
ALTER TABLE opportunities
    DROP COLUMN IF EXISTS research_id,
    DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS research_requests;
