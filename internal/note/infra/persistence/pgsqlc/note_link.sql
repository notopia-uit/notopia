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
  CASE
    WHEN sqlc.narg('source_id')::uuid IS NOT NULL THEN source_id = sqlc.arg('source_id')::uuid
    ELSE TRUE
  END
  AND CASE
    WHEN sqlc.narg('source_ids')::uuid[] IS NOT NULL THEN source_id = ANY(sqlc.arg('source_ids')::uuid[])
    ELSE TRUE
  END;

-- name: GetNoteBacklinks :many
SELECT
  source_id
FROM
  note_links
WHERE
  target_id = sqlc.arg('target_id');

-- name: GetNoteLinksInWorkspace :many
SELECT
    nl.*
FROM note_links AS nl
JOIN notes AS sn ON nl.source_id = sn.id
JOIN folders AS sf ON sn.folder_id = sf.id
WHERE sf.workspace_id = sqlc.arg('workspace_id')::uuid
  AND sn.trashed_at IS NULL
  AND sf.trashed_at IS NULL;
