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
    ('BEN', 'Bénin', 'XOF', true),
    ('BFA', 'Burkina Faso', 'XOF', false),
    ('CIV', 'Côte d''Ivoire', 'XOF', false),
    ('GNB', 'Guinée-Bissau', 'XOF', false),
    ('MLI', 'Mali', 'XOF', false),
    ('NER', 'Niger', 'XOF', false),
    ('SEN', 'Sénégal', 'XOF', false),
    ('TGO', 'Togo', 'XOF', false)
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS payment_countries;
