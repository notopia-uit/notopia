-- name: InsertProviderEvent :execrows
INSERT INTO processed_provider_events (event_id, event_type, received_at)
VALUES (sqlc.arg('event_id'), sqlc.arg('event_type'), sqlc.arg('received_at'))
ON CONFLICT (event_id) DO NOTHING;

-- name: DeleteProviderEventsBefore :execrows
DELETE FROM processed_provider_events WHERE received_at < sqlc.arg('received_before');
