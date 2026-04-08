-- name: ReadGetWorkspaceBySlug :one
SELECT
  *
FROM
  workspaces
WHERE
  slug = sqlc.arg('slug')::text
  AND deleted_at IS NULL;

-- name: ReadCheckSlugExists :one
SELECT EXISTS(
  SELECT
    1
  FROM
    workspaces
  WHERE
    slug = sqlc.arg('slug')::text
    AND deleted_at IS NULL
) AS exists;
