-- name: GetUserPreferences :one
SELECT * FROM user_preferences WHERE user_id = $1;

-- name: InsertUserPreferences :one
INSERT INTO user_preferences (user_id, language, theme)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO NOTHING
RETURNING *;

-- name: UpdateUserPreferences :one
UPDATE user_preferences
SET language = $2, theme = $3, updated_at = now()
WHERE user_id = $1
RETURNING *;
