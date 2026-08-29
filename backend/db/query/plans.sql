-- name: ListPlans :many
SELECT * FROM plans WHERE is_active = true ORDER BY sort_order;
