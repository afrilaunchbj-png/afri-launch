-- name: InsertAuditLog :exec
INSERT INTO audit_logs (user_id, action, entity, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE (@user_id::uuid IS NULL OR user_id = @user_id::uuid)
  AND (@action::text = '' OR action = @action::text)
  AND (@entity::text = '' OR entity = @entity::text)
ORDER BY created_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: CountAuditLogs :one
SELECT count(*) FROM audit_logs
WHERE (@user_id::uuid IS NULL OR user_id = @user_id::uuid)
  AND (@action::text = '' OR action = @action::text)
  AND (@entity::text = '' OR entity = @entity::text);
