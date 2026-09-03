-- Suivi global superadmin.

-- name: ListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users WHERE deleted_at IS NULL;

-- name: SetUserRole :one
UPDATE users SET role = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: CountProjects :one
SELECT count(*) FROM projects;

-- name: CountAssets :one
SELECT count(*) FROM assets;

-- name: CountConversations :one
SELECT count(*) FROM conversations;

-- name: CountJobsByStatus :many
SELECT status, count(*) AS total FROM generation_jobs GROUP BY status;

-- name: SumCreditsConsumed :one
SELECT COALESCE(sum(credits_consumed), 0)::bigint AS total FROM projects;

-- name: CountOpenTickets :one
SELECT count(*) FROM support_tickets WHERE status = 'open';
