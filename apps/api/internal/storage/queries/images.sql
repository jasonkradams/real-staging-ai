-- name: CreateImage :one
INSERT INTO images (project_id, original_url, room_type, style, seed)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at;

-- name: GetImageByID :one
SELECT id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at
FROM images
WHERE id = $1;

-- name: GetImagesByProjectID :many
SELECT id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at
FROM images
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: UpdateImageStatus :one
UPDATE images
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at;

-- name: UpdateImageWithStagedURL :one
UPDATE images
SET staged_url = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at;

-- name: UpdateImageWithError :one
UPDATE images
SET status = 'error', error = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, status, error, created_at, updated_at;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;

-- name: DeleteImagesByProjectID :exec
DELETE FROM images
WHERE project_id = $1;
