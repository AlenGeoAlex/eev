-- name: GetTargetEmailsForUser :many
SELECT target_email, starred
FROM user_target_history ut
WHERE ut.user_id = ?
  AND ut.target_email LIKE '%' || ? || '%'
ORDER BY ut.starred DESC, ut.occurences DESC;

-- name: UpsertTargetsForUser :exec
INSERT INTO user_target_history (user_id, target_email)
VALUES (?, ?)
ON CONFLICT(user_id, target_email)
    DO UPDATE SET occurences = occurences + 1;

-- name: UpdateStarForTargetWithUser :exec
UPDATE user_target_history
SET starred = ?
WHERE user_id = ?