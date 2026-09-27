-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, name, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, name, role, email_verified_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, role, email_verified_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, role, email_verified_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, $2), updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, password_hash, name, role, email_verified_at, created_at, updated_at
FROM users
WHERE ($1::text = '' OR name ILIKE '%' || $1::text || '%' OR email ILIKE '%' || $1::text || '%')
  AND ($2::text = '' OR role = $2::text)
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE ($1::text = '' OR name ILIKE '%' || $1::text || '%' OR email ILIKE '%' || $1::text || '%')
  AND ($2::text = '' OR role = $2::text);

-- name: UpdateUserRole :one
UPDATE users
SET role = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, name, role, email_verified_at, created_at, updated_at;
