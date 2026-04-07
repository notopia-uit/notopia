-- name: SaveFolder :exec
INSERT INTO folders (
  id,
  name,
  icon,
  workspace_id,
  parent_id,
  created_at,
  updated_at,
  trashed_by,
  trashed_at
)
VALUES (
  sqlc.arg('id'),
  sqlc.arg('name'),
  sqlc.arg('icon'),
  sqlc.arg('workspace_id'),
  sqlc.arg('parent_id'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at'),
  sqlc.arg('trashed_by'),
  sqlc.arg('trashed_at')
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  icon = EXCLUDED.icon,
  parent_id = EXCLUDED.parent_id,
  updated_at = EXCLUDED.updated_at,
  trashed_by = EXCLUDED.trashed_by,
  trashed_at = EXCLUDED.trashed_at;

-- name: InsertTempFolders :copyfrom
INSERT INTO temp_folders (
  id,
  name,
  icon,
  workspace_id,
  parent_id,
  created_at,
  updated_at,
  trashed_by,
  trashed_at
) VALUES (
  @id,
  @name,
  @icon,
  @workspace_id,
  @parent_id,
  @created_at,
  @updated_at,
  @trashed_by,
  @trashed_at
);

-- name: SaveFromTempFolders :exec
INSERT INTO folders (
  id,
  name,
  icon,
  workspace_id,
  parent_id,
  created_at,
  updated_at,
  trashed_by,
  trashed_at
)
SELECT
  id,
  name,
  icon,
  workspace_id,
  parent_id,
  created_at,
  updated_at,
  trashed_by,
  trashed_at
FROM
  temp_folders
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  icon = EXCLUDED.icon,
  parent_id = EXCLUDED.parent_id,
  updated_at = EXCLUDED.updated_at,
  trashed_by = EXCLUDED.trashed_by,
  trashed_at = EXCLUDED.trashed_at;

-- name: GetFolder :one
SELECT
  *
FROM
  folders
WHERE
  id = sqlc.arg('id')
FOR UPDATE -- :if @for_update
;

-- name: GetFoldersByWorkspaceID :many
SELECT
  *
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
ORDER BY
  created_at DESC
FOR UPDATE -- :if @for_update
;

-- name: GetFolders :many
SELECT
  *
FROM
  folders
WHERE
  id = ANY(sqlc.narg('ids')::uuid[]) -- :if @ids
  AND workspace_id = sqlc.narg('workspace_id')::uuid -- :if @workspace_id
  AND ( -- :if @trashed_by
    trashed_by = sqlc.narg('trashed_by')::text
    OR trashed_by IS NULL
  )
  AND trashed_by IS NOT NULL -- :if @trashed_only
ORDER BY
  created_at DESC
FOR UPDATE -- :if @for_update
;

-- name: GetWorkspaceIDByFolderID :one
SELECT
  workspace_id
FROM
  folders
WHERE
  id = sqlc.arg('id');

-- name: CountFoldersInWorkspaceByIDs :one
SELECT
  COUNT(*)
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND id = ANY(sqlc.arg('ids')::uuid[]);

-- name: PermanentlyDeleteFolderByID :exec
DELETE FROM
  folders
WHERE
  id = sqlc.arg('id');

-- name: PermanentlyDeleteFoldersByIDs :exec
DELETE FROM
  folders
WHERE
  id = ANY(sqlc.arg('ids')::uuid[]);
