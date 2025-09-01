-- name: CreateUser :one
INSERT INTO users (auth0_sub, stripe_customer_id, role)
VALUES ($1, $2, $3)
RETURNING id, auth0_sub, stripe_customer_id, role, created_at;

-- name: GetUserByID :one
SELECT id, auth0_sub, stripe_customer_id, role, created_at
FROM users
WHERE id = $1;

-- name: GetUserByAuth0Sub :one
SELECT id, auth0_sub, stripe_customer_id, role, created_at
FROM users
WHERE auth0_sub = $1;

-- name: GetUserByStripeCustomerID :one
SELECT id, auth0_sub, stripe_customer_id, role, created_at
FROM users
WHERE stripe_customer_id = $1;

-- name: UpdateUserStripeCustomerID :one
UPDATE users
SET stripe_customer_id = $2
WHERE id = $1
RETURNING id, auth0_sub, stripe_customer_id, role, created_at;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2
WHERE id = $1
RETURNING id, auth0_sub, stripe_customer_id, role, created_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, auth0_sub, stripe_customer_id, role, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users;
