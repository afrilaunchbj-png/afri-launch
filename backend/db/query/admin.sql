-- Suivi global superadmin.

-- name: ListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users WHERE deleted_at IS NULL;

-- name: ListUsersFiltered :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (@search::text = '' OR email ILIKE '%' || @search || '%' OR full_name ILIKE '%' || @search || '%')
  AND (@role::text = '' OR role = @role::text)
ORDER BY created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountUsersFiltered :one
SELECT count(*) FROM users
WHERE deleted_at IS NULL
  AND (@search::text = '' OR email ILIKE '%' || @search || '%' OR full_name ILIKE '%' || @search || '%')
  AND (@role::text = '' OR role = @role::text);

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

-- name: ListAllProjects :many
SELECT p.*, u.email AS user_email, u.full_name AS user_name
FROM projects p
JOIN users u ON u.id = p.user_id
WHERE (@status::text = '' OR p.status = @status::text)
  AND (@search::text = '' OR p.title ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%')
ORDER BY p.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllProjects :one
SELECT count(*) FROM projects p
JOIN users u ON u.id = p.user_id
WHERE (@status::text = '' OR p.status = @status::text)
  AND (@search::text = '' OR p.title ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%');

-- name: ListAllConversations :many
SELECT c.*, u.email AS user_email, u.full_name AS user_name
FROM conversations c
JOIN users u ON u.id = c.user_id
WHERE (@search::text = '' OR c.title ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%')
ORDER BY c.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllConversations :one
SELECT count(*) FROM conversations c
JOIN users u ON u.id = c.user_id
WHERE (@search::text = '' OR c.title ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%');

-- name: ListAllAssets :many
SELECT a.*, p.title AS project_title, u.email AS user_email
FROM assets a
JOIN projects p ON p.id = a.project_id
JOIN users u ON u.id = a.user_id
WHERE (@search::text = '' OR a.filename ILIKE '%' || @search || '%' OR p.title ILIKE '%' || @search || '%')
ORDER BY a.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllAssets :one
SELECT count(*) FROM assets a
JOIN projects p ON p.id = a.project_id
JOIN users u ON u.id = a.user_id
WHERE (@search::text = '' OR a.filename ILIKE '%' || @search || '%' OR p.title ILIKE '%' || @search || '%');

-- name: ListAllJobs :many
SELECT j.*, u.email AS user_email, u.full_name AS user_name
FROM generation_jobs j
JOIN users u ON u.id = j.user_id
WHERE (@status::text = '' OR j.status = @status::text)
  AND (@search::text = '' OR j.kind ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%')
ORDER BY j.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllJobs :one
SELECT count(*) FROM generation_jobs j
JOIN users u ON u.id = j.user_id
WHERE (@status::text = '' OR j.status = @status::text)
  AND (@search::text = '' OR j.kind ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%');

-- name: ListAllCreditTransactions :many
SELECT t.*, u.email AS user_email
FROM credit_transactions t
JOIN credit_accounts a ON a.id = t.account_id
JOIN users u ON u.id = a.user_id
WHERE (@type::text = '' OR t.type = @type::text)
  AND (@search::text = '' OR t.operation ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%')
ORDER BY t.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllCreditTransactions :one
SELECT count(*) FROM credit_transactions t
JOIN credit_accounts a ON a.id = t.account_id
JOIN users u ON u.id = a.user_id
WHERE (@type::text = '' OR t.type = @type::text)
  AND (@search::text = '' OR t.operation ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%');
