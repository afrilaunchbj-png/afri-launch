-- name: UpsertCommerceConnection :exec
INSERT INTO commerce_connections
    (user_id, provider, status, api_key_enc, webhook_secret_enc,
     external_store_id, store_name, store_url, store_currency, connected_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (user_id, provider) DO UPDATE SET
    status = excluded.status,
    api_key_enc = excluded.api_key_enc,
    webhook_secret_enc = excluded.webhook_secret_enc,
    external_store_id = excluded.external_store_id,
    store_name = excluded.store_name,
    store_url = excluded.store_url,
    store_currency = excluded.store_currency,
    connected_at = excluded.connected_at,
    updated_at = now();

-- name: GetCommerceConnection :one
SELECT * FROM commerce_connections WHERE user_id = $1 AND provider = $2;

-- name: GetCommerceConnectionByStoreID :one
SELECT * FROM commerce_connections
WHERE external_store_id = $1 AND provider = 'chariow'
LIMIT 1;

-- name: ListCommerceConnections :many
SELECT * FROM commerce_connections WHERE user_id = $1 ORDER BY created_at ASC;

-- name: UpdateCommerceConnectionState :exec
UPDATE commerce_connections SET
    status = $3,
    last_error = $4,
    last_error_at = CASE WHEN $4 <> '' THEN now() ELSE last_error_at END,
    last_sync_at = $5,
    connected_at = $6,
    updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: DeleteCommerceConnection :exec
DELETE FROM commerce_connections WHERE id = $1 AND user_id = $2;

-- name: UpsertCommerceProductLink :one
INSERT INTO commerce_product_links
    (user_id, connection_id, project_id, external_product_id, external_product_slug,
     external_product_name, price_minor, currency, status, is_public, public_token)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (connection_id, external_product_id) DO UPDATE SET
    user_id = excluded.user_id,
    project_id = excluded.project_id,
    external_product_slug = excluded.external_product_slug,
    external_product_name = excluded.external_product_name,
    price_minor = excluded.price_minor,
    currency = excluded.currency,
    status = excluded.status,
    is_public = excluded.is_public,
    public_token = excluded.public_token,
    updated_at = now()
RETURNING *;

-- name: GetCommerceProductLinkByProject :one
SELECT * FROM commerce_product_links
WHERE user_id = $1 AND project_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: GetCommerceProductLinkByExternal :one
SELECT * FROM commerce_product_links
WHERE connection_id = $1 AND external_product_id = $2;

-- name: GetCommerceProductLink :one
SELECT * FROM commerce_product_links WHERE id = $1 AND user_id = $2;

-- name: GetCommerceProductLinkByToken :one
SELECT * FROM commerce_product_links WHERE public_token = $1 AND is_public = true;

-- name: UpdateCommerceProductLinkPublic :exec
UPDATE commerce_product_links SET
    is_public = $3,
    public_token = $4,
    updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: DeleteCommerceProductLink :exec
DELETE FROM commerce_product_links WHERE id = $1 AND user_id = $2;

-- name: DeleteCommerceProductLinksByProject :exec
DELETE FROM commerce_product_links WHERE user_id = $1 AND project_id = $2;

-- name: ListCommerceProductLinks :many
SELECT * FROM commerce_product_links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpsertCommerceSale :one
INSERT INTO commerce_sales
    (user_id, connection_id, product_link_id, external_sale_id, status,
     amount_minor, currency, buyer_email, buyer_name, checkout_url,
     custom_metadata, payload, completed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (connection_id, external_sale_id) DO UPDATE SET
    status = excluded.status,
    amount_minor = excluded.amount_minor,
    currency = excluded.currency,
    buyer_email = excluded.buyer_email,
    buyer_name = excluded.buyer_name,
    -- On préserve le checkout_url/metadata/payload existants si l'upsert
    -- (ex. sync) n'en apporte pas (ne jamais effacer l'URL de paiement).
    checkout_url = CASE WHEN excluded.checkout_url <> '' THEN excluded.checkout_url ELSE commerce_sales.checkout_url END,
    custom_metadata = CASE WHEN excluded.custom_metadata = '{}'::jsonb THEN commerce_sales.custom_metadata ELSE excluded.custom_metadata END,
    payload = CASE WHEN excluded.payload = '{}'::jsonb THEN commerce_sales.payload ELSE excluded.payload END,
    completed_at = excluded.completed_at,
    updated_at = now()
RETURNING *;

-- name: GetCommerceSaleByExternalID :one
SELECT * FROM commerce_sales
WHERE connection_id = $1 AND external_sale_id = $2;

-- name: ListCommerceSalesByProductLink :many
SELECT * FROM commerce_sales
WHERE product_link_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListCommerceSalesByUser :many
SELECT * FROM commerce_sales
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: InsertCommerceWebhookEvent :one
INSERT INTO commerce_webhook_events
    (user_id, provider, external_pulse_id, external_delivery_id, event_type, payload, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (provider, external_delivery_id) DO NOTHING
RETURNING *;
