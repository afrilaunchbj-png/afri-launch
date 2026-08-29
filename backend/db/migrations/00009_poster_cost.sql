-- +goose Up
-- Coût de génération des affiches publicitaires (3 images).
INSERT INTO generation_costs (operation, name, credits) VALUES
    ('poster_generation', 'Génération d''affiches publicitaires (x3)', 9);

-- +goose Down
DELETE FROM generation_costs WHERE operation = 'poster_generation';
