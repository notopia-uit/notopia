-- +goose Up
CREATE EXTENSION pg_trgm;

CREATE TABLE notes (
  id UUID PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMPTZ
);

-- CREATE INDEX idx_tags_trgm
-- ON projects
-- USING gin (tags gin_trgm_ops);

-- +goose Down

DROP EXTENSION pg_trgm;
DROP TABLE notes;
-- DROP INDEX idx_tags_trgm;
