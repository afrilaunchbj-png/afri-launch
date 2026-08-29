-- name: ListMarkets :many
SELECT * FROM markets WHERE is_active = true ORDER BY name;
