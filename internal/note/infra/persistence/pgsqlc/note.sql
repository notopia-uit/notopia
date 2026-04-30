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

-- name: GetNoteByID :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id')
FOR UPDATE -- :if @for_update
;

-- name: GetNotes :many
SELECT
  *
FROM
  notes
WHERE
  id = ANY(sqlc.narg('ids')::uuid[]) -- :if @ids
  AND folder_id IN ( -- :if @workspace_id
    SELECT id FROM folders WHERE workspace_id = sqlc.narg('workspace_id')::uuid
  )
  AND ( -- :if @trashed_by
    trashed_by = sqlc.narg('trashed_by')::text
    OR trashed_by IS NULL
  )
  AND trashed_by IS NOT NULL -- :if @trashed_only
FOR UPDATE -- :if @for_update
;

-- name: GetRecursiveChildrenFromFolder :many
WITH RECURSIVE subfolders AS (
  SELECT
    id
  FROM
    folders
  WHERE
    id = sqlc.arg('folder_id')::uuid
  UNION ALL
  SELECT
    f.id
  FROM
    folders f
  INNER JOIN
    subfolders sf
    ON f.parent_id = sf.id
)
SELECT
  n.*
FROM
  notes n
INNER JOIN
  subfolders sf
  ON n.folder_id = sf.id
FOR UPDATE -- :if @for_update
;


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

-- name: GetWorkspaceIDsByNoteIDs :many
SELECT
  n.id,
  f.workspace_id
FROM
  notes AS n
INNER JOIN
  folders f
  ON n.folder_id = f.id
WHERE
  n.id = ANY(sqlc.arg('ids')::uuid[]);

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
