# Billing Service

Plans, prices, subscriptions, entitlements. Stripe (sandbox) through hosted Checkout and
Customer Portal. DDD + CQRS like `internal/note`, but without an outbox.
Entrypoint: `cmd/billing/`. HTTP spec: `api/billing/`. gRPC: `proto/billing/`.
Spec and plans: `docs/superpowers/specs/2026-10-11-billing-03-*`, `docs/superpowers/plans/2026-10-11-billing-03-*`.

## Layers

| Dir | Contains | Rules |
| --- | --- | --- |
| `domain/` | Aggregates (`Subscription` + `Refund`, `BillingAccount`, `Price`, `Plan`, `BillingPolicy`), `ResolveEntitlements`, repo interfaces, `UnitOfWork`, `ProviderEventLog` | Pure. `uint8` enums with `IsValid`/`String` |
| `app/` | `cmd_*` / `q_*` use cases, ports (`svc_*`, `readmodel.go`), app models | Commands use `domain.UnitOfWork`; queries use `app.BillingReadModel` only |
| `infra/` | `persistence/` (pgrepo, pgreadmodel, pgmapper, pgjsonb, pgsqlc, pgmigration), `notifier/` (Kafka), `service/` (authorization gRPC), `redis/` | Each adapter maps enums explicitly (`toX`/`fromX`), never casts |
| `controller/` | `http/` (gin strict server), `grpc/`, `health/` | Thin — translate, don't decide |
| `component/` | Validator, `eventtopic.go` (every topic published or consumed) | No inline topic strings elsewhere |
| `errs/` | Typed errors (`Code` is the wire code) | Mapped in `controller/http/err.go` and `controller/grpc/err.go` |

## Persistence

- Write side: `pgsqlc/<aggregate>.sql`, lock-aware via `FOR UPDATE -- :if @for_update` (dynamic filter plugin).
- Read side: `pgsqlc/<aggregate>_read.sql`, every query named `Read…`, used only by `pgreadmodel`.
- `plans.features` / `plans.limits` are typed JSONB (`pgjsonb.Features`/`Limits`); unknown keys are skipped with a warning.
- Webhook de-duplication (`processed_provider_events`) is part of `RepoRegistry`, so "processed" commits with its effect.

## Stripe

- Sandbox only. Stripe does not onboard VN businesses for live payouts; a VN gateway can replace it behind `app.PaymentProvider`.
- Webhooks are fetch-and-sync: payloads are used for ids only, state is read back from Stripe, so ordering does not matter.
- Every Stripe write that must not duplicate carries an idempotency key (`customer:{workspaceId}`, `refund:{refundId}`, `price:{priceId}`).
- Local: `stripe listen --forward-to http://api.notopia.localhost/billing/webhooks/stripe`; keys go in `cmd/billing/.env.local`.

## Notifications

`events.integration.billing.subscription.changed` is **best-effort**: published straight to Kafka after
`uow.Execute` returns, errors only logged. Consumers never cache billing state; they call gRPC on every check.

## Crons

`controller/cron/`, Redis-locked on `billing_redis`: hourly reconcile, daily prune of processed provider events.
`// NOTE:` move into `cmd/billing-worker` when splitting.

## TODO

- Free trial
- Live payments via a VN gateway behind `PaymentProvider`
