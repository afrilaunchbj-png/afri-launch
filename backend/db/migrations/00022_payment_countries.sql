-- +goose Up
-- Pays acceptés pour les paiements (PawaPay, XOF) — activables/désactivables
-- depuis l'administration, sans toucher aux variables d'environnement.
-- Seul le Bénin est activé par défaut (checkout validé) ; les autres pays
-- UEMOA/XOF sont disponibles mais désactivés jusqu'à confirmation PawaPay.

CREATE TABLE payment_countries (
    code       TEXT PRIMARY KEY,
    name       TEXT NOT NULL DEFAULT '',
    currency   TEXT NOT NULL DEFAULT 'XOF',
    enabled    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO payment_countries (code, name, currency, enabled) VALUES
    ('BJ', 'Bénin', 'XOF', true),
    ('BF', 'Burkina Faso', 'XOF', false),
    ('CI', 'Côte d''Ivoire', 'XOF', false),
    ('GW', 'Guinée-Bissau', 'XOF', false),
    ('ML', 'Mali', 'XOF', false),
    ('NE', 'Niger', 'XOF', false),
    ('SN', 'Sénégal', 'XOF', false),
    ('TG', 'Togo', 'XOF', false)
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS payment_countries;
