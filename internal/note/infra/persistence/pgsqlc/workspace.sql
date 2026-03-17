-- name: GetWorkspaceBySlug :one
SELECT
  *
FROM
  workspaces
WHERE
  slug = sqlc.arg('slug')
  AND deleted_at IS NULL;

-- name: GetWorkspaceBySlugForUpdate :one
SELECT
  *
FROM
  workspaces
WHERE
  slug = sqlc.arg('slug')
  AND deleted_at IS NULL
FOR UPDATE;

-- name: GetWorkspaceIDBySlug :one
SELECT
  id
FROM
  workspaces
WHERE
  slug = sqlc.arg('slug')
  AND deleted_at IS NULL;

-- name: CheckSlugExists :one
SELECT EXISTS(
  SELECT
    1
  FROM
    workspaces
  WHERE
    slug = sqlc.arg('slug')
    AND deleted_at IS NULL
) AS exists;

-- name: SaveWorkspace :exec
INSERT INTO workspaces (
  id,
  slug,
  name,
  created_at,
  updated_at,
  deleted_at
)
VALUES (
  sqlc.arg('id'),
  sqlc.arg('slug'),
  sqlc.arg('name'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at'),
  sqlc.arg('deleted_at')
)
ON CONFLICT (id) DO UPDATE SET
  slug = EXCLUDED.slug,
  name = EXCLUDED.name,
  updated_at = EXCLUDED.updated_at,
  deleted_at = EXCLUDED.deleted_at;

-- name: GetWorkspaceByID :one
SELECT *
FROM workspaces
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL;
