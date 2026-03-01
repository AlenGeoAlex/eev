-- name: ListShareableOfUser :many
SELECT * FROM shareable WHERE user_id = ?;

-- name: GetShareable :many
SELECT s.*, so.option_key, so.value, u.email
FROM shareable s
JOIN shareable_options so ON s.id = so.share_id
JOIN user u ON u.id = s.user_id
WHERE s.id = ?;

-- name: InsertShareable :exec
INSERT INTO shareable (id, name, user_id, source_ip, expiry_at, shareable_type, shareable_data, active_from)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertShareableOption :exec
INSERT INTO shareable_options (share_id, option_key, value)
VALUES (?, ?, ?);

-- name: UpsertShareableOption :exec
INSERT INTO shareable_options (share_id, option_key, value)
VALUES (?, ?, ?)
ON CONFLICT(share_id, option_key) DO UPDATE SET value = excluded.value;

-- name: UpdateShareableOption :exec
UPDATE shareable_options
SET value = ?
WHERE share_id = ? AND option_key = ?;

-- name: InsertShareableFile :exec
INSERT INTO shareable_files (id, share_id, file_name, content_type, s3_key)
VALUES (?, ?, ?, ?, ?);

-- name: GetShareableFiles :many
SELECT id, share_id, file_name, content_type, s3_key, created_at
FROM shareable_files
WHERE share_id = ?;

-- name: DeleteShareableFile :exec
DELETE FROM shareable_files WHERE id = ?;

-- name: DeleteShareableFilesByShareID :exec
DELETE FROM shareable_files WHERE share_id = ?;