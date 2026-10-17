-- +goose Up
ALTER TABLE subscriptions ADD COLUMN past_due_since TIMESTAMPTZ;

-- +goose Down
ALTER TABLE subscriptions DROP COLUMN past_due_since;
