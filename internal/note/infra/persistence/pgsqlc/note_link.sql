-- name: GetNoteIncomingLinks :many
SELECT source_id
FROM note_links
WHERE target_id = sqlc.arg('target_id');

-- name: DeleteAllNoteLinks :exec
DELETE FROM note_links WHERE source_id = sqlc.arg('source_id') OR target_id = sqlc.arg('target_id');
