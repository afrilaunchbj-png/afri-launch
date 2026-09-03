-- +goose Up
-- Préférences utilisateur (langue, thème) — chargées par le front après login
-- pour une expérience identique sur tous les appareils. La langue du compte
-- pilote aussi la langue des réponses du copilote.

CREATE TABLE user_preferences (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language   TEXT NOT NULL DEFAULT 'fr',
    theme      TEXT NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_preferences;
