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
  CASE
    WHEN sqlc.narg('id')::uuid IS NOT NULL
    THEN id = sqlc.narg('id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('workspace_id')::uuid IS NOT NULL
    THEN workspace_id = sqlc.narg('workspace_id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('parent_id')::uuid IS NOT NULL
    THEN parent_id = sqlc.narg('parent_id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.arg('is_root_folder')::bool = TRUE
    THEN parent_id IS NULL
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.arg('trashed_by')::text <> ''
    THEN trashed_by = sqlc.arg('trashed_by')::text
    ELSE TRUE
  END
ORDER BY
  created_at DESC;

-- name: GetFolders :many
SELECT
  *
FROM
  folders
WHERE
  CASE
    WHEN sqlc.narg('ids')::uuid[] IS NOT NULL
    THEN id = ANY(sqlc.narg('ids')::uuid[])
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('workspace_id')::uuid IS NOT NULL
    THEN workspace_id = sqlc.narg('workspace_id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('parent_id')::uuid IS NOT NULL
    THEN parent_id = sqlc.narg('parent_id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.arg('is_root_folder')::bool = TRUE
    THEN parent_id IS NULL
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('trashed_by')::text IS NOT NULL
    THEN trashed_by = sqlc.narg('trashed_by')::text
    ELSE TRUE
  END
ORDER BY
  created_at DESC;

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
