-- name: CreateJob :one
INSERT INTO jobs (image_id, type, payload_json)
VALUES ($1, $2, $3)
RETURNING id, image_id, type, payload_json, status, error, created_at, started_at, finished_at;

-- name: GetJobByID :one
SELECT id, image_id, type, payload_json, status, error, created_at, started_at, finished_at
FROM jobs
WHERE id = $1;

-- name: GetJobsByImageID :many
SELECT id, image_id, type, payload_json, status, error, created_at, started_at, finished_at
FROM jobs
WHERE image_id = $1
ORDER BY created_at DESC;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, image_id, type, payload_json, status, error, created_at, started_at, finished_at;

-- name: StartJob :one
UPDATE jobs
SET status = 'processing', started_at = now()
WHERE id = $1
RETURNING id, image_id, type, payload_json, status, error, created_at, started_at, finished_at;

-- name: CompleteJob :one
UPDATE jobs
SET status = 'completed', finished_at = now()
WHERE id = $1
RETURNING id, image_id, type, payload_json, status, error, created_at, started_at, finished_at;

-- name: FailJob :one
UPDATE jobs
SET status = 'failed', error = $2, finished_at = now()
WHERE id = $1
RETURNING id, image_id, type, payload_json, status, error, created_at, started_at, finished_at;

-- name: GetPendingJobs :many
SELECT id, image_id, type, payload_json, status, error, created_at, started_at, finished_at
FROM jobs
WHERE status = 'queued'
ORDER BY created_at ASC
LIMIT $1;

-- name: DeleteJob :exec
DELETE FROM jobs
WHERE id = $1;

-- name: DeleteJobsByImageID :exec
DELETE FROM jobs
WHERE image_id = $1;
