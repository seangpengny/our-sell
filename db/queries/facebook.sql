-- name: UpsertFacebookConnection :one
INSERT INTO facebook_connections (
    id, user_id, facebook_user_id, facebook_user_name, encrypted_access_token,
    token_expires_at, scopes, revoked_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULL)
ON CONFLICT (user_id, facebook_user_id) DO UPDATE
SET facebook_user_name = EXCLUDED.facebook_user_name,
    encrypted_access_token = EXCLUDED.encrypted_access_token,
    token_expires_at = EXCLUDED.token_expires_at,
    scopes = EXCLUDED.scopes,
    revoked_at = NULL,
    updated_at = NOW()
RETURNING id, user_id, facebook_user_id, facebook_user_name, encrypted_access_token,
          token_expires_at, scopes, created_at, updated_at, revoked_at;

-- name: ListFacebookConnections :many
SELECT id, user_id, facebook_user_id, facebook_user_name, encrypted_access_token,
       token_expires_at, scopes, created_at, updated_at, revoked_at
FROM facebook_connections
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY updated_at DESC, id DESC;

-- name: GetFacebookConnection :one
SELECT id, user_id, facebook_user_id, facebook_user_name, encrypted_access_token,
       token_expires_at, scopes, created_at, updated_at, revoked_at
FROM facebook_connections
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL;

-- name: DeleteFacebookConnection :exec
DELETE FROM facebook_connections
WHERE id = $1 AND user_id = $2;
