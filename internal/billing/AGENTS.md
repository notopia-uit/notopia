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
| `controller/` | `http/` (gin strict server, Stripe webhook), `grpc/`, `event/` (Kafka consumers, one consumer group per topic, typed via `decode[T]`), `cron/`, `health/` | Thin — translate, don't decide |
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
- Webhook use-case ordering: provider mutations that must not be lost (cancel a duplicate, change default card) run **before** the write transaction and are idempotent; the transaction records the event (`ProviderEventLog.MarkProcessed`) with the state change; refunds run **after** commit via `RefundRunner` and the reconcile cron retries `pending` refunds without a provider id.
- `PaymentProvider` is the only port to Stripe; stripe-go is imported only in `infra/payment/stripe` and `controller/http/webhook.go`.

## Local development

- Keys: `cmd/billing/.env.local` (`NOTOPIA_BILLING_STRIPE_SECRET_KEY`, `NOTOPIA_BILLING_STRIPE_WEBHOOK_SECRET`), `cmd/billingseed/.env.local` (`NOTOPIA_BILLINGSEED_STRIPE_SECRET_KEY`).
- `just seed-billing`: Stripe products/prices + portal configuration, idempotent (needed in every environment).
- `just fixtures-billing`: dev-only workspace with fake Stripe ids (no Stripe calls).
- `just stripe-listen`: forwards sandbox events to `http://api.notopia.localhost/billing/webhooks/stripe`; put the printed `whsec_` in `.env.local`.
- Test cards: `4242 4242 4242 4242` (succeeds), `4000 0000 0000 0341` (card saves, charge fails → `past_due`), `4000 0025 0000 3155` (3-D Secure), `4000 0000 0000 9995` (declined). Any future expiry and CVC.
- Full manual walkthrough: `tmp/TEST_FLOW_BILLING.md`; diagrams: `docs/src/references/sequence/billing.md`.

## Notifications

`events.integration.billing.subscription.changed` is **best-effort**: published straight to Kafka after
`uow.Execute` returns, errors only logged. Consumers never cache billing state; they call gRPC on every check.

## Crons

`controller/cron/`, Redis-locked on `billing_redis`: hourly reconcile, daily prune of processed provider events.
`// NOTE:` move into `cmd/billing-worker` when splitting.

## TODO

- Free trial
- Live payments via a VN gateway behind `PaymentProvider` (see `story/BILLING_PAYMENT_PROVIDERS.md`)
- More currencies (prices are VND, zero-decimal)
