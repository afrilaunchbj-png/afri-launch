-- +goose Up
-- Identité visuelle et configuration de génération par projet.
-- config = { palette: {primary, secondary, accent, background, text, source},
--            style, ebook_min_pages, ebook_max_pages }
-- Le workflow cover-first : la palette est proposée par l'IA (source "ai")
-- puis ajustable par l'utilisateur (source "user") ; l'ebook embarque la
-- cover validée en première page.

ALTER TABLE projects ADD COLUMN config JSONB NOT NULL DEFAULT '{}';

-- Paramètres d'un job (ex. instructions de régénération de cover).
ALTER TABLE generation_jobs ADD COLUMN params JSONB NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE generation_jobs DROP COLUMN IF EXISTS params;
ALTER TABLE projects DROP COLUMN IF EXISTS config;
