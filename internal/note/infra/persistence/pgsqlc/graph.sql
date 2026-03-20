-- name: GetWorkspaceGraph :one
WITH valid_notes AS (
    SELECT
     n.id,
     n.name,
     n.tags
    FROM
      notes AS n
    INNER JOIN
      folders AS f
    ON n.folder_id = f.id
    WHERE
      f.workspace_id = sqlc.arg('workspace_id')
      AND n.trashed_at IS NULL
      AND f.trashed_at IS NULL
),
graph_nodes AS (
    SELECT
      id::text AS id,
      name,
      'note' AS type
    FROM valid_notes
    UNION
    SELECT
      DISTINCT
        '#' || tag AS id,
        tag AS name,
        'tag' AS type
    FROM
      valid_notes, unnest(tags) AS tag
),
graph_links AS (
    SELECT
      nl.source_id::text AS source,
      nl.target_id::text AS target
    FROM
      note_links AS nl
    WHERE
      nl.source_id IN (SELECT id FROM valid_notes)
      AND nl.target_id IN (SELECT id FROM valid_notes)
    UNION
    SELECT
      id::text AS source,
      '#' || tag AS target
    FROM
    valid_notes, unnest(tags) AS tag
)
SELECT
    COALESCE((SELECT json_agg(json_build_object('Id', id, 'Name', name, 'Type', type)) FROM graph_nodes), '[]'::json) AS nodes,
    COALESCE((SELECT json_agg(json_build_object('Source', source, 'Target', target)) FROM graph_links), '[]'::json) AS links;

-- name: GetNoteGraph :one
WITH RECURSIVE valid_notes AS (
    SELECT n.id, n.name, n.tags
    FROM notes n
    JOIN folders f ON n.folder_id = f.id
    WHERE f.workspace_id = sqlc.arg('workspace_id')
      AND n.trashed_at IS NULL
      AND f.trashed_at IS NULL
),
traversal_edges AS (
    -- Forward Note -> Note
    SELECT source_id::text AS source, target_id::text AS target FROM note_links
    WHERE source_id IN (SELECT id FROM valid_notes) AND target_id IN (SELECT id FROM valid_notes)
    UNION
    -- Backward Note -> Note (Allows finding backlinks)
    SELECT target_id::text AS source, source_id::text AS target FROM note_links
    WHERE source_id IN (SELECT id FROM valid_notes) AND target_id IN (SELECT id FROM valid_notes)
    UNION
    -- Note -> Tag (Prefix applied here for traversal)
    SELECT id::text AS source, '#' || tag AS target FROM valid_notes, unnest(tags) AS tag
    UNION
    -- Tag -> Note (Prefix applied here for traversal)
    SELECT '#' || tag AS source, id::text AS target FROM valid_notes, unnest(tags) AS tag
),
traverse (node_id, depth) AS (
    -- Base Case
    SELECT sqlc.arg('start_node_id')::text AS node_id, 0 AS depth
    UNION
    -- Recursive Step
    SELECT e.target AS node_id, t.depth + 1
    FROM traverse t
    JOIN traversal_edges e ON t.node_id = e.source
    WHERE t.depth < sqlc.arg('max_depth')::int
),
reachable_nodes AS (
    SELECT DISTINCT node_id FROM traverse
),
graph_nodes AS (
    -- Reconstruct Note Nodes
    SELECT n.id::text AS id, n.name, 'note' AS type
    FROM valid_notes n
    JOIN reachable_nodes r ON n.id::text = r.node_id
    UNION
    -- Reconstruct Tag Nodes (No need to append '#' again, r.node_id already has it)
    SELECT DISTINCT r.node_id AS id, r.node_id AS name, 'tag' AS type
    FROM reachable_nodes r
    WHERE r.node_id NOT IN (SELECT id::text FROM valid_notes)
),
graph_links AS (
    -- Filter Note -> Note links to only those we reached
    SELECT nl.source_id::text AS source, nl.target_id::text AS target
    FROM note_links nl
    WHERE nl.source_id::text IN (SELECT node_id FROM reachable_nodes)
      AND nl.target_id::text IN (SELECT node_id FROM reachable_nodes)
    UNION
    -- Filter Note -> Tag links to only those we reached (Prefix target with '#')
    SELECT id::text AS source, '#' || tag AS target
    FROM valid_notes, unnest(tags) AS tag
    WHERE id::text IN (SELECT node_id FROM reachable_nodes)
      AND ('#' || tag) IN (SELECT node_id FROM reachable_nodes)
)
SELECT
    COALESCE((SELECT json_agg(json_build_object('Id', id, 'Name', name, 'Type', type)) FROM graph_nodes), '[]'::json) AS nodes,
    COALESCE((SELECT json_agg(json_build_object('Source', source, 'Target', target)) FROM graph_links), '[]'::json) AS links;
