-- +goose Up
DROP TABLE IF EXISTS eventsToForward;
DROP TABLE IF EXISTS "watermill_offsets_forwarder_topic";

CREATE TABLE eventsToForward (
  "offset" SERIAL PRIMARY KEY,
  "uuid" VARCHAR(36) NOT NULL,
  "payload" JSON DEFAULT NULL,
  "metadata" JSON DEFAULT NULL,
  "acked" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS eventsToForward;
