-- name: GetNote :one
SELECT
  *
FROM
  notes
WHERE
  id = sqlc.arg('id')
  AND deleted_at IS NULL;

-- name: SaveNote :exec
INSERT INTO notes (id, name, icon, folder_id, tags, size, created_at, updated_at, deleted_by, deleted_at)
VALUES (sqlc.arg('id'), sqlc.arg('name'), sqlc.arg('icon'), sqlc.arg('folder_id'), sqlc.arg('tags'), sqlc.arg('size'), sqlc.arg('created_at'), sqlc.arg('updated_at'), sqlc.arg('deleted_by'), sqlc.arg('deleted_at'))
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  icon = EXCLUDED.icon,
  folder_id = EXCLUDED.folder_id,
  tags = EXCLUDED.tags,
  size = EXCLUDED.size,
  updated_at = EXCLUDED.updated_at,
  deleted_by = EXCLUDED.deleted_by,
  deleted_at = EXCLUDED.deleted_at;

-- name: GetTrashedNotesByWorkspaceID :many
SELECT n.*
FROM notes n
JOIN folders f ON n.folder_id = f.id
WHERE f.workspace_id = sqlc.arg('workspace_id')
  AND n.deleted_at IS NOT NULL
ORDER BY n.deleted_at DESC;

-- name: PermanentlyDeleteNoteByID :exec
DELETE FROM notes WHERE id = sqlc.arg('id');

-- name: PermanentlyDeleteNotesByIDs :exec
DELETE FROM notes WHERE id = ANY(sqlc.arg('ids')::uuid[]);

-- name: GetNotesByFolderID :many
SELECT *
FROM notes
WHERE folder_id = sqlc.arg('folder_id')
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: AddNoteLink :exec
INSERT INTO note_links (source_id, target_id)
VALUES (sqlc.arg('source_id'), sqlc.arg('target_id'))
ON CONFLICT DO NOTHING;

-- name: RemoveNoteLink :exec
DELETE FROM note_links WHERE source_id = sqlc.arg('source_id') AND target_id = sqlc.arg('target_id');

-- name: GetNoteOutgoingLinks :many
SELECT target_id
FROM note_links
WHERE source_id = sqlc.arg('source_id');

-- name: DeleteNoteOutgoingLinks :exec
DELETE FROM note_links WHERE source_id = sqlc.arg('source_id');
