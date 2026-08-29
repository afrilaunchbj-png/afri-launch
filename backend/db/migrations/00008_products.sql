-- +goose Up
-- Parcours MVP : idées de produit, projets, assets générés et jobs de génération.

CREATE TABLE product_ideas (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id   UUID REFERENCES opportunities(id) ON DELETE SET NULL,
    title            TEXT NOT NULL,
    subtitle         TEXT NOT NULL DEFAULT '',
    audience         TEXT NOT NULL DEFAULT '',
    problem          TEXT NOT NULL DEFAULT '',
    promise          TEXT NOT NULL DEFAULT '',
    format           TEXT NOT NULL DEFAULT '',
    estimated_price  TEXT NOT NULL DEFAULT '',
    difficulty       TEXT NOT NULL DEFAULT '',
    market_evidence  TEXT NOT NULL DEFAULT '',
    why_now          TEXT NOT NULL DEFAULT '',
    competitive_angle TEXT NOT NULL DEFAULT '',
    is_selected      BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_ideas_user ON product_ideas (user_id, created_at DESC);
CREATE INDEX idx_product_ideas_opportunity ON product_ideas (opportunity_id) WHERE opportunity_id IS NOT NULL;

CREATE TABLE projects (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id   UUID REFERENCES opportunities(id) ON DELETE SET NULL,
    idea_id          UUID REFERENCES product_ideas(id) ON DELETE SET NULL,
    title            TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft','idea_selected','generating','content_ready','completed','failed')),
    credits_consumed INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_user ON projects (user_id, created_at DESC);

CREATE TABLE assets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    kind         TEXT NOT NULL,
    storage_key  TEXT NOT NULL,
    filename     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assets_project ON assets (project_id, created_at);

CREATE TABLE generation_jobs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id     UUID REFERENCES projects(id) ON DELETE CASCADE,
    opportunity_id UUID REFERENCES opportunities(id) ON DELETE CASCADE,
    kind           TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','processing','completed','failed')),
    error          TEXT,
    cost           INTEGER NOT NULL DEFAULT 0,
    result         JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at   TIMESTAMPTZ
);

CREATE INDEX idx_generation_jobs_user ON generation_jobs (user_id, created_at DESC);
CREATE INDEX idx_generation_jobs_project ON generation_jobs (project_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS generation_jobs;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS product_ideas;
