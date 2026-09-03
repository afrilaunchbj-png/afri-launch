-- name: CreateConversation :one
INSERT INTO conversations (user_id)
VALUES ($1)
RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = $1 AND user_id = $2;

-- name: ListConversations :many
SELECT * FROM conversations WHERE user_id = $1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3;

-- name: TouchConversation :one
UPDATE conversations SET updated_at = now() WHERE id = $1 RETURNING *;

-- name: SetConversationOpportunity :one
UPDATE conversations SET opportunity_id = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: SetConversationTitle :one
UPDATE conversations SET title = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: CreateConversationMessage :one
INSERT INTO conversation_messages (id, conversation_id, user_id, role, content, payload)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListConversationMessages :many
SELECT * FROM conversation_messages WHERE conversation_id = $1 ORDER BY created_at ASC;

-- name: ListIdeasByConversation :many
SELECT * FROM product_ideas WHERE user_id = $1 AND conversation_id = $2 ORDER BY created_at ASC;
