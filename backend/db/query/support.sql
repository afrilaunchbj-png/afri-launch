-- name: CreateTicket :one
INSERT INTO support_tickets (user_id, subject, message)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListTicketsByUser :many
SELECT * FROM support_tickets WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListAllTickets :many
SELECT t.*, u.email AS user_email, u.full_name AS user_name
FROM support_tickets t
JOIN users u ON u.id = t.user_id
ORDER BY t.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllTickets :one
SELECT count(*) FROM support_tickets;

-- name: GetTicket :one
SELECT * FROM support_tickets WHERE id = $1;

-- name: SetTicketStatus :one
UPDATE support_tickets SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;
