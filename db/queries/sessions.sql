-- name: CreateSession :exec
INSERT INTO auth_sessions (
    id, user_id, token_hash, user_agent, ip_address, expires_at, last_used_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CreateSessionToken :exec
INSERT INTO auth_session_tokens (id, session_id, token_hash)
VALUES ($1, $2, $3);

-- name: GetSessionToken :one
SELECT
    s.id,
    s.user_id,
    s.token_hash,
    s.user_agent,
    s.ip_address,
    s.expires_at,
    s.last_used_at,
    s.created_at,
    s.revoked_at,
    t.token_hash AS presented_token_hash,
    t.used_at AS token_used_at,
    t.revoked_at AS token_revoked_at
FROM auth_session_tokens t
JOIN auth_sessions s ON s.id = t.session_id
WHERE t.token_hash = $1
FOR UPDATE OF s, t;

-- name: MarkSessionTokenUsed :exec
UPDATE auth_session_tokens
SET used_at = $3
WHERE session_id = $1 AND token_hash = $2 AND used_at IS NULL;

-- name: UpdateSessionToken :exec
UPDATE auth_sessions
SET token_hash = $2, last_used_at = $3
WHERE id = $1;

-- name: RevokeSession :exec
UPDATE auth_sessions
SET revoked_at = COALESCE(revoked_at, NOW())
WHERE id = $1;

-- name: RevokeAllSessions :exec
UPDATE auth_sessions
SET revoked_at = COALESCE(revoked_at, NOW())
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: RevokeOtherSessions :exec
UPDATE auth_sessions
SET revoked_at = COALESCE(revoked_at, NOW())
WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL;

-- name: ListActiveSessions :many
SELECT id, user_id, user_agent, ip_address, expires_at, last_used_at, created_at, revoked_at
FROM auth_sessions
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
ORDER BY last_used_at DESC;

-- name: RevokeUserSession :execresult
UPDATE auth_sessions
SET revoked_at = COALESCE(revoked_at, NOW())
WHERE id = $1 AND user_id = $2;
