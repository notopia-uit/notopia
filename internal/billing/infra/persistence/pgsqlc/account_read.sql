-- name: ReadGetBillingAccount :one
SELECT * FROM billing_accounts WHERE workspace_id = sqlc.arg('workspace_id');
