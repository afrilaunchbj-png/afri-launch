-- name: ListOpportunities :many
SELECT * FROM opportunities
WHERE (sqlc.arg('country')::text = '' OR country = sqlc.arg('country')::text)
  AND (sqlc.arg('sector')::text = '' OR sector = sqlc.arg('sector')::text)
  AND (sqlc.arg('difficulty')::text = '' OR difficulty = sqlc.arg('difficulty')::text)
  AND (sqlc.arg('query')::text = '' OR title ILIKE '%' || sqlc.arg('query')::text || '%')
ORDER BY score DESC, created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountOpportunities :one
SELECT count(*) FROM opportunities
WHERE (sqlc.arg('country')::text = '' OR country = sqlc.arg('country')::text)
  AND (sqlc.arg('sector')::text = '' OR sector = sqlc.arg('sector')::text)
  AND (sqlc.arg('difficulty')::text = '' OR difficulty = sqlc.arg('difficulty')::text)
  AND (sqlc.arg('query')::text = '' OR title ILIKE '%' || sqlc.arg('query')::text || '%');

-- name: GetOpportunity :one
SELECT * FROM opportunities WHERE id = $1;

-- name: ListSavedOpportunityIDs :many
SELECT opportunity_id FROM saved_opportunities WHERE user_id = $1;

-- name: SaveOpportunity :exec
INSERT INTO saved_opportunities (user_id, opportunity_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnsaveOpportunity :exec
DELETE FROM saved_opportunities WHERE user_id = $1 AND opportunity_id = $2;

-- name: ListDistinctCountries :many
SELECT DISTINCT country FROM opportunities WHERE country <> '' ORDER BY country;

-- name: ListDistinctSectors :many
SELECT DISTINCT sector FROM opportunities WHERE sector <> '' ORDER BY sector;
