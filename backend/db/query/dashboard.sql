-- name: CountUserProjects :one
SELECT count(*) FROM projects WHERE user_id = $1;

-- name: CountUserConversations :one
SELECT count(*) FROM conversations WHERE user_id = $1;

-- name: CountUserOpenTickets :one
SELECT count(*) FROM support_tickets WHERE user_id = $1 AND status = 'open';

-- name: CountUserJobsByStatus :many
SELECT status, count(*) AS total FROM generation_jobs WHERE user_id = $1 GROUP BY status;

-- name: GetUserCreditAccount :one
SELECT COALESCE(balance, 0)::bigint AS balance, COALESCE(reserved, 0)::bigint AS reserved
FROM credit_accounts WHERE user_id = $1;

-- name: SumUserCreditsUsedSince :one
SELECT COALESCE(sum(t.amount), 0)::bigint AS total
FROM credit_transactions t
JOIN credit_accounts a ON a.id = t.account_id
WHERE a.user_id = $1
  AND t.type = 'debit'
  AND t.status = 'completed'
  AND t.created_at >= $2;

-- name: UserCreditsPerDay :many
SELECT date_trunc('day', t.created_at)::timestamptz AS day, COALESCE(sum(t.amount), 0)::bigint AS total
FROM credit_transactions t
JOIN credit_accounts a ON a.id = t.account_id
WHERE a.user_id = $1
  AND t.type = 'debit'
  AND t.status = 'completed'
  AND t.created_at >= $2
GROUP BY day
ORDER BY day;

-- name: UserProjectsPerWeek :many
SELECT date_trunc('week', created_at)::timestamptz AS week, count(*)::bigint AS total
FROM projects
WHERE user_id = $1
  AND created_at >= $2
GROUP BY week
ORDER BY week;
