-- name: CreateTempTableNotes :exec
CREATE TEMP TABLE temp_notes (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  folder_id UUID NOT NULL,
  tags TEXT[],
  size INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  trashed_by trashed_by,
  trashed_at TIMESTAMPTZ
) ON COMMIT DROP;

-- name: CreateTempTableFolders :exec
CREATE TEMP TABLE temp_folders (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  workspace_id UUID NOT NULL,
  parent_id UUID,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  trashed_by trashed_by,
  trashed_at TIMESTAMPTZ
) ON COMMIT DROP;
