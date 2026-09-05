-- name: ListActivePlans :many
SELECT * FROM plans WHERE is_active = true ORDER BY sort_order ASC;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = $1;

-- name: CreatePayment :one
INSERT INTO payments (user_id, plan_id, amount_minor, currency, provider, status, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdatePaymentCheckout :exec
UPDATE payments
SET provider_reference = $2, checkout_url = $3, updated_at = now()
WHERE id = $1;

-- name: MarkPaymentStatus :one
UPDATE payments
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = $1 AND user_id = $2;

-- name: GetPaymentByProviderReference :one
SELECT * FROM payments WHERE provider_reference = $1 LIMIT 1;

-- name: ListPaymentsByUser :many
SELECT * FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;
