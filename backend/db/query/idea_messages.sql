-- name: CreateIdeaMessage :one
INSERT INTO idea_messages (idea_id, user_id, role, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListIdeaMessages :many
SELECT * FROM idea_messages WHERE idea_id = $1 ORDER BY created_at ASC;
