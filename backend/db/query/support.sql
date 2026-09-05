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
WHERE (@status::text = '' OR t.status = @status::text)
  AND (@search::text = '' OR t.subject ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%')
ORDER BY t.created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAllTickets :one
SELECT count(*) FROM support_tickets t
JOIN users u ON u.id = t.user_id
WHERE (@status::text = '' OR t.status = @status::text)
  AND (@search::text = '' OR t.subject ILIKE '%' || @search || '%' OR u.email ILIKE '%' || @search || '%');

-- name: GetTicket :one
SELECT * FROM support_tickets WHERE id = $1;

-- name: GetTicketWithUser :one
SELECT t.*, u.email AS user_email, u.full_name AS user_name
FROM support_tickets t
JOIN users u ON u.id = t.user_id
WHERE t.id = $1;

-- name: ListTicketMessages :many
SELECT m.*, u.email AS author_email, u.full_name AS author_name
FROM support_ticket_messages m
JOIN users u ON u.id = m.author_id
WHERE m.ticket_id = $1
ORDER BY m.created_at ASC;

-- name: InsertTicketMessage :one
WITH new_message AS (
    INSERT INTO support_ticket_messages (ticket_id, author_id, content, is_admin)
    VALUES ($1, $2, $3, $4)
    RETURNING *
)
SELECT m.*, u.email AS author_email, u.full_name AS author_name
FROM new_message m
JOIN users u ON u.id = m.author_id;

-- name: SetTicketStatus :one
UPDATE support_tickets SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: CreateSupportAttachment :one
INSERT INTO support_attachments (user_id, filename, storage_key, content_type, size_bytes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: BindSupportAttachments :exec
UPDATE support_attachments
SET ticket_id = $3,
    message_id = NULLIF($4, '')::uuid
WHERE id = ANY($2::uuid[]) AND user_id = $1 AND ticket_id IS NULL AND message_id IS NULL;

-- name: GetSupportAttachment :one
SELECT * FROM support_attachments WHERE id = $1 AND user_id = $2;

-- name: ListSupportAttachmentsByTicket :many
SELECT * FROM support_attachments WHERE ticket_id = $1 ORDER BY created_at ASC;

-- name: ListSupportAttachmentsByMessages :many
SELECT * FROM support_attachments
WHERE message_id = ANY($1::uuid[])
ORDER BY created_at ASC;

-- name: GetSupportAttachmentByID :one
SELECT * FROM support_attachments WHERE id = $1;
