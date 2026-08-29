-- name: CreateJob :one
INSERT INTO generation_jobs (user_id, project_id, opportunity_id, kind, status, cost)
VALUES ($1, $2, $3, $4, 'pending', $5)
RETURNING *;

-- name: GetJob :one
SELECT * FROM generation_jobs WHERE id = $1 AND user_id = $2;

-- name: UpdateJobStatus :one
UPDATE generation_jobs SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: CompleteJob :one
UPDATE generation_jobs
SET status = 'completed', result = $2, cost = $3, completed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: FailJob :one
UPDATE generation_jobs
SET status = 'failed', error = $2, updated_at = now(), completed_at = now()
WHERE id = $1
RETURNING *;

-- name: ListJobsByProject :many
SELECT * FROM generation_jobs WHERE project_id = $1 ORDER BY created_at DESC;
