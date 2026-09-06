-- +goose Up
-- Un même produit Chariow peut être associé à PLUSIEURS projets SaaS.
-- On garantit au plus un lien par (projet, connexion) au lieu de forcer un
-- produit unique par connexion.

ALTER TABLE commerce_product_links DROP CONSTRAINT IF EXISTS commerce_product_links_connection_id_external_product_id_key;
ALTER TABLE commerce_product_links ADD CONSTRAINT commerce_product_links_connection_project_key UNIQUE (connection_id, project_id);
CREATE INDEX idx_commerce_links_external ON commerce_product_links (connection_id, external_product_id);

-- +goose Down
DROP INDEX IF EXISTS idx_commerce_links_external;
ALTER TABLE commerce_product_links DROP CONSTRAINT IF EXISTS commerce_product_links_connection_project_key;
ALTER TABLE commerce_product_links ADD CONSTRAINT commerce_product_links_connection_id_external_product_id_key UNIQUE (connection_id, external_product_id);
