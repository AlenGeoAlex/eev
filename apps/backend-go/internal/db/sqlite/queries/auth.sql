-- name: InsertRefreshToken :exec
INSERT INTO refresh_token (jti, user_id, expires_at)
VALUES (?, ?, ?);

-- name: GetRefreshToken :one
SELECT jti, user_id, expires_at, revoked
FROM refresh_token
WHERE jti = ?
LIMIT 1;

-- name: RevokeRefreshToken :exec
DELETE FROM refresh_token
WHERE jti = ?;