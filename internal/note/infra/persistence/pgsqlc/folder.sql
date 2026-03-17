-- name: GetFolder :one
SELECT
  *
FROM
  folders
WHERE
  id = sqlc.arg('id')
  AND trashed_at IS NULL;

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

-- name: GetFolders :many
SELECT
  *
FROM
  folders
WHERE
  id = ANY(sqlc.arg('ids')::uuid[])
  AND trashed_at IS NULL;

-- name: CountFoldersInWorkspaceByIDs :one
SELECT
  COUNT(*)
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND id = ANY(sqlc.arg('ids')::uuid[]);

-- name: GetTrashedFoldersByWorkspaceID :many
SELECT
  *
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND trashed_by = sqlc.arg('trashed_by')::string
ORDER BY
  trashed_at DESC;

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

-- name: GetFoldersByID :many
SELECT
  *
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND trashed_at IS NULL
ORDER BY
  created_at DESC;

-- name: GetRootFolder :one
SELECT
  *
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND parent_id IS NULL
  AND trashed_at IS NULL;

-- name: GetFoldersByParentID :many
SELECT
  *
FROM
  folders
WHERE
  parent_id = sqlc.arg('parent_id')
  AND trashed_at IS NULL
ORDER BY
  created_at DESC;
