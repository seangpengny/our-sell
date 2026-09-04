-- name: InvalidateVerificationTokens :exec
UPDATE email_verification_tokens
SET used_at = COALESCE(used_at, NOW())
WHERE user_id = $1 AND used_at IS NULL;

-- name: CreateVerificationToken :exec
INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetVerificationToken :one
SELECT id, user_id, token_hash, expires_at, used_at, created_at
FROM email_verification_tokens
WHERE token_hash = $1
FOR UPDATE;

-- name: MarkVerificationTokenUsed :exec
UPDATE email_verification_tokens
SET used_at = $2
WHERE id = $1 AND used_at IS NULL;

-- name: InvalidatePasswordResetTokens :exec
UPDATE password_reset_tokens
SET used_at = COALESCE(used_at, NOW())
WHERE user_id = $1 AND used_at IS NULL;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetPasswordResetToken :one
SELECT id, user_id, token_hash, expires_at, used_at, created_at
FROM password_reset_tokens
WHERE token_hash = $1
FOR UPDATE;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = $2
WHERE id = $1 AND used_at IS NULL;
