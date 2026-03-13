-- name: SaveNote :exec
INSERT INTO notes (
  id,
  name,
  icon,
  folder_id,
  tags,
  size,
  created_at,
  updated_at,
  trashed_by,
  trashed_at
)
VALUES (
  sqlc.arg('id'),
  sqlc.arg('name'),
  sqlc.arg('icon'),
  sqlc.arg('folder_id'),
  sqlc.arg('tags'),
  sqlc.arg('size'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at'),
  sqlc.arg('trashed_by'),
  sqlc.arg('trashed_at')
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  icon = EXCLUDED.icon,
  folder_id = EXCLUDED.folder_id,
  tags = EXCLUDED.tags,
  size = EXCLUDED.size,
  updated_at = EXCLUDED.updated_at,
  trashed_by = EXCLUDED.trashed_by,
  trashed_at = EXCLUDED.trashed_at;

-- name: GetNote :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id')
  AND trashed_at IS NULL;

-- name: GetNotesByFolderID :many
SELECT
  *
FROM
  notes
WHERE
  folder_id = sqlc.arg('folder_id')
  AND trashed_at IS NULL
ORDER BY
  created_at DESC;

-- name: GetTrashedNotesByWorkspaceID :many
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


-- name: PermanentlyDeleteNoteByID :exec
DELETE FROM
  notes
WHERE
  id = sqlc.arg('id');

-- name: PermanentlyDeleteNotesByIDs :exec
DELETE FROM
  notes
WHERE
  id = ANY(sqlc.arg('ids')::uuid[]);
