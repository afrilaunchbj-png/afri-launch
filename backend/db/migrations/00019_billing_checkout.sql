-- +goose Up
-- Checkout paiements (ADR-018) : URL de checkout persistée + packs de
-- crédits par défaut (seed idempotent, prix en XOF unités mineures).

ALTER TABLE payments ADD COLUMN IF NOT EXISTS checkout_url TEXT;

INSERT INTO plans (id, name, credits, price_minor, currency, sort_order)
VALUES
    ('11111111-1111-4111-8111-111111111111', 'Pack Découverte', 50,  5000,  'XOF', 1),
    ('22222222-2222-4222-8222-222222222222', 'Pack Business',   120, 10000, 'XOF', 2),
    ('33333333-3333-4333-8333-333333333333', 'Pack Pro',        350, 25000, 'XOF', 3)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM plans WHERE id IN (
    '11111111-1111-4111-8111-111111111111',
    '22222222-2222-4222-8222-222222222222',
    '33333333-3333-4333-8333-333333333333'
);
ALTER TABLE payments DROP COLUMN IF EXISTS checkout_url;
