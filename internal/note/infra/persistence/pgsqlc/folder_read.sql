-- TODO: Check not use params
-- name: ReadGetFolder :one
SELECT
  *
FROM
  folders
WHERE
  id = sqlc.arg('id')
  AND trashed_by IS NULL -- :if @only_non_trashed
;

-- name: ReadGetFolderByID :one
SELECT
  *
FROM
  folders
WHERE
  id = sqlc.arg('id')
  AND trashed_at IS NULL;

-- name: ReadGetTrashedFolderByWorkspaceID :many
SELECT
  *
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND trashed_at IS NOT NULL;

-- name: ReadGetRootFolderIDsByWorkspaceID :many
SELECT
  id
FROM
  folders
WHERE
  workspace_id = sqlc.arg('workspace_id')
  AND parent_id IS NULL;

-- TODO: Should give sqlc dynamic a try, if it run nicely, then it would be more performant

-- name: ReadGetRecursiveFolderByParentID :many
WITH RECURSIVE subfolders AS (
  SELECT
    *,
    1 AS depth
  FROM
    folders
  WHERE
    parent_id = sqlc.arg('parent_id')::uuid
    AND CASE
      WHEN sqlc.arg('include_trashed')::bool = FALSE
      THEN trashed_at IS NULL
      ELSE TRUE
    END
  UNION ALL
  SELECT
    f.*,
    s.depth + 1 AS depth
  FROM
    folders AS f
    INNER JOIN subfolders s ON f.parent_id = s.id
  WHERE
    s.depth < COALESCE(sqlc.narg('depth')::int, 9999)
    AND CASE
      WHEN sqlc.arg('include_trashed')::bool = FALSE
      THEN f.trashed_at IS NULL
      ELSE TRUE
    END
)
SELECT
  *
FROM
  subfolders;
