-- name: GetPolicy :one
SELECT * FROM billing_policy
WHERE id = 1
FOR UPDATE -- :if @for_update
;

-- name: UpsertPolicy :exec
INSERT INTO billing_policy (id, past_due_grace_days, lifetime_refund_window_days, portal_configuration_id, updated_at)
VALUES (1, sqlc.arg('past_due_grace_days'), sqlc.arg('lifetime_refund_window_days'), sqlc.arg('portal_configuration_id'), sqlc.arg('updated_at'))
ON CONFLICT (id) DO UPDATE SET
  past_due_grace_days = EXCLUDED.past_due_grace_days,
  lifetime_refund_window_days = EXCLUDED.lifetime_refund_window_days,
  portal_configuration_id = EXCLUDED.portal_configuration_id,
  updated_at = EXCLUDED.updated_at;
