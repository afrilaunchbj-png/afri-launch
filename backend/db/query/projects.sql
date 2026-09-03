-- name: CreateProject :one
INSERT INTO projects (user_id, opportunity_id, idea_id, title, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1 AND user_id = $2;

-- name: ListProjectsByUser :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateProjectStatus :one
UPDATE projects SET status = $2, updated_at = now() WHERE id = $1 AND user_id = $3 RETURNING *;

-- name: AddProjectCredits :one
UPDATE projects SET credits_consumed = credits_consumed + $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: UpdateProjectConfig :one
UPDATE projects
SET config = $2, updated_at = now()
WHERE id = $1 AND user_id = $3
RETURNING *;
