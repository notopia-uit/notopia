# Billing Review Refactor

Branch: `refactor/billing-review`. One conventional commit per task. Before each commit:
`nx lint billing --fix` (and other touched go projects), `go test ./internal/billing/... ./pkg/...`,
`nx wire billing` if providers changed (never hand-edit `wire_gen.go`, never `go generate`),
`nx gen api` / `nx gen proto` for spec changes, `oxfmt` for JS/TS. Commit trailer:
`Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

Rules: no unnecessary comments; no `common.go` grab-bag helpers; don't swallow errors; app and domain
never import stripe-go (controller and infra may); commands never read via read model; value objects
immutable with private fields.

Deferred (not this PR): handler decorator refactor, see `tmp/billing/deferred-handler-decorator.md`.

## T1 Mechanical renames (haiku)

- [x] `domain.Status` → `SubscriptionStatus`, constants `StatusX` → `SubscriptionStatusX` (all layers, mappers, tests)
- [x] Inline `latest(a, b)` in `Subscription.QuoteCancelNow`, delete the func
- [x] `domain.LookupKey(plan, interval)` → `PriceLookupKey(plan, interval)`; `Price.lookupKey`/`LookupKey()` → `providerLookupKey`/`ProviderLookupKey()`; `UnmarshalPriceParams.LookupKey` → `ProviderLookupKey` (DB column and API field names unchanged)
- [x] `app.Canceller` → `SubscriptionCanceller` (file `subscriptioncanceller.go`, field `subscriptionCanceller`, `ProvideSubscriptionCanceller`)
- [x] Delete `app/common.go`: inline `hasCode` (`var e errs.Error; errors.As(err, &e) && e.Code() == code`) and `cancelAtProvider` (with its log) at every call site
- [x] `pgrepo/common.go`: inline `wrap` → `errs.NewInternal("failed to <msg>", err)`, `notFound` → `errors.Is(err, pgx.ErrNoRows)`. `drain`/`typeName` stay until T5
- [x] `pkg/cron`: delete `slogAdapter`, pass `gocron.WithLogger(logger)`; drop `type Locker = gocron.Locker`, use `gocron.Locker`; delete billing `infra/redis.NewCronLocker`, wire `pkgcron.NewRedisLocker` directly (bind `*redis.Client` → `redis.UniversalClient`); same for note

## T2 Immutable value objects (sonnet)

- [x] `Money`, `Payer`, `ProviderSnapshot`, `PaymentInfo`, `RefundLine`, `RefundQuote`: private fields, getters, validating constructors (`NewMoney`, `NewPayer`, `NewProviderSnapshot(ProviderSnapshotParams)`, `NewPaymentInfo(...)`, `NewRefundLine(...)`, `NewRefundQuote(lines, unusedSeconds, periodSeconds)` computing the total)
- [x] `Money.Add(Money) (Money, error)`, currency mismatch → invalid
- [x] `Currency` uint8 enum (`CurrencyUSD`, … whatever is in use) with `String`/`IsValid`/`ParseCurrency`; explicit mappers at ports
- [x] `QuoteCancelNow`/`FullRefundQuote` build lines then `NewRefundQuote`, no field mutation

## T3 Singleton repos + Save guard (sonnet)

- [ ] Wire provides `domain.SubscriptionRepo`, `BillingAccountRepo`, `PriceRepo`, `PlanRepo`, `BillingPolicyRepo`, `ProviderEventLog` backed by the pool (no tx, no publisher)
- [ ] `Save` on a pool-backed repo returns `errs.NewInternal("save must run inside a unit of work", nil)`
- [ ] App handlers use injected repos for plain reads; `uow.Execute` only for writes (lock + save)

## T4 Note outbox → queue schema (sonnet)

- [ ] `sql.PostgreSQLQueueSchema{GenerateMessagesTableName}` + `sql.PostgreSQLQueueOffsetsAdapter{DeleteOnAck: true, GenerateMessagesTableName}`, no consumer group
- [ ] Goose migration creates the queue table (drop old default-schema tables); `InitializeSchema: false`
- [ ] Explicit matching `ForwarderTopic` in `forwarder.PublisherConfig` and `forwarder.Config`

## T5 Billing outbox, phase 1 (opus)

- [ ] `internal/billing/infra/outbox/`: per-tx `sql.NewPublisher(sql.TxFromPgx(tx))` → `forwarder.NewPublisher` → otel decorator; forwarder daemon (`sql.BeginnerFromPgx(pool)` subscriber → kafka publisher) run from `server.go`. Queue schema as T4. Config `DomainEvent` like note
- [ ] UoW creates the publisher per tx and passes it to repos; `Save` publishes `PopEvents()` (replaces `drain`)
- [ ] Topics: domain `events.internal.billing.{aggregate}.{event}`; integration `events.integration.billing.subscription.changed` with `share.SubscriptionChangedEvent`
- [ ] Aggregate records `SubscriptionStateChangedEvent{WorkspaceID, Plan, Status}` whenever plan or status changes; outbox infra maps it to the share payload on the integration topic
- [ ] Delete `app.Notifier`, `notifySubscription`, `infra/notifier`

## T6 Billing outbox, phase 2 (opus)

- [ ] Billing event controller consumes its own domain topics:
  - `SubscriptionCanceledNowEvent` → `CancelProviderSubscriptionCmd` (provider cancel, then `MarkProviderSynced` in a write tx; no-op if already synced / no provider id)
  - `RefundRequestedEvent` (refund without provider id) → `SubmitRefundCmd{SubscriptionID, RefundID}` (provider `CreateRefund` keyed by refund id; `CodeInvalid` → `FailRefund`; other errors returned → router retry)
- [ ] Delete `RefundRunner` and post-commit tails in `SubscriptionCanceller`, `CompleteCheckout`, etc.; handlers end at commit
- [ ] Reconcile cron stays as a safety net
- [ ] Events carry what consumers need (subscription id, refund id)

## T7 Webhook typing (sonnet) — dedicated commit

- [ ] `api/billing/paths/webhooks_stripe.yaml` body schema `x-go-type: json.RawMessage`, `x-go-type-import: {path: encoding/json}`; `nx gen api`
- [ ] Delete `captureRawBody`, `rawBodyKey`, `rawBody(ctx)`
- [ ] Controller (may import stripe-go): verify with `webhook.ConstructEventWithOptions(*req.Body, …)`, switch on `ev.Type` with `stripe.EventTypeX`, parse `ev.Data.Raw` into `stripe.CheckoutSession` / `stripe.Subscription` / `stripe.Invoice` / `stripe.Refund`; delete `stripeObject`/`stripeID`
- [ ] App stays agnostic: commands take plain ids + event id/type string; remove `eventCheckoutCompleted` from app (controller passes event type)
- [ ] Manual check: `stripe listen` + `stripe trigger checkout.session.completed` verifies signature with the bound bytes

## T8 Docs restructure (sonnet, worktree)

- [ ] New `docs/src` layout: `getting-started/`, `architecture/{overview,messaging,auth-and-gateway}.md`, `services/{note,document,authorization,billing,search-worker}/{index,domain,data,flows,events}.md`, `api/{rest,grpc}`, `conventions/`
- [ ] Move existing content, inline every `.d2`/`.puml` into fenced code blocks; delete `docs/src/diagrams/`
- [ ] protoc-gen-doc: mise `"github:pseudomuto/protoc-gen-doc" = "1.5.1"`, buf plugin `local: protoc-gen-doc`, `out: docs/src/api/grpc`, `strategy: all`, `opt: [markdown,index.md]`

## T9 Billing docs (sonnet, after T1–T7)

- [ ] `services/billing/domain.md` PlantUML class diagram (domain + app); `services/billing/data.md` d2 db diagram from migrations (no outbox table)
- [ ] Update `docs/superpowers/specs/2026-10-11-billing-03-billing-service-design.md` and `internal/billing/AGENTS.md`
