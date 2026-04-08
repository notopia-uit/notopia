-- name: ReadGetNoteBacklinks :many
SELECT
  source_id
FROM
  note_links
WHERE
  target_id = sqlc.arg('target_id');

-- name: ReadGetNoteOutgoingLinks :many
SELECT
  target_id
FROM
  note_links
WHERE
  source_id = sqlc.arg('source_id')::uuid
;

-- name: ReadGetNoteLinksInWorkspace :many
SELECT
    nl.*
FROM note_links AS nl
JOIN notes AS sn ON nl.source_id = sn.id
JOIN folders AS sf ON sn.folder_id = sf.id
WHERE sf.workspace_id = sqlc.arg('workspace_id')::uuid
  AND sn.trashed_at IS NULL
  AND sf.trashed_at IS NULL;
