-- name: ReadGetWorkspaceByNoteID :one
SELECT
  w.*
FROM
  workspaces w
JOIN
  folders f ON f.workspace_id = w.id
JOIN
  notes n ON n.folder_id = f.id
WHERE
  n.id = sqlc.arg('note_id')::uuid
  AND w.deleted_at IS NULL
  AND f.trashed_at IS NULL
  AND n.trashed_at IS NULL;

-- name: ReadGetWorkspacesByIDs :many
SELECT
  *
FROM
  workspaces
WHERE
  id = ANY(sqlc.arg('ids')::uuid[])
  AND deleted_at IS NULL;

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
