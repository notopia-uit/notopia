-- name: GetNote :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id')
  AND deleted_at IS NULL;
