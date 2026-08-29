-- name: CreateIdea :one
INSERT INTO product_ideas (user_id, opportunity_id, title, subtitle, audience, problem, promise, format, estimated_price, difficulty, market_evidence, why_now, competitive_angle)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: ListIdeasByUser :many
SELECT * FROM product_ideas WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListIdeasByOpportunity :many
SELECT * FROM product_ideas WHERE user_id = $1 AND opportunity_id = $2 ORDER BY created_at ASC;

-- name: GetIdea :one
SELECT * FROM product_ideas WHERE id = $1 AND user_id = $2;

-- name: SelectIdea :one
UPDATE product_ideas SET is_selected = true WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: UnselectIdea :exec
UPDATE product_ideas SET is_selected = false WHERE id = $1 AND user_id = $2;
