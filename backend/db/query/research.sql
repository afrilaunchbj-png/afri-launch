-- name: CreateResearchRequest :one
INSERT INTO research_requests (user_id, query, sector, markets, language)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetResearchRequest :one
SELECT * FROM research_requests WHERE id = $1 AND user_id = $2;

-- name: UpdateResearchRequestStatus :one
UPDATE research_requests SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;
