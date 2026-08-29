-- name: GetCreditAccountByUserID :one
SELECT * FROM credit_accounts WHERE user_id = $1;

-- name: GetCreditAccountByID :one
SELECT * FROM credit_accounts WHERE id = $1;

-- name: GetCreditAccountForUpdate :one
SELECT * FROM credit_accounts WHERE user_id = $1 FOR UPDATE;

-- name: CreateCreditAccount :one
INSERT INTO credit_accounts (user_id, balance)
VALUES ($1, $2)
RETURNING *;

-- name: AddCredits :one
UPDATE credit_accounts
SET balance = balance + $2, version = version + 1, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SubtractCredits :one
UPDATE credit_accounts
SET balance = balance - $2, version = version + 1, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: AddReserved :one
UPDATE credit_accounts
SET reserved = reserved + $2, version = version + 1, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SubtractReserved :one
UPDATE credit_accounts
SET reserved = reserved - $2, version = version + 1, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateCreditTransaction :one
INSERT INTO credit_transactions (account_id, type, amount, operation, reference, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCreditTransactionByReference :one
SELECT * FROM credit_transactions WHERE reference = $1;

-- name: ListCreditTransactions :many
SELECT * FROM credit_transactions
WHERE account_id = sqlc.arg('account_id')
  AND (sqlc.arg('type_filter')::text = '' OR type = sqlc.arg('type_filter')::text)
  AND (sqlc.arg('operation_filter')::text = '' OR operation = sqlc.arg('operation_filter')::text)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountCreditTransactions :one
SELECT count(*) FROM credit_transactions
WHERE account_id = sqlc.arg('account_id')
  AND (sqlc.arg('type_filter')::text = '' OR type = sqlc.arg('type_filter')::text)
  AND (sqlc.arg('operation_filter')::text = '' OR operation = sqlc.arg('operation_filter')::text);

-- name: SumCreditTransactionsByTypeSince :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total
FROM credit_transactions
WHERE account_id = $1 AND type = $2 AND status = 'completed' AND created_at >= $3;

-- name: GetCreditReservationByReference :one
SELECT * FROM credit_reservations WHERE reference = $1;

-- name: CreateCreditReservation :one
INSERT INTO credit_reservations (account_id, amount, operation, reference, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCreditReservationStatus :one
UPDATE credit_reservations
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListGenerationCosts :many
SELECT * FROM generation_costs WHERE is_active = true ORDER BY operation;

-- name: GetGenerationCost :one
SELECT * FROM generation_costs WHERE operation = $1;
