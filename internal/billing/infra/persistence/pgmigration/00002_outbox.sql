-- +goose Up
CREATE TABLE outbox_messages (
  "offset" SERIAL PRIMARY KEY,
  "uuid" VARCHAR(36) NOT NULL,
  "payload" JSON DEFAULT NULL,
  "metadata" JSON DEFAULT NULL,
  "acked" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS outbox_messages;
