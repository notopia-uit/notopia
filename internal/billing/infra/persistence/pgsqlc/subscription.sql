-- name: GetSubscription :one
SELECT * FROM subscriptions
WHERE id = sqlc.arg('id')
FOR UPDATE -- :if @for_update
;

-- name: GetCurrentSubscription :one
SELECT * FROM subscriptions
WHERE workspace_id = sqlc.arg('workspace_id') AND status NOT IN ('canceled', 'incomplete_expired')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionByProviderSubscriptionID :one
SELECT * FROM subscriptions
WHERE provider_subscription_id = sqlc.arg('provider_subscription_id')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionByProviderPaymentID :one
SELECT * FROM subscriptions
WHERE provider_payment_id = sqlc.arg('provider_payment_id')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionIDByRefundID :one
SELECT subscription_id FROM refunds WHERE id = sqlc.arg('id');

-- name: ListCurrentSubscriptionsByWorkspaces :many
SELECT * FROM subscriptions
WHERE workspace_id = ANY(sqlc.arg('workspace_ids')::uuid[]) AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ListSubscriptionsNeedingReconcile :many
SELECT * FROM subscriptions
WHERE status NOT IN ('canceled', 'incomplete_expired')
  AND billing_interval <> 'lifetime'
  AND (current_period_end + INTERVAL '1 hour' < sqlc.arg('now')::timestamptz
       OR provider_synced_at < sqlc.arg('now')::timestamptz - INTERVAL '24 hours')
ORDER BY provider_synced_at
LIMIT sqlc.arg('row_limit');

-- name: ListSubscriptionIDsWithUnsubmittedRefunds :many
SELECT DISTINCT subscription_id FROM refunds
WHERE status = 'pending' AND provider_refund_id IS NULL
LIMIT sqlc.arg('row_limit');

-- name: UpsertSubscription :exec
INSERT INTO subscriptions (
  id, workspace_id, plan, billing_interval, price_id, status, provider_subscription_id, provider_payment_id,
  current_period_start, current_period_end, cancel_at_period_end, canceled_at, paid_at, provider_synced_at,
  created_at, updated_at
) VALUES (
  sqlc.arg('id'), sqlc.arg('workspace_id'), sqlc.arg('plan'), sqlc.arg('billing_interval'), sqlc.arg('price_id'),
  sqlc.arg('status'), sqlc.arg('provider_subscription_id'), sqlc.arg('provider_payment_id'),
  sqlc.arg('current_period_start'), sqlc.arg('current_period_end'), sqlc.arg('cancel_at_period_end'),
  sqlc.arg('canceled_at'), sqlc.arg('paid_at'), sqlc.arg('provider_synced_at'),
  sqlc.arg('created_at'), sqlc.arg('updated_at')
)
ON CONFLICT (id) DO UPDATE SET
  plan = EXCLUDED.plan,
  billing_interval = EXCLUDED.billing_interval,
  price_id = EXCLUDED.price_id,
  status = EXCLUDED.status,
  current_period_start = EXCLUDED.current_period_start,
  current_period_end = EXCLUDED.current_period_end,
  cancel_at_period_end = EXCLUDED.cancel_at_period_end,
  canceled_at = EXCLUDED.canceled_at,
  paid_at = EXCLUDED.paid_at,
  provider_synced_at = EXCLUDED.provider_synced_at,
  updated_at = EXCLUDED.updated_at;

-- name: ListRefundsBySubscriptionIDs :many
SELECT * FROM refunds WHERE subscription_id = ANY(sqlc.arg('subscription_ids')::uuid[]) ORDER BY created_at;

-- name: UpsertRefund :exec
INSERT INTO refunds (id, subscription_id, provider_refund_id, payment_ref, amount, currency, reason, status, failure_message, created_at, updated_at)
VALUES (
  sqlc.arg('id'), sqlc.arg('subscription_id'), sqlc.arg('provider_refund_id'), sqlc.arg('payment_ref'), sqlc.arg('amount'),
  sqlc.arg('currency'), sqlc.arg('reason'), sqlc.arg('status'), sqlc.arg('failure_message'), sqlc.arg('created_at'), sqlc.arg('updated_at')
)
ON CONFLICT (id) DO UPDATE SET
  provider_refund_id = EXCLUDED.provider_refund_id,
  status = EXCLUDED.status,
  failure_message = EXCLUDED.failure_message,
  updated_at = EXCLUDED.updated_at;
