-- name: ReadListActivePrices :many
SELECT * FROM prices WHERE active ORDER BY plan, billing_interval;

-- name: ReadListPrices :many
SELECT * FROM prices ORDER BY created_at DESC;
