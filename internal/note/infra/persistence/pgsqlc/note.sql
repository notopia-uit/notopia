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

-- name: InsertTempNotes :copyfrom
INSERT INTO temp_notes (
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
) VALUES (
  @id,
  @name,
  @icon,
  @folder_id,
  @tags,
  @size,
  @created_at,
  @updated_at,
  @trashed_by,
  @trashed_at
);

-- name: SaveFromTempNotes :exec
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
SELECT
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
FROM
  temp_notes
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

-- name: GetNoteForUpdate :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id')
  AND trashed_at IS NULL
FOR UPDATE;

-- name: GetNotes :many
SELECT
  *
FROM
  notes
WHERE
  id = ANY(sqlc.arg('ids')::uuid[])
  AND trashed_at IS NULL;

-- name: GetNotesForUpdate :many
SELECT
  *
FROM
  notes
WHERE
  id = ANY(sqlc.arg('ids')::uuid[])
  AND trashed_at IS NULL
FOR UPDATE;

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

-- name: GetWorkspaceIDByNoteID :one
SELECT
  f.workspace_id
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  n.id = sqlc.arg('id');

-- name: GetNotesInWorkspace :many
SELECT
  n.*
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  f.workspace_id = sqlc.arg('workspace_id')
  AND n.trashed_at IS NULL
  AND f.trashed_at IS NULL;

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

-- name: CountNotesInWorkspaceByIDs :one
SELECT
  COUNT(*)
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  f.workspace_id = sqlc.arg('workspace_id')
  AND n.id = ANY(sqlc.arg('ids')::uuid[]);

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
