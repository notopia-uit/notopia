-- name: ReadGetNoteByID :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id');

-- name: ReadGetNotesByIDs :many
SELECT
  *
FROM
  notes
WHERE
  id = ANY(sqlc.arg('ids')::uuid[]);

-- TODO: Check if backlinks or outgoing link need those params
-- name: ReadGetNotes :many
SELECT
  *
FROM
  notes
WHERE
  id = ANY(sqlc.narg('ids')::uuid[]) -- :if @ids
  AND folder_id IN (
    SELECT id FROM folders WHERE workspace_id = sqlc.narg('workspace_id')::uuid
  ) -- :if @workspace_id
  AND ( -- :if @trashed_by
    trashed_by = sqlc.narg('trashed_by')::text
    OR trashed_by IS NULL
  )
  AND trashed_by IS NULL -- :if @non_trashed_only
  AND trashed_by IS NOT NULL -- :if @trashed_only
;

-- name: ReadGetNotesByFolderIDs :many
SELECT
  *
FROM
  notes
WHERE
  folder_id = ANY(sqlc.arg('folder_ids')::uuid[])
  AND trashed_at IS NULL -- :if @exclude_trash
;

-- name: ReadGetNotesInWorkspace :many
SELECT
  n.*
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  f.workspace_id = sqlc.arg('workspace_id')
  AND n.trashed_at IS NULL -- :if @exclude_trash
;

-- name: ReadGetTrashedNotesByWorkspaceID :many
SELECT
  n.*
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  f.workspace_id = sqlc.arg('workspace_id')
  AND n.trashed_at IS NOT NULL
ORDER BY
  n.trashed_at DESC;

-- name: ReadCountNoteBacklinks :one
SELECT
  COUNT(*)
FROM
  note_links
WHERE
  target_id = sqlc.arg('note_id');

-- name: ReadCountNoteOutgoingLinks :one
SELECT
  COUNT(*)
FROM
  note_links
WHERE
  source_id = sqlc.arg('note_id');
