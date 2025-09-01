-- name: CreateProject :one
INSERT INTO projects (name, user_id)
VALUES ($1, $2)
RETURNING id, name, user_id, created_at;

-- name: GetProjectByID :one
SELECT id, name, user_id, created_at
FROM projects
WHERE id = $1;

-- name: GetProjectsByUserID :many
SELECT id, name, user_id, created_at
FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetAllProjects :many
SELECT id, name, user_id, created_at
FROM projects
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET name = $2
WHERE id = $1
RETURNING id, name, user_id, created_at;

-- name: UpdateProjectByUserID :one
UPDATE projects
SET name = $3
WHERE id = $1 AND user_id = $2
RETURNING id, name, user_id, created_at;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

-- name: DeleteProjectByUserID :exec
DELETE FROM projects
WHERE id = $1 AND user_id = $2;

-- name: CountProjectsByUserID :one
SELECT COUNT(*)
FROM projects
WHERE user_id = $1;
