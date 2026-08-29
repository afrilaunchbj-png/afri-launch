-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: UpsertUser :one
INSERT INTO users (id, email, full_name, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    email      = EXCLUDED.email,
    full_name  = EXCLUDED.full_name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = now()
RETURNING *;
