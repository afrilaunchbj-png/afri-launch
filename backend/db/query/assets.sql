-- name: CreateAsset :one
INSERT INTO assets (user_id, project_id, kind, storage_key, filename, content_type, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAsset :one
SELECT * FROM assets WHERE id = $1 AND user_id = $2;

-- name: ListAssetsByProject :many
SELECT * FROM assets WHERE project_id = $1 ORDER BY created_at ASC;
