-- name: GetPlan :one
SELECT * FROM plans
WHERE tier = sqlc.arg('tier')
FOR UPDATE -- :if @for_update
;

-- name: ListPlans :many
SELECT * FROM plans ORDER BY tier;

-- name: UpsertPlan :exec
INSERT INTO plans (tier, features, limits, updated_at)
VALUES (sqlc.arg('tier'), sqlc.arg('features'), sqlc.arg('limits'), sqlc.arg('updated_at'))
ON CONFLICT (tier) DO UPDATE SET features = EXCLUDED.features, limits = EXCLUDED.limits, updated_at = EXCLUDED.updated_at;
