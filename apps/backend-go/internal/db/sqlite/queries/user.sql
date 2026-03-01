-- name: GetUserByEmail :one
SELECT id, email, source, created_at, avatar_url
FROM user
WHERE email = ?;

-- name: GetUserById :one
SELECT id, email, source, created_at, avatar_url
FROM user
WHERE id = ?;

-- name: CreateUser :one
INSERT INTO user (id, email, source, avatar_url)
VALUES (?, ?, ?, ?)
RETURNING id, email, source, created_at, avatar_url;