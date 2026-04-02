-- +goose Up
CREATE TYPE trashed_by AS ENUM ('purpose', 'parent');

CREATE TABLE workspaces (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE folders (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  trashed_by trashed_by,
  trashed_at TIMESTAMPTZ
);

CREATE TABLE notes (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
  tags TEXT[],
  size INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  trashed_by trashed_by,
  trashed_at TIMESTAMPTZ
);

CREATE TABLE note_links (
  source_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  target_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  PRIMARY KEY (source_id, target_id)
);

CREATE INDEX idx_folders_workspace_id
  ON folders(workspace_id);

CREATE INDEX idx_folders_parent_id
  ON folders(parent_id);

CREATE INDEX idx_notes_folder_id
  ON notes(folder_id);

CREATE INDEX idx_workspaces_slug
  ON workspaces(slug);

CREATE INDEX idx_folders_trashed_at
  ON folders(trashed_at) WHERE trashed_at IS NOT NULL;

CREATE INDEX idx_notes_trashed_at
  ON notes(trashed_at) WHERE trashed_at IS NOT NULL;

-- +goose Down

DROP TYPE IF EXISTS trashed_by CASCADE;
DROP TABLE IF EXISTS note_links CASCADE;
DROP TABLE IF EXISTS notes CASCADE;
DROP TABLE IF EXISTS folders CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;
