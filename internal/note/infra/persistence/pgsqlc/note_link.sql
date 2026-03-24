-- name: CreateNoteLink :exec
INSERT INTO note_links (
  source_id,
  target_id
)
VALUES (
  sqlc.arg('source_id'),
  sqlc.arg('target_id')
)
ON CONFLICT DO NOTHING;

-- name: GetNoteOutgoingLinks :many
SELECT
  target_id
FROM
  note_links
WHERE
  source_id = sqlc.arg('source_id');

-- name: GetNotesOutgoingLinks :many
SELECT
  *
FROM
  note_links
WHERE
  source_id = ANY(sqlc.arg('source_ids')::uuid[]);

-- name: GetNoteBacklinks :many
SELECT
  source_id
FROM
  note_links
WHERE
  target_id = sqlc.arg('target_id');

-- name: GetNoteLinksInWorkspace :many
SELECT
    nl.*
FROM note_links AS nl
JOIN notes AS sn ON nl.source_id = sn.id
JOIN folders AS sf ON sn.folder_id = sf.id
WHERE sf.workspace_id = sqlc.arg('workspace_id')::uuid
  AND sn.trashed_at IS NULL
  AND sf.trashed_at IS NULL;

-- name: DeleteAnyNoteLinks :exec
DELETE FROM
  note_links
WHERE
  source_id = sqlc.arg('source_id')
  OR target_id = sqlc.arg('target_id');

-- name: DeleteNoteLink :exec
DELETE FROM
  note_links
WHERE
  source_id = sqlc.arg('source_id')
  AND target_id = sqlc.arg('target_id');
