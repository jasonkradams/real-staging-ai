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

-- name: GetUserProfileByID :one
SELECT 
  id, 
  auth0_sub, 
  stripe_customer_id, 
  role, 
  email,
  full_name,
  company_name,
  phone,
  billing_address,
  profile_photo_url,
  preferences,
  created_at,
  updated_at
FROM users
WHERE id = $1;

-- name: GetUserProfileByAuth0Sub :one
SELECT 
  id, 
  auth0_sub, 
  stripe_customer_id, 
  role, 
  email,
  full_name,
  company_name,
  phone,
  billing_address,
  profile_photo_url,
  preferences,
  created_at,
  updated_at
FROM users
WHERE auth0_sub = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET 
  email = COALESCE(sqlc.narg('email'), email),
  full_name = sqlc.narg('full_name'),
  company_name = sqlc.narg('company_name'),
  phone = sqlc.narg('phone'),
  billing_address = sqlc.narg('billing_address'),
  profile_photo_url = sqlc.narg('profile_photo_url'),
  preferences = COALESCE(sqlc.narg('preferences'), preferences)
WHERE id = $1
RETURNING 
  id, 
  auth0_sub, 
  stripe_customer_id, 
  role, 
  email,
  full_name,
  company_name,
  phone,
  billing_address,
  profile_photo_url,
  preferences,
  created_at,
  updated_at;
