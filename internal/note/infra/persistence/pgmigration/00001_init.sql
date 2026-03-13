-- +goose Up
CREATE EXTENSION pg_trgm;

CREATE TABLE workspaces (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  root_folder_id UUID,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE folders (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  parent_id UUID REFERENCES folders(id),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_by TEXT,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE notes (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT,
  folder_id UUID NOT NULL REFERENCES folders(id),
  tags TEXT[],
  size INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_by TEXT,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE note_links (
  source_id UUID NOT NULL REFERENCES notes(id),
  target_id UUID NOT NULL REFERENCES notes(id),
  PRIMARY KEY (source_id, target_id)
);

-- Indexes for faster queries
CREATE INDEX idx_folders_workspace_id ON folders(workspace_id);
CREATE INDEX idx_folders_parent_id ON folders(parent_id);
CREATE INDEX idx_notes_folder_id ON notes(folder_id);
CREATE INDEX idx_notes_tags_trgm ON notes USING gin(tags gin_trgm_ops);
CREATE INDEX idx_workspaces_slug ON workspaces(slug);
CREATE INDEX idx_folders_deleted_at ON folders(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_notes_deleted_at ON notes(deleted_at) WHERE deleted_at IS NOT NULL;

-- +goose Down

DROP TABLE IF EXISTS note_links CASCADE;
DROP TABLE IF EXISTS notes CASCADE;
DROP TABLE IF EXISTS folders CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;
DROP EXTENSION IF EXISTS pg_trgm;
