-- name: GetBillingAccount :one
SELECT * FROM billing_accounts
WHERE workspace_id = sqlc.arg('workspace_id')
FOR UPDATE -- :if @for_update
;

-- name: ListBillingAccountsByPayer :many
SELECT * FROM billing_accounts WHERE payer_id = sqlc.arg('payer_id');

-- name: UpsertBillingAccount :exec
INSERT INTO billing_accounts (workspace_id, provider_customer_id, payer_id, payer_email, payer_name, payer_missing_since, created_at, updated_at)
VALUES (
  sqlc.arg('workspace_id'), sqlc.arg('provider_customer_id'), sqlc.arg('payer_id'), sqlc.arg('payer_email'),
  sqlc.arg('payer_name'), sqlc.arg('payer_missing_since'), sqlc.arg('created_at'), sqlc.arg('updated_at')
)
ON CONFLICT (workspace_id) DO UPDATE SET
  payer_id = EXCLUDED.payer_id,
  payer_email = EXCLUDED.payer_email,
  payer_name = EXCLUDED.payer_name,
  payer_missing_since = EXCLUDED.payer_missing_since,
  updated_at = EXCLUDED.updated_at;
