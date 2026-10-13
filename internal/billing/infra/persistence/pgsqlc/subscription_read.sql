-- name: ReadGetCurrentSubscription :one
SELECT * FROM subscriptions
WHERE workspace_id = sqlc.arg('workspace_id') AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ReadListCurrentSubscriptionsByWorkspaces :many
SELECT * FROM subscriptions
WHERE workspace_id = ANY(sqlc.arg('workspace_ids')::uuid[]) AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ReadListRefundsBySubscriptionIDs :many
SELECT * FROM refunds WHERE subscription_id = ANY(sqlc.arg('subscription_ids')::uuid[]) ORDER BY created_at;
