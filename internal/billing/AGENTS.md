# Billing Service

Plans, prices, subscriptions, entitlements. Stripe sandbox via hosted Checkout and Customer Portal. DDD + CQRS like `internal/note`, Watermill transactional outbox (`infra/outbox`).
Entrypoint: `cmd/billing/`. HTTP spec: `api/billing/`. gRPC: `proto/billing/`.
Spec and plans: `docs/superpowers/specs/2026-10-11-billing-03-*` (see Revision 2026-11-01 note), `docs/superpowers/plans/2026-10-11-billing-03-*`.
Docs: `docs/src/services/billing/` (domain and data diagrams inline; update with model or schema changes).

## Layers

| Dir | Contains | Rules |
| --- | --- | --- |
| `domain/` | Aggregates (`Subscription` + `Refund`, `BillingAccount`, `Price`, `Plan`, `BillingPolicy`), `ResolveEntitlements`, repo interfaces, `UnitOfWork`, `ProviderEventLog` | Pure. `uint8` enums with `IsValid`/`String` (incl. `Currency`). Immutable value objects (`Money`, `Payer`, `ProviderSnapshot`, `PaymentInfo`, `RefundLine`, `RefundQuote`): unexported fields, validated `NewX`, getters, defensive copies |
| `app/` | `cmd_*` / `q_*` use cases, `SubscriptionCanceller` (quote + cancel-now), ports (`svc_*`, `readmodel.go`), app models | Commands use `domain.UnitOfWork`; queries only `app.BillingReadModel` |
| `infra/` | `persistence/` (pgrepo, pgreadmodel, pgmapper, pgjsonb, pgsqlc, pgmigration), `outbox/` (per-tx publisher + Kafka forwarder), `service/` (authorization gRPC), `redis/` | Adapters map enums explicitly (`toX`/`fromX`), never cast |
| `controller/` | `http/` (gin strict server, Stripe webhook), `grpc/`, `event/` (Kafka consumers, one group per topic, typed via `decode[T]`), `cron/`, `health/` | Thin — translate, don't decide |
| `component/` | Validator, `eventtopic.go` (every topic published or consumed) | No inline topic strings elsewhere |
| `errs/` | Typed errors (`Code` is wire code) | Mapped in `controller/http/err.go` and `controller/grpc/err.go` |

## Persistence

- Repos singleton: injected directly for reads; `pgrepo.registry` gives tx-bound repos to `UnitOfWork.Execute`. `Save` and `forUpdate` reads fail outside unit of work (`guard.go`); writes always in tx.
- Write side: `pgsqlc/<aggregate>.sql`, lock-aware via `FOR UPDATE -- :if @for_update` (dynamic filter plugin).
- Read side: `pgsqlc/<aggregate>_read.sql`, all queries named `Read…`, only `pgreadmodel` uses them.
- `plans.features` / `plans.limits` typed JSONB (`pgjsonb.Features`/`Limits`); unknown keys skipped with warning.
- Webhook de-duplication (`processed_provider_events`) in `RepoRegistry`; "processed" commits with effect.

## Stripe

- Sandbox only. Stripe does not onboard VN businesses for live payouts; VN gateway can replace it behind `app.PaymentProvider`.
- Webhooks fetch-and-sync: payloads provide ids only; state read from Stripe, so order irrelevant.
- Every Stripe write requiring no duplicates carries idempotency key (`customer:{workspaceId}`, `refund:{refundId}`, `price:{priceId}`).
- Use-case order: provider reads and idempotent mutations that must not be lost (change default card) run **before** write transaction; transaction records event (`ProviderEventLog.MarkProcessed`) with state change; handlers end at commit. Local cancel provider side effects driven by billing outbox events, consumed in `controller/event`: `subscription.canceled_now` → `CancelProviderSubscription` (provider cancel, then `MarkProviderSynced`), `subscription.refund_requested` → `SubmitRefund` (`CreateRefund` keyed by refund id; provider `invalid` → `FailRefund`). Reconcile cron retries both for rows still awaiting them.
- `PaymentProvider` only Stripe port; stripe-go imported only in `infra/payment/stripe` and `controller/http/webhook.go`.

## Local development

- Keys: `cmd/billing/.env.local` (`NOTOPIA_BILLING_STRIPE_SECRET_KEY`, `NOTOPIA_BILLING_STRIPE_WEBHOOK_SECRET`).
- Seed implementation: `internal/billing/seed`; run `nx run billing:migrate` before `nx run billing:seed`.
- `just seed-billing`: Stripe products/prices + portal configuration, idempotent (needed every environment).
- `just fixtures-billing`: dev-only workspace with fake Stripe ids (no Stripe calls).
- `just stripe-listen`: forwards sandbox events to `http://api.notopia.localhost/billing/webhooks/stripe`; put printed `whsec_` in `.env.local`.
- Test cards: `4242 4242 4242 4242` (succeeds), `4000 0000 0000 0341` (card saves, charge fails → `past_due`), `4000 0025 0000 3155` (3-D Secure), `4000 0000 0000 9995` (declined). Any future expiry and CVC.
- Manual walkthrough: `tmp/TEST_FLOW_BILLING.md`; diagrams: `docs/src/references/sequence/billing.md`.

## Events

Repo `Save` writes aggregate's popped events into `outbox_messages` (Watermill `queue` schema, `pgmigration/00002_outbox.sql`) in same transaction
(`pgrepo.PublisherFactory` → `outbox.PublisherFactory`); `outbox.Forwarder` relays to Kafka, keyed by
`partition_key` metadata (aggregate id, workspace id for integration events). Topics in
`component/eventtopic.go`: domain events go to `events.internal.billing.{aggregate}.{event}`;
`SubscriptionStateChangedEvent` (recorded on plan or status change, never duplicate checkout) maps to
`share.SubscriptionChangedEvent` on `events.integration.billing.subscription.changed`.
Billing also consumes own `subscription.canceled_now` and `subscription.refund_requested` topics (phase 2 above). No Kafka `Notifier` or `RefundRunner`; never publish to Kafka outside outbox.
Integration event consumers never cache billing state; call gRPC every check.

## Crons

`controller/cron/`, Redis-locked on `billing_redis`: hourly reconcile, daily prune of processed provider events.
`pkg/cron.Scheduler` takes `gocron.Locker` and `*slog.Logger` directly. Reconcile re-syncs stale subscriptions and is phase 2 safety net (retries pending provider cancels and refund submissions). Each job configured by `NOTOPIA_BILLING_CRON_{RECONCILE_SUBSCRIPTIONS|PRUNE_PROVIDER_EVENTS}_{NAME|SCHEDULE|TIMEOUT|ENABLED}` (defaults in `config.New`; `..._ENABLED=false` disables).
`// NOTE:` move into `cmd/billing-worker` when splitting.

## TODO

- Free trial
- Live payments via VN gateway behind `PaymentProvider` (see `docs/src/stories/BILLING_PAYMENT_PROVIDERS.md`)
- More currencies (prices VND, zero-decimal)
