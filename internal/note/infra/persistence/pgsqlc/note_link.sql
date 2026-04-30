-- name: InsertTempNoteLinks :copyfrom
INSERT INTO temp_note_links (
  source_id,
  target_id
) VALUES (
  @source_id,
  @target_id
);

-- name: DeleteObsoleteNoteLinks :exec
DELETE FROM note_links AS nl
WHERE
  EXISTS (
    SELECT 1
    FROM temp_note_links AS tnl
    WHERE tnl.source_id = nl.source_id
  )
  AND NOT EXISTS (
    SELECT 1
    FROM temp_note_links AS tnl
    WHERE tnl.source_id = nl.source_id
    AND tnl.target_id = nl.target_id
  );

-- name: SaveFromTempNoteLinks :exec
INSERT INTO note_links (
  source_id,
  target_id
)
SELECT
  source_id,
  target_id
FROM
  temp_note_links
ON CONFLICT DO NOTHING;

-- name: GetNoteOutgoingLinks :many
SELECT
  target_id
FROM
  note_links
WHERE
  source_id = sqlc.arg('source_id')::uuid;

-- name: GetNotesOutgoingLinks :many
SELECT
  source_id,
  ARRAY_AGG(target_id)::uuid[] AS target_ids
FROM
  note_links
WHERE
  source_id = ANY(sqlc.arg('source_ids')::uuid[])
GROUP BY
  source_id;
