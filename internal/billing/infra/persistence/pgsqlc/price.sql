-- name: GetPrice :one
SELECT * FROM prices
WHERE id = sqlc.arg('id')
FOR UPDATE -- :if @for_update
;

-- name: GetPriceByProviderPriceID :one
SELECT * FROM prices WHERE provider_price_id = sqlc.arg('provider_price_id');

-- name: GetActivePrice :one
SELECT * FROM prices
WHERE plan = sqlc.arg('plan') AND billing_interval = sqlc.arg('billing_interval') AND currency = sqlc.arg('currency') AND active
FOR UPDATE -- :if @for_update
;

-- name: ListActivePrices :many
SELECT * FROM prices WHERE active ORDER BY plan, billing_interval;

-- name: ListPrices :many
SELECT * FROM prices ORDER BY created_at DESC;

-- name: UpsertPrice :exec
INSERT INTO prices (id, plan, billing_interval, amount, currency, provider_price_id, lookup_key, active, created_at, archived_at)
VALUES (
  sqlc.arg('id'), sqlc.arg('plan'), sqlc.arg('billing_interval'), sqlc.arg('amount'), sqlc.arg('currency'),
  sqlc.arg('provider_price_id'), sqlc.arg('lookup_key'), sqlc.arg('active'), sqlc.arg('created_at'), sqlc.arg('archived_at')
)
ON CONFLICT (id) DO UPDATE SET active = EXCLUDED.active, archived_at = EXCLUDED.archived_at;
