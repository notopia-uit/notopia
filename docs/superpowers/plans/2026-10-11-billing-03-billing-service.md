# Billing 03 — Billing Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A new Go `billing` service owning plans, prices, subscriptions and entitlements, integrated with Stripe (sandbox) via hosted Checkout and Customer Portal, exposing entitlements over gRPC and HTTP.

**Architecture:** Mirrors `internal/note` (DDD + CQRS, wire, sqlc + goose, gin strict server, gRPC, Watermill Kafka subscriber, `pkg/cron`). No outbox: other services read billing over gRPC; billing only publishes a best-effort `subscription.changed` notification to Kafka after commit (note forwards it to SSE). Admin plan edits are not pushed. Stripe sits behind `app.PaymentProvider`; webhooks use fetch-and-sync and are de-duplicated in the same transaction as their effect.

**Tech Stack:** Go 1.26, gin + oapi-codegen strict server, grpc + protovalidate, pgx v5 + sqlc + goose, Watermill Kafka, go-redis v9 + `pkg/cron`, `github.com/stripe/stripe-go/v87` `v87.0.0`, OpenTelemetry via `pkg/otel`.

**Spec:** `docs/superpowers/specs/2026-10-11-billing-03-billing-service-design.md`

**Depends on:** plan 01 Task 5 (`events.integration.note.workspace.deleted`), plan 02 (Tasks 1, 2, 3, 5, 5b, 7).

## Decisions made while planning (deviations from the spec — confirm in review)

1. **Payer change is claim-only.** `POST /billing/workspaces/{id}/payer/claim` lets any owner who is not the payer become payer by completing Checkout `mode=setup`. No separate "transfer to user X" endpoint: billing cannot get another user's email/name (only the caller's, from gateway headers), and owners are equal anyway. The note-side guard (plan 04) tells the payer "another owner must claim billing first".
2. **Webhook de-duplication lives in the transaction:** `domain.ProviderEventLog` is part of `RepoRegistry` (not a standalone app port) so "event processed" and its effect commit together.
3. **CQRS split in persistence:** writes go through `pgrepo` + domain sqlc queries (`FOR UPDATE -- :if @for_update`); query handlers (`q_*`) go through `app.BillingReadModel` + `pgreadmodel` + separate `Read…` sqlc queries, even where the SQL looks alike. Shared sqlc ↔ domain mapping lives in `pgmapper`.
4. **Refund webhooks:** `refund.created`, `refund.updated`, `refund.failed` only (`charge.refunded` dropped; the refund events carry everything).
5. **Portal configuration** is created anew (`Create`) whenever the active price set changes; the newest id is stored in `billing_policy`. Old configurations stay in Stripe unused.
6. **Payer name** comes from `X-Forwarded-Name` (added to Traefik `headerMap` in plan 02 Task 1); falls back to email when empty.
7. **Enums are `uint8` in Go; each wire keeps its own enum representation.** Postgres keeps `CREATE TYPE ... AS ENUM`, OpenAPI and `api/share` keep string enums, the proto uses enums `PlanTier`/`SubscriptionStatus` (zero `*_UNSPECIFIED`, required fields `defined_only` + `not_in: [0]`) while feature/limit maps stay string-keyed (proto map keys cannot be enums). Each adapter maps explicitly (see Global Constraints). Feature/limit keys: admin HTTP input rejects unknown keys (400); a DB read skips unknown keys with a warning log, so removing a key from code never breaks entitlement reads.
8. **Refund quote is per invoice payment.** An upgrade is invoiced immediately (`always_invoice`), so one period can have several paid invoices. Cancel-now quotes every paid invoice payment of the current period on its own coverage window and creates one `Refund` per payment (Stripe refunds per PaymentIntent). The refund cap is per payment.
9. **The Stripe webhook is an inbound adapter in `controller/http`.** It verifies the signature, reads type and object id, and calls `Cmds.CompleteCheckout` / `SyncSubscription` / `SyncRefund` directly. `PaymentProvider` has no `ParseWebhook`, and there is no `HandleWebhook` command or `WebhookEventType` enum. The endpoint is Stripe-specific anyway (`/billing/webhooks/stripe`, `Stripe-Signature`); another gateway adds its own endpoint that calls the same commands.
10. **`PaymentProvider` is Stripe-shaped on purpose.** Hosted Checkout, Customer Portal, provider-owned recurring billing and per-invoice refunds are Stripe concepts. A VN gateway without recurring billing needs new use cases (we own renewals), not just a new adapter. The port keeps Stripe types out of `app`/`domain`; it does not promise drop-in replacement.
11. **App settings are app-owned, mapped at the composition root.** `app.Settings` lists only what use cases need; `internal/billing/settings.go` fills it from the viper-bound `config.Billing`/`commonconfig.Identity`. `app` never imports `config`, and `config` never imports `app`.
12. **Seed vs fixtures.** `cmd/billingseed` (mandatory catalog: prices, portal configuration) has its own small config (`NOTOPIA_BILLINGSEED_*`: database, Stripe key, catalog, service account) and calls only `EnsureDefaults`, `ListPrices`, `CreatePrice`, `EnsurePortalConfiguration`. `cmd/billingfixtures` (dev-only: lifetime `pro` for workspace 110, payer user 110) saves domain objects through the repositories, with fake Stripe ids. There is no `SeedCatalog` command.

## Global Constraints

- Env prefix `NOTOPIA_BILLING_`; ports HTTP `8084`, gRPC `18084`, health `28084`; Postgres host port `5435`, Redis host port `6375`.
- Currency `vnd` (zero-decimal: `49000` = 49.000₫). Money is `int64` minor units + `currency`.
- `-1` = unlimited. Feature keys: `auto_commit`, `workspace_graph`. Limit keys: `members`, `published_notes`, `note_graph_max_depth`, `auto_commit_min_interval_seconds`, `revision_history_days`, `trash_retention_days`, `ai_credits_monthly`.
- Stripe API version is the one pinned by stripe-go v87 (basil+): periods on `items.data[].current_period_*`; invoice → subscription via `parent.subscription_details.subscription`; invoice payments via `/v1/invoice_payments`.
- Webhooks: `webhook.ConstructEventWithOptions(..., webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})` in `controller/http/webhook.go` — payloads are only used for ids.
- Stripe tracing comes from an `otelhttp` transport on the stripe-go client; no hand-written spans in the adapter.
- Every Stripe write that must not duplicate carries an idempotency key (`customer:{workspaceId}`, `refund:{refundId}`, `price:{priceId}`).
- Every mutating domain method emits a domain event; repos drain them on `Save` and log at debug.
- The notification (`events.integration.billing.subscription.changed`) is published by use cases **after** `uow.Execute` returns nil; publish errors are logged, never returned.
- Go enums (domain and app) are `uint8` `iota` types with a zero `XUnspecified` (or `XNone` where "none" is a real state), plus `IsValid()` and `String()`. `String()` returns the lowercase/snake value used on the wire and in logs. Never cast an enum with `string(x)` or `domain.X(s)` across a port: each adapter (pgrepo, HTTP, gRPC, Kafka notifier, Stripe) owns explicit `toX`/`fromX` mappers between the domain enum and its own type. Feature/limit keys cross ports through `k.String()` outbound and `domain.ParseFeatureKey`/`ParseLimitKey` inbound (unknown key = error; HTTP returns 400, the pgrepo JSONB reader skips the key and logs a warning).
- `IsValid()` returns false for the `Unspecified` zero (true for real "none" states such as `StatusNone`), deliberately unlike note's `TrashedBy.IsValid()`, which accepts `Unspecified`.
- Exception: `errs.Code` stays `type Code string`, because it is the wire error code itself (same as `internal/note/errs/common.go`).
- Codegen: `nx run api:gen`, `nx run proto:gen`, `nx run billing:gen:sqlc`, `nx run billing:gen:wire`. Never edit generated files.
- Go tests use testify (`require` for preconditions, `assert` for checks); no `t.Fatal`/`t.Errorf` assertions. Skip obvious tests. `nx lint billing` before each commit. Conventional commits.

## Review Focus

- Stripe sends the same `checkout.session.completed` twice or concurrently: one subscription row, no double refund (Task 8 test on `CompleteCheckout` + inbox).
- Renewal webhook arrives after the stored `current_period_end`: workspace keeps access (status stays `active` until Stripe says otherwise) (Task 2 `GrantsAccess` test).
- Cancel-now when a payment was already partially refunded: refund never exceeds what that payment paid (Task 2 refund invariant test).
- Cancel-now after a mid-period upgrade: both the original and the proration invoice are refunded for their unused part (Task 2 `TestQuoteCancelNowAfterUpgrade`).
- Two checkouts completed for one workspace (two tabs): the second is cancelled and fully refunded (Task 8 test).
- Webhook with a bad signature or from a different Stripe API version: bad signature → 400; version mismatch accepted (Task 8).

---

### Task 1: Contracts — proto and OpenAPI

**Files:**

- Create: `proto/billing/billing.proto`
- Create: `api/billing/openapi.yaml`, `api/billing/project.json`, `api/billing/oapi-codegen-server.yaml`, `api/billing/oapi-codegen-model.yaml`
- Create: `api/billing/paths/*.yaml` (listed below), `api/billing/components/schemas/*.yaml`, `api/billing/components/parameters/*.yaml`
- Modify: `api/redocly.yaml`, `api/project.json` (`implicitDependencies`), `api/package.json` (`redocly:join`), `api/api.go` (`BillingSpec`), `apps/api-web/project.json` (sources/dependsOn, mirror the `api/note` entries)

**Interfaces:**

- Produces: `pb.BillingServiceServer/Client` with `GetWorkspaceEntitlements`, `BatchGetWorkspaceEntitlements`, `GetWorkspacePayer`, `ListPlans`; proto enums `pb.PlanTier` (`PLAN_TIER_UNSPECIFIED = 0`, `FREE`, `PLUS`, `PRO`) and `pb.SubscriptionStatus` (`SUBSCRIPTION_STATUS_UNSPECIFIED = 0`, `NONE`, `INCOMPLETE`, `ACTIVE`, `PAST_DUE`, `UNPAID`, `CANCELED`, `INCOMPLETE_EXPIRED`), used by `Entitlements.plan`/`status` and `BillingPlan.tier` (TS: `PlanTier`, `SubscriptionStatus` enums in `@notopia-uit/pb/billing`); Go package `pkg/api/billing` with `StrictServerInterface`; TS `@notopia-uit/pb/billing`; `api.BillingSpec`.

- [ ] **Step 1: Proto**

```proto
// proto/billing/billing.proto
syntax = "proto3";

package billing;

import "buf/validate/validate.proto";
import "google/protobuf/timestamp.proto";

service BillingService {
  rpc GetWorkspaceEntitlements(GetWorkspaceEntitlementsRequest) returns (GetWorkspaceEntitlementsResponse);
  rpc BatchGetWorkspaceEntitlements(BatchGetWorkspaceEntitlementsRequest) returns (BatchGetWorkspaceEntitlementsResponse);
  rpc GetWorkspacePayer(GetWorkspacePayerRequest) returns (GetWorkspacePayerResponse);
  rpc ListPlans(ListPlansRequest) returns (ListPlansResponse);
}

enum PlanTier {
  PLAN_TIER_UNSPECIFIED = 0;
  PLAN_TIER_FREE = 1;
  PLAN_TIER_PLUS = 2;
  PLAN_TIER_PRO = 3;
}

enum SubscriptionStatus {
  SUBSCRIPTION_STATUS_UNSPECIFIED = 0;
  SUBSCRIPTION_STATUS_NONE = 1;
  SUBSCRIPTION_STATUS_INCOMPLETE = 2;
  SUBSCRIPTION_STATUS_ACTIVE = 3;
  SUBSCRIPTION_STATUS_PAST_DUE = 4;
  SUBSCRIPTION_STATUS_UNPAID = 5;
  SUBSCRIPTION_STATUS_CANCELED = 6;
  SUBSCRIPTION_STATUS_INCOMPLETE_EXPIRED = 7;
}

// features/limits stay string-keyed (proto map keys cannot be enums): keys are FeatureKey/LimitKey names.
message Entitlements {
  string workspace_id = 1;
  PlanTier plan = 2 [(buf.validate.field).enum = {defined_only: true, not_in: [0]}];
  SubscriptionStatus status = 3 [(buf.validate.field).enum = {defined_only: true, not_in: [0]}];
  optional google.protobuf.Timestamp access_until = 4;
  map<string, bool> features = 5;
  map<string, int64> limits = 6;
}

message GetWorkspaceEntitlementsRequest {
  string workspace_id = 1 [(buf.validate.field).string.uuid = true];
}

message GetWorkspaceEntitlementsResponse {
  Entitlements entitlements = 1;
}

message BatchGetWorkspaceEntitlementsRequest {
  repeated string workspace_ids = 1 [(buf.validate.field).repeated = {
    min_items: 1
    max_items: 500
    items: {string: {uuid: true}}
  }];
}

message BatchGetWorkspaceEntitlementsResponse {
  map<string, Entitlements> entitlements = 1;
}

message GetWorkspacePayerRequest {
  string workspace_id = 1 [(buf.validate.field).string.uuid = true];
}

message GetWorkspacePayerResponse {
  optional string payer_id = 1;
  bool requires_payer = 2;
}

message BillingPlan {
  PlanTier tier = 1 [(buf.validate.field).enum = {defined_only: true, not_in: [0]}];
  map<string, bool> features = 2;
  map<string, int64> limits = 3;
}

message ListPlansRequest {}

message ListPlansResponse {
  repeated BillingPlan plans = 1;
}
```

Run: `nx run proto:gen`. Check the generated Go package path (same `pkg/pb` package as note/authorization, or a sub-package) and use that import everywhere below as `pb`.

- [ ] **Step 2: OpenAPI surface**

`api/billing/openapi.yaml`:

```yaml
## yaml-language-server: $schema=https://www.schemastore.org/openapi-3.X.json
# vim: ft=yaml.openapi
openapi: 3.0.4

servers:
  - url: http://api.notopia.localhost

info:
  description: >
    Notopia Billing API - plans, subscriptions (Stripe) and workspace entitlements.
  version: 0.0.0
  title: Notopia API
  license:
    name: MIT
    url: https://github.com/notopia-uit/notopia/blob/main/LICENSE
  x-apiName: billing

tags:
  - name: Plan
    description: Public plan catalog
  - name: WorkspaceBilling
    description: Subscription, entitlements and payer of a workspace
  - name: BillingAdmin
    description: Admin-only catalog management (group notopia-admins)
  - name: Webhook
    description: Payment provider webhooks

paths:
  /billing/plans:
    $ref: ./paths/plans.yaml
  /billing/workspaces/{workspaceId}/entitlements:
    $ref: ./paths/workspaces_{workspaceId}_entitlements.yaml
  /billing/workspaces/{workspaceId}/subscription:
    $ref: ./paths/workspaces_{workspaceId}_subscription.yaml
  /billing/workspaces/{workspaceId}/checkout-session:
    $ref: ./paths/workspaces_{workspaceId}_checkout-session.yaml
  /billing/workspaces/{workspaceId}/portal-session:
    $ref: ./paths/workspaces_{workspaceId}_portal-session.yaml
  /billing/workspaces/{workspaceId}/subscription/cancel-now/preview:
    $ref: ./paths/workspaces_{workspaceId}_subscription_cancel-now_preview.yaml
  /billing/workspaces/{workspaceId}/subscription/cancel-now:
    $ref: ./paths/workspaces_{workspaceId}_subscription_cancel-now.yaml
  /billing/workspaces/{workspaceId}/payer/claim:
    $ref: ./paths/workspaces_{workspaceId}_payer_claim.yaml
  /billing/admin/plans:
    $ref: ./paths/admin_plans.yaml
  /billing/admin/plans/{tier}:
    $ref: ./paths/admin_plans_{tier}.yaml
  /billing/admin/policy:
    $ref: ./paths/admin_policy.yaml
  /billing/admin/prices:
    $ref: ./paths/admin_prices.yaml
  /billing/admin/prices/{priceId}/archive:
    $ref: ./paths/admin_prices_{priceId}_archive.yaml
  /billing/webhooks/stripe:
    $ref: ./paths/webhooks_stripe.yaml

components:
  securitySchemes:
    OIDC:
      $ref: ../common/components/securitySchemes/OIDC.yaml
    Oauth2:
      $ref: ../common/components/securitySchemes/Oauth2.yaml

externalDocs:
  url: https://notopia-uit.github.io/notopia
```

Parameters (`components/parameters/`):

```yaml
# workspaceIdPath.yaml
# vim: ft=yaml.openapi
name: workspaceId
in: path
required: true
schema:
  type: string
  format: uuid
```

```yaml
# priceIdPath.yaml
# vim: ft=yaml.openapi
name: priceId
in: path
required: true
schema:
  type: string
  format: uuid
```

```yaml
# tierPath.yaml
# vim: ft=yaml.openapi
name: tier
in: path
required: true
schema:
  $ref: ../schemas/PlanTier.yaml
```

Schemas (`components/schemas/`), one file each, header `# vim: ft=yaml.openapi`, every scalar property with an `example` (Redocly rule `scalar-property-missing-example`):

```yaml
# PlanTier.yaml
type: string
enum: [free, plus, pro]
example: plus
```

```yaml
# BillingInterval.yaml
type: string
enum: [month, year, lifetime]
example: month
```

```yaml
# SubscriptionStatus.yaml
type: string
description: '`none` when the workspace never subscribed.'
enum: [none, incomplete, active, past_due, unpaid, canceled, incomplete_expired]
example: active
```

```yaml
# Features.yaml
type: object
description: Feature flags by key (auto_commit, workspace_graph).
additionalProperties:
  type: boolean
example:
  auto_commit: true
  workspace_graph: true
```

```yaml
# Limits.yaml
type: object
description: Limits by key. -1 means unlimited.
additionalProperties:
  type: integer
  format: int64
example:
  members: 10
  revision_history_days: 30
```

```yaml
# Entitlements.yaml
type: object
properties:
  workspaceId:
    type: string
    format: uuid
    example: 00000000-0000-4000-8000-000000000110
  plan:
    $ref: ./PlanTier.yaml
  status:
    $ref: ./SubscriptionStatus.yaml
  accessUntil:
    type: string
    format: date-time
    nullable: true
    example: 2026-11-11T00:00:00Z
  features:
    $ref: ./Features.yaml
  limits:
    $ref: ./Limits.yaml
required: [workspaceId, plan, status, features, limits]
```

```yaml
# Price.yaml
type: object
properties:
  id:
    type: string
    format: uuid
    example: 0192f0a0-0000-7000-8000-000000000001
  plan:
    $ref: ./PlanTier.yaml
  interval:
    $ref: ./BillingInterval.yaml
  amount:
    type: integer
    format: int64
    example: 49000
  currency:
    type: string
    example: vnd
  lookupKey:
    type: string
    example: plus_month
  active:
    type: boolean
    example: true
  createdAt:
    type: string
    format: date-time
    example: 2026-10-11T00:00:00Z
  archivedAt:
    type: string
    format: date-time
    nullable: true
    example: 2026-10-12T00:00:00Z
required: [id, plan, interval, amount, currency, lookupKey, active, createdAt]
```

```yaml
# Plan.yaml
type: object
properties:
  tier:
    $ref: ./PlanTier.yaml
  features:
    $ref: ./Features.yaml
  limits:
    $ref: ./Limits.yaml
  prices:
    type: array
    items:
      $ref: ./Price.yaml
required: [tier, features, limits, prices]
```

```yaml
# Refund.yaml
type: object
properties:
  id:
    type: string
    format: uuid
    example: 0192f0a0-0000-7000-8000-000000000002
  amount:
    type: integer
    format: int64
    example: 24500
  currency:
    type: string
    example: vnd
  reason:
    type: string
    enum:
      [
        cancel_now,
        lifetime_window,
        replaced_by_lifetime,
        workspace_deleted,
        duplicate_checkout,
        external,
      ]
    example: cancel_now
  status:
    type: string
    enum: [pending, succeeded, failed, canceled]
    example: succeeded
  createdAt:
    type: string
    format: date-time
    example: 2026-10-11T00:00:00Z
required: [id, amount, currency, reason, status, createdAt]
```

```yaml
# Payer.yaml
type: object
properties:
  id:
    type: string
    example: '112'
  email:
    type: string
    example: owner@notopia.com
  name:
    type: string
    example: Kevin Nitro
  missingSince:
    type: string
    format: date-time
    nullable: true
    example: 2026-10-11T00:00:00Z
required: [id, email, name]
```

```yaml
# Subscription.yaml
type: object
properties:
  id:
    type: string
    format: uuid
    example: 0192f0a0-0000-7000-8000-000000000003
  plan:
    $ref: ./PlanTier.yaml
  interval:
    $ref: ./BillingInterval.yaml
  status:
    $ref: ./SubscriptionStatus.yaml
  currentPeriodStart:
    type: string
    format: date-time
    nullable: true
    example: 2026-10-11T00:00:00Z
  currentPeriodEnd:
    type: string
    format: date-time
    nullable: true
    example: 2026-11-11T00:00:00Z
  cancelAtPeriodEnd:
    type: boolean
    example: false
  canceledAt:
    type: string
    format: date-time
    nullable: true
    example: 2026-10-20T00:00:00Z
  accessUntil:
    type: string
    format: date-time
    nullable: true
    example: 2026-11-11T00:00:00Z
  refunds:
    type: array
    items:
      $ref: ./Refund.yaml
required: [id, plan, interval, status, cancelAtPeriodEnd, refunds]
```

```yaml
# WorkspaceBilling.yaml
type: object
properties:
  subscription:
    allOf:
      - $ref: ./Subscription.yaml
    nullable: true
  payer:
    allOf:
      - $ref: ./Payer.yaml
    nullable: true
```

```yaml
# RefundQuote.yaml
type: object
properties:
  amount:
    type: integer
    format: int64
    example: 24500
  currency:
    type: string
    example: vnd
  unusedSeconds:
    type: integer
    format: int64
    example: 1296000
  periodSeconds:
    type: integer
    format: int64
    example: 2592000
  payments:
    type: array
    description: One line per paid invoice payment of the current period
    items:
      $ref: ./RefundQuotePayment.yaml
required: [amount, currency, unusedSeconds, periodSeconds, payments]
```

```yaml
# RefundQuotePayment.yaml
type: object
properties:
  paidAmount:
    type: integer
    format: int64
    example: 49000
  amount:
    type: integer
    format: int64
    description: Refund for this payment
    example: 24500
  coverageSeconds:
    type: integer
    format: int64
    example: 2592000
  unusedSeconds:
    type: integer
    format: int64
    example: 1296000
required: [paidAmount, amount, coverageSeconds, unusedSeconds]
```

```yaml
# RedirectURL.yaml
type: object
properties:
  url:
    type: string
    format: uri
    example: https://checkout.stripe.com/c/pay/cs_test_123
required: [url]
```

```yaml
# CheckoutSessionRequest.yaml
type: object
properties:
  priceId:
    type: string
    format: uuid
    example: 0192f0a0-0000-7000-8000-000000000001
required: [priceId]
```

```yaml
# UpdatePlanRequest.yaml
type: object
properties:
  features:
    $ref: ./Features.yaml
  limits:
    $ref: ./Limits.yaml
```

```yaml
# BillingPolicy.yaml
type: object
properties:
  pastDueGraceDays:
    type: integer
    minimum: 0
    example: 3
  lifetimeRefundWindowDays:
    type: integer
    minimum: 0
    example: 14
required: [pastDueGraceDays, lifetimeRefundWindowDays]
```

```yaml
# UpdatePolicyRequest.yaml
type: object
properties:
  pastDueGraceDays:
    type: integer
    minimum: 0
    example: 3
  lifetimeRefundWindowDays:
    type: integer
    minimum: 0
    example: 14
```

```yaml
# CreatePriceRequest.yaml
type: object
properties:
  plan:
    type: string
    enum: [plus, pro]
    example: plus
  interval:
    $ref: ./BillingInterval.yaml
  amount:
    type: integer
    format: int64
    minimum: 1
    example: 49000
  currency:
    type: string
    example: vnd
required: [plan, interval, amount, currency]
```

Path files (`paths/`). Authenticated operations use the same `security` block as `api/note/paths/workspaces_{workspaceId}_publish.yaml`; public ones use `security: []`. Every operation lists `400`, `401` (authenticated only), `500` from `../../common/components/responses/`; add `403` / `404` where stated.

| File                                                            | Method      | operationId                            | Tag              | Request                                                                                | 2xx                         | Extra errors  |
| --------------------------------------------------------------- | ----------- | -------------------------------------- | ---------------- | -------------------------------------------------------------------------------------- | --------------------------- | ------------- |
| `plans.yaml`                                                    | get         | `listPlans`                            | Plan             | —                                                                                      | 200 `Plan[]`                | — (public)    |
| `workspaces_{workspaceId}_entitlements.yaml`                    | get         | `getWorkspaceEntitlements`             | WorkspaceBilling | —                                                                                      | 200 `Entitlements`          | 403           |
| `workspaces_{workspaceId}_subscription.yaml`                    | get         | `getWorkspaceBilling`                  | WorkspaceBilling | —                                                                                      | 200 `WorkspaceBilling`      | 403           |
| `workspaces_{workspaceId}_checkout-session.yaml`                | post        | `createCheckoutSession`                | WorkspaceBilling | `CheckoutSessionRequest`                                                               | 200 `RedirectURL`           | 403, 404, 409 |
| `workspaces_{workspaceId}_portal-session.yaml`                  | post        | `createPortalSession`                  | WorkspaceBilling | —                                                                                      | 200 `RedirectURL`           | 403, 404      |
| `workspaces_{workspaceId}_subscription_cancel-now_preview.yaml` | get         | `previewCancelNow`                     | WorkspaceBilling | —                                                                                      | 200 `RefundQuote`           | 403, 404, 409 |
| `workspaces_{workspaceId}_subscription_cancel-now.yaml`         | post        | `cancelNow`                            | WorkspaceBilling | —                                                                                      | 200 `WorkspaceBilling`      | 403, 404, 409 |
| `workspaces_{workspaceId}_payer_claim.yaml`                     | post        | `claimPayer`                           | WorkspaceBilling | —                                                                                      | 200 `RedirectURL`           | 403, 404, 409 |
| `admin_plans.yaml`                                              | get         | `adminListPlans`                       | BillingAdmin     | —                                                                                      | 200 `Plan[]`                | 403           |
| `admin_plans_{tier}.yaml`                                       | patch       | `adminUpdatePlan`                      | BillingAdmin     | `UpdatePlanRequest`                                                                    | 200 `Plan`                  | 403, 404      |
| `admin_policy.yaml`                                             | get / patch | `adminGetPolicy` / `adminUpdatePolicy` | BillingAdmin     | — / `UpdatePolicyRequest`                                                              | 200 `BillingPolicy`         | 403           |
| `admin_prices.yaml`                                             | get / post  | `adminListPrices` / `adminCreatePrice` | BillingAdmin     | — / `CreatePriceRequest`                                                               | 200 `Price[]` / 201 `Price` | 403           |
| `admin_prices_{priceId}_archive.yaml`                           | post        | `adminArchivePrice`                    | BillingAdmin     | —                                                                                      | 200 `Price`                 | 403, 404      |
| `webhooks_stripe.yaml`                                          | post        | `stripeWebhook`                        | Webhook          | header `Stripe-Signature` (required, string); body `application/json` free-form object | 204                         | 400 (public)  |

409 responses reuse `../../common/components/responses/BadRequestError.yaml` shape inline (schema `../../common/components/schemas/Error.yaml`, description `Conflict with current billing state`).

Example (`workspaces_{workspaceId}_checkout-session.yaml`):

```yaml
# vim: ft=yaml.openapi
post:
  security:
    - Oauth2:
        - openid
    - OIDC:
        - openid
  tags:
    - WorkspaceBilling
  summary: Start Stripe Checkout for a plan price (owner only)
  operationId: createCheckoutSession
  parameters:
    - $ref: ../components/parameters/workspaceIdPath.yaml
  requestBody:
    required: true
    content:
      application/json:
        schema:
          $ref: ../components/schemas/CheckoutSessionRequest.yaml
  responses:
    200:
      description: Redirect the browser to this URL
      content:
        application/json:
          schema:
            $ref: ../components/schemas/RedirectURL.yaml
    400:
      $ref: ../../common/components/responses/BadRequestError.yaml
    401:
      $ref: ../../common/components/responses/UnauthorizedError.yaml
    403:
      $ref: ../../common/components/responses/ForbiddenError.yaml
    404:
      $ref: ../../common/components/responses/NotFoundError.yaml
    409:
      description: Conflict with current billing state
      content:
        application/json:
          schema:
            $ref: ../../common/components/schemas/Error.yaml
    500:
      $ref: ../../common/components/responses/InternalServerError.yaml
```

Webhook (`webhooks_stripe.yaml`):

```yaml
# vim: ft=yaml.openapi
post:
  security: []
  tags:
    - Webhook
  summary: Stripe webhook endpoint (signature verified against the raw body)
  operationId: stripeWebhook
  parameters:
    - name: Stripe-Signature
      in: header
      required: true
      schema:
        type: string
        example: t=1700000000,v1=abc
  requestBody:
    required: true
    content:
      application/json:
        schema:
          type: object
          additionalProperties: true
  responses:
    204:
      description: Event accepted
    400:
      $ref: ../../common/components/responses/BadRequestError.yaml
    500:
      $ref: ../../common/components/responses/InternalServerError.yaml
```

- [ ] **Step 3: Codegen config + registration**

`api/billing/oapi-codegen-server.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: billing
generate:
  gin-server: true
  strict-server: true
output: ../../pkg/api/billing/server.gen.go
output-options:
  nullable-type: true
```

`api/billing/oapi-codegen-model.yaml`: same with `generate: { models: true }`, `output: ../../pkg/api/billing/model.gen.go`, plus `compatibility: { always-prefix-enum-values: true }` so enum consts are always `<Type><Value>` (`billing.PlanTierFree`, `billing.SubscriptionStatusCanceled`, `billing.RefundStatusCanceled`, `billing.CreatePriceRequestPlanPlus`); without it oapi-codegen prefixes only colliding values (`canceled`, `plus`, `pro` collide here) and the HTTP mappers would not compile predictably.

`api/billing/project.json`: copy `api/note/project.json`, replace every `note` with `billing`, ports `9081`→`9084`, `8081`→`8084`.

`api/redocly.yaml` `apis`: add

```yaml
billing:
  root: ./billing/openapi.yaml
  output: ./bundled/billing.json
```

`api/project.json` `implicitDependencies`: add `"api/billing"`. `api/package.json` `redocly:join`: insert `./bundled/billing.json` after `./bundled/note.json`. `api/api.go`: `BillingSpec Spec = "billing.json"`. `apps/api-web/project.json`: add `api/billing` wherever `api/note` appears.

- [ ] **Step 4: Generate and lint**

```bash
nx lint api && nx run api:gen
ls pkg/api/billing   # model.gen.go server.gen.go
```

- [ ] **Step 5: Commit**

```bash
git add proto/billing api pkg/api/billing pkg/pb packages/pb packages/api-gen apps/api-web/project.json
git commit -m "feat(billing): add billing OpenAPI and proto contracts"
```

---

### Task 2: Domain

**Files:**

- Create: `internal/billing/domain/{types.go,event.go,plan.go,policy.go,price.go,billingaccount.go,subscription.go,refund.go,entitlement.go,repo.go,wire.go}`
- Create: `internal/billing/domain/subscription_test.go`
- Create: `internal/billing/errs/errs.go`

**Interfaces:**

- Produces (used by every later task):
  - `uint8` enums `PlanTier` (zero `PlanTierUnspecified`), `Interval` (`IntervalUnspecified`), `Status` (`StatusNone`), `RefundReason` (`RefundReasonUnspecified`), `RefundStatus` (`RefundStatusUnspecified`), `FeatureKey` (`FeatureKeyUnspecified`), `LimitKey` (`LimitKeyUnspecified`), each with `IsValid()` (false for the Unspecified zero) and `String()`; `ParseFeatureKey(string) (FeatureKey, error)`, `ParseLimitKey(string) (LimitKey, error)`; `Money{Amount int64; Currency string}`; `Payer{ID, Email, Name string}`.
  - `ProviderSnapshot`, `PaymentInfo`, `RefundQuote`, `Entitlements`.
  - aggregates `Plan`, `BillingPolicy`, `Price`, `BillingAccount`, `Subscription` (+ `Refund`) with the methods below.
  - `ResolveEntitlements(workspaceID uuid.UUID, sub *Subscription, plans map[PlanTier]*Plan, policy *BillingPolicy, now time.Time) Entitlements`
  - repo interfaces + `RepoRegistry` + `UnitOfWork` + `ProviderEventLog`.

- [ ] **Step 1: Errors**

```go
// internal/billing/errs/errs.go
package errs

import "fmt"

type Code string

func (c Code) String() string { return string(c) }

const (
	CodeUnauthorized              Code = "unauthorized"
	CodeForbidden                 Code = "forbidden"
	CodeInvalid                   Code = "invalid"
	CodeNotFound                  Code = "notFound"
	CodeInternal                  Code = "internal"
	CodeSubscriptionAlreadyActive Code = "subscriptionAlreadyActive"
	CodeNoBillingAccount          Code = "noBillingAccount"
	CodeNoActiveSubscription      Code = "noActiveSubscription"
	CodeRefundWindowExpired       Code = "refundWindowExpired"
	CodeAlreadyPayer              Code = "alreadyPayer"
	CodeNotOwner                  Code = "notOwner"
	CodePriceNotActive            Code = "priceNotActive"
	CodePaymentProvider           Code = "paymentProviderError"
	CodeAuthorizationService      Code = "authorizationServiceError"
	CodeWebhookSignature          Code = "webhookSignatureInvalid"
)

type Error interface {
	error
	Code() Code
	Message() string
	Unwrap() error
}

type Err struct {
	code    Code
	message string
	err     error
}

var _ Error = (*Err)(nil)

func New(code Code, message string, err error) *Err {
	return &Err{code: code, message: message, err: err}
}

func (e *Err) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *Err) Unwrap() error   { return e.err }
func (e *Err) Code() Code      { return e.code }
func (e *Err) Message() string { return e.message }

func NewInvalid(msg string) *Err   { return New(CodeInvalid, msg, nil) }
func NewForbidden(msg string) *Err { return New(CodeForbidden, msg, nil) }
func NewNotFound(msg string) *Err  { return New(CodeNotFound, msg, nil) }
func NewInternal(msg string, err error) *Err {
	return New(CodeInternal, msg, err)
}
```

- [ ] **Step 2: Types and events**

```go
// internal/billing/domain/types.go
package domain

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

var PlanTiers = []PlanTier{PlanTierFree, PlanTierPlus, PlanTierPro}

func (t PlanTier) IsValid() bool {
	switch t {
	case PlanTierFree, PlanTierPlus, PlanTierPro:
		return true
	default:
		return false
	}
}

func (t PlanTier) String() string {
	switch t {
	case PlanTierUnspecified:
		return "unspecified"
	case PlanTierFree:
		return "free"
	case PlanTierPlus:
		return "plus"
	case PlanTierPro:
		return "pro"
	default:
		return "unknown"
	}
}

type Interval uint8

const (
	IntervalUnspecified Interval = iota
	IntervalMonth
	IntervalYear
	IntervalLifetime
)

func (i Interval) IsValid() bool {
	switch i {
	case IntervalMonth, IntervalYear, IntervalLifetime:
		return true
	default:
		return false
	}
}

func (i Interval) String() string {
	switch i {
	case IntervalUnspecified:
		return "unspecified"
	case IntervalMonth:
		return "month"
	case IntervalYear:
		return "year"
	case IntervalLifetime:
		return "lifetime"
	default:
		return "unknown"
	}
}

// StatusNone is the zero value and a real state: the workspace has no subscription.
type Status uint8

const (
	StatusNone Status = iota
	StatusIncomplete
	StatusActive
	StatusPastDue
	StatusUnpaid
	StatusCanceled
	StatusIncompleteExpired
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNone, StatusIncomplete, StatusActive, StatusPastDue, StatusUnpaid, StatusCanceled, StatusIncompleteExpired:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	switch s {
	case StatusNone:
		return "none"
	case StatusIncomplete:
		return "incomplete"
	case StatusActive:
		return "active"
	case StatusPastDue:
		return "past_due"
	case StatusUnpaid:
		return "unpaid"
	case StatusCanceled:
		return "canceled"
	case StatusIncompleteExpired:
		return "incomplete_expired"
	default:
		return "unknown"
	}
}

type RefundReason uint8

const (
	RefundReasonUnspecified RefundReason = iota
	RefundReasonCancelNow
	RefundReasonLifetimeWindow
	RefundReasonReplacedByLifetime
	RefundReasonWorkspaceDeleted
	RefundReasonDuplicateCheckout
	RefundReasonExternal
)

func (r RefundReason) IsValid() bool {
	switch r {
	case RefundReasonCancelNow, RefundReasonLifetimeWindow, RefundReasonReplacedByLifetime,
		RefundReasonWorkspaceDeleted, RefundReasonDuplicateCheckout, RefundReasonExternal:
		return true
	default:
		return false
	}
}

func (r RefundReason) String() string {
	switch r {
	case RefundReasonUnspecified:
		return "unspecified"
	case RefundReasonCancelNow:
		return "cancel_now"
	case RefundReasonLifetimeWindow:
		return "lifetime_window"
	case RefundReasonReplacedByLifetime:
		return "replaced_by_lifetime"
	case RefundReasonWorkspaceDeleted:
		return "workspace_deleted"
	case RefundReasonDuplicateCheckout:
		return "duplicate_checkout"
	case RefundReasonExternal:
		return "external"
	default:
		return "unknown"
	}
}

type RefundStatus uint8

const (
	RefundStatusUnspecified RefundStatus = iota
	RefundStatusPending
	RefundStatusSucceeded
	RefundStatusFailed
	RefundStatusCanceled
)

func (s RefundStatus) IsValid() bool {
	switch s {
	case RefundStatusPending, RefundStatusSucceeded, RefundStatusFailed, RefundStatusCanceled:
		return true
	default:
		return false
	}
}

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusUnspecified:
		return "unspecified"
	case RefundStatusPending:
		return "pending"
	case RefundStatusSucceeded:
		return "succeeded"
	case RefundStatusFailed:
		return "failed"
	case RefundStatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type FeatureKey uint8

const (
	FeatureKeyUnspecified FeatureKey = iota
	FeatureKeyAutoCommit
	FeatureKeyWorkspaceGraph
)

var FeatureKeys = []FeatureKey{FeatureKeyAutoCommit, FeatureKeyWorkspaceGraph}

func (k FeatureKey) IsValid() bool {
	switch k {
	case FeatureKeyAutoCommit, FeatureKeyWorkspaceGraph:
		return true
	default:
		return false
	}
}

func (k FeatureKey) String() string {
	switch k {
	case FeatureKeyUnspecified:
		return "unspecified"
	case FeatureKeyAutoCommit:
		return "auto_commit"
	case FeatureKeyWorkspaceGraph:
		return "workspace_graph"
	default:
		return "unknown"
	}
}

func ParseFeatureKey(s string) (FeatureKey, error) {
	for _, k := range FeatureKeys {
		if k.String() == s {
			return k, nil
		}
	}
	return FeatureKeyUnspecified, errs.NewInvalid(fmt.Sprintf("unknown feature %q", s))
}

type LimitKey uint8

const (
	LimitKeyUnspecified LimitKey = iota
	LimitKeyMembers
	LimitKeyPublishedNotes
	LimitKeyNoteGraphMaxDepth
	LimitKeyAutoCommitMinInterval
	LimitKeyRevisionHistoryDays
	LimitKeyTrashRetentionDays
	LimitKeyAICreditsMonthly
)

var LimitKeys = []LimitKey{
	LimitKeyMembers, LimitKeyPublishedNotes, LimitKeyNoteGraphMaxDepth, LimitKeyAutoCommitMinInterval,
	LimitKeyRevisionHistoryDays, LimitKeyTrashRetentionDays, LimitKeyAICreditsMonthly,
}

func (k LimitKey) IsValid() bool {
	switch k {
	case LimitKeyMembers, LimitKeyPublishedNotes, LimitKeyNoteGraphMaxDepth, LimitKeyAutoCommitMinInterval,
		LimitKeyRevisionHistoryDays, LimitKeyTrashRetentionDays, LimitKeyAICreditsMonthly:
		return true
	default:
		return false
	}
}

func (k LimitKey) String() string {
	switch k {
	case LimitKeyUnspecified:
		return "unspecified"
	case LimitKeyMembers:
		return "members"
	case LimitKeyPublishedNotes:
		return "published_notes"
	case LimitKeyNoteGraphMaxDepth:
		return "note_graph_max_depth"
	case LimitKeyAutoCommitMinInterval:
		return "auto_commit_min_interval_seconds"
	case LimitKeyRevisionHistoryDays:
		return "revision_history_days"
	case LimitKeyTrashRetentionDays:
		return "trash_retention_days"
	case LimitKeyAICreditsMonthly:
		return "ai_credits_monthly"
	default:
		return "unknown"
	}
}

func ParseLimitKey(s string) (LimitKey, error) {
	for _, k := range LimitKeys {
		if k.String() == s {
			return k, nil
		}
	}
	return LimitKeyUnspecified, errs.NewInvalid(fmt.Sprintf("unknown limit %q", s))
}

const Unlimited int64 = -1

type Money struct {
	Amount   int64
	Currency string
}

type Payer struct {
	ID    string
	Email string
	Name  string
}
```

```go
// internal/billing/domain/event.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	GetID() uuid.UUID
	GetAggregateID() string
	GetUserID() string
}

type BaseEvent struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	AggregateID string
	UserID      string
}

func NewBaseEvent(aggregateID, userID string) BaseEvent {
	return BaseEvent{ID: uuid.Must(uuid.NewV7()), OccurredAt: time.Now(), AggregateID: aggregateID, UserID: userID}
}

func (e *BaseEvent) GetID() uuid.UUID       { return e.ID }
func (e *BaseEvent) GetAggregateID() string { return e.AggregateID }
func (e *BaseEvent) GetUserID() string      { return e.UserID }

type eventRecorder struct{ events []Event }

func (r *eventRecorder) record(e Event) { r.events = append(r.events, e) }

func (r *eventRecorder) PopEvents() []Event {
	events := r.events
	r.events = nil
	return events
}

type (
	PlanUpdatedEvent struct {
		BaseEvent
		Features map[FeatureKey]bool
		Limits   map[LimitKey]int64
	}
	BillingPolicyUpdatedEvent       struct{ BaseEvent }
	PortalConfigurationChangedEvent struct {
		BaseEvent
		ConfigurationID string
	}
	PriceCreatedEvent  struct{ BaseEvent }
	PriceArchivedEvent struct{ BaseEvent }

	BillingAccountOpenedEvent  struct{ BaseEvent }
	PayerChangedEvent          struct {
		BaseEvent
		PayerID string
	}
	PayerMissingEvent          struct{ BaseEvent }
	BillingContactUpdatedEvent struct{ BaseEvent }

	SubscriptionCreatedEvent struct {
		BaseEvent
		Plan     PlanTier
		Interval Interval
	}
	SubscriptionActivatedEvent      struct{ BaseEvent }
	SubscriptionRenewedEvent        struct{ BaseEvent }
	SubscriptionPastDueEvent        struct{ BaseEvent }
	SubscriptionPlanChangedEvent    struct {
		BaseEvent
		From, To PlanTier
	}
	SubscriptionCancelScheduledEvent struct{ BaseEvent }
	SubscriptionCancelResumedEvent   struct{ BaseEvent }
	SubscriptionCanceledNowEvent     struct {
		BaseEvent
		Reason RefundReason
	}
	SubscriptionEndedEvent struct{ BaseEvent }
	RefundRequestedEvent   struct {
		BaseEvent
		RefundID uuid.UUID
		Amount   Money
	}
	RefundSucceededEvent struct {
		BaseEvent
		RefundID uuid.UUID
	}
	RefundFailedEvent struct {
		BaseEvent
		RefundID uuid.UUID
	}
)
```

- [ ] **Step 3: Plan, policy, price, billing account**

```go
// internal/billing/domain/plan.go
package domain

import (
	"fmt"
	"maps"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Plan struct {
	eventRecorder
	tier      PlanTier
	features  map[FeatureKey]bool
	limits    map[LimitKey]int64
	updatedAt time.Time
}

func NewPlan(tier PlanTier, features map[FeatureKey]bool, limits map[LimitKey]int64, now time.Time) (*Plan, error) {
	p := &Plan{tier: tier, features: map[FeatureKey]bool{}, limits: map[LimitKey]int64{}, updatedAt: now}
	if !tier.IsValid() {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid tier %q", tier))
	}
	if err := p.apply(features, limits); err != nil {
		return nil, err
	}
	for _, k := range FeatureKeys {
		if _, ok := p.features[k]; !ok {
			p.features[k] = false
		}
	}
	for _, k := range LimitKeys {
		if _, ok := p.limits[k]; !ok {
			p.limits[k] = 0
		}
	}
	return p, nil
}

func UnmarshalPlan(tier PlanTier, features map[FeatureKey]bool, limits map[LimitKey]int64, updatedAt time.Time) *Plan {
	return &Plan{tier: tier, features: features, limits: limits, updatedAt: updatedAt}
}

func (p *Plan) apply(features map[FeatureKey]bool, limits map[LimitKey]int64) error {
	for k, v := range features {
		if !k.IsValid() {
			return errs.NewInvalid(fmt.Sprintf("unknown feature %q", k))
		}
		p.features[k] = v
	}
	for k, v := range limits {
		if !k.IsValid() {
			return errs.NewInvalid(fmt.Sprintf("unknown limit %q", k))
		}
		if v < Unlimited {
			return errs.NewInvalid(fmt.Sprintf("limit %q must be >= -1", k))
		}
		p.limits[k] = v
	}
	return nil
}

func (p *Plan) Update(features map[FeatureKey]bool, limits map[LimitKey]int64, now time.Time, userID string) error {
	if err := p.apply(features, limits); err != nil {
		return err
	}
	p.updatedAt = now
	p.record(&PlanUpdatedEvent{BaseEvent: NewBaseEvent(p.tier.String(), userID), Features: features, Limits: limits})
	return nil
}

func (p *Plan) Tier() PlanTier                   { return p.tier }
func (p *Plan) Feature(k FeatureKey) bool         { return p.features[k] }
func (p *Plan) Limit(k LimitKey) int64            { return p.limits[k] }
func (p *Plan) Features() map[FeatureKey]bool     { return maps.Clone(p.features) }
func (p *Plan) Limits() map[LimitKey]int64        { return maps.Clone(p.limits) }
func (p *Plan) UpdatedAt() time.Time              { return p.updatedAt }

func DefaultPlans(now time.Time) []*Plan {
	mk := func(t PlanTier, autoCommit, graph bool, members, published, depth, interval, revisions, trash, ai int64) *Plan {
		p, _ := NewPlan(t,
			map[FeatureKey]bool{FeatureKeyAutoCommit: autoCommit, FeatureKeyWorkspaceGraph: graph},
			map[LimitKey]int64{
				LimitKeyMembers: members, LimitKeyPublishedNotes: published, LimitKeyNoteGraphMaxDepth: depth,
				LimitKeyAutoCommitMinInterval: interval, LimitKeyRevisionHistoryDays: revisions,
				LimitKeyTrashRetentionDays: trash, LimitKeyAICreditsMonthly: ai,
			}, now)
		return p
	}
	return []*Plan{
		mk(PlanTierFree, false, false, 3, 3, 5, 0, 7, 7, 0),
		mk(PlanTierPlus, true, true, 10, Unlimited, Unlimited, 300, 30, 30, 200),
		mk(PlanTierPro, true, true, Unlimited, Unlimited, Unlimited, 60, Unlimited, 90, 2000),
	}
}
```

```go
// internal/billing/domain/policy.go
package domain

import (
	"time"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type BillingPolicy struct {
	eventRecorder
	pastDueGraceDays         int
	lifetimeRefundWindowDays int
	portalConfigurationID    *string
	updatedAt                time.Time
}

func DefaultBillingPolicy(now time.Time) *BillingPolicy {
	return &BillingPolicy{pastDueGraceDays: 3, lifetimeRefundWindowDays: 14, updatedAt: now}
}

func UnmarshalBillingPolicy(grace, window int, portalConfigurationID *string, updatedAt time.Time) *BillingPolicy {
	return &BillingPolicy{pastDueGraceDays: grace, lifetimeRefundWindowDays: window, portalConfigurationID: portalConfigurationID, updatedAt: updatedAt}
}

func (p *BillingPolicy) Update(grace, window *int, now time.Time, userID string) error {
	if (grace != nil && *grace < 0) || (window != nil && *window < 0) {
		return errs.NewInvalid("policy values must be >= 0")
	}
	if grace != nil {
		p.pastDueGraceDays = *grace
	}
	if window != nil {
		p.lifetimeRefundWindowDays = *window
	}
	p.updatedAt = now
	p.record(&BillingPolicyUpdatedEvent{BaseEvent: NewBaseEvent("policy", userID)})
	return nil
}

func (p *BillingPolicy) SetPortalConfiguration(id string, now time.Time, userID string) {
	p.portalConfigurationID = &id
	p.updatedAt = now
	p.record(&PortalConfigurationChangedEvent{BaseEvent: NewBaseEvent("policy", userID), ConfigurationID: id})
}

func (p *BillingPolicy) PastDueGrace() time.Duration {
	return time.Duration(p.pastDueGraceDays) * 24 * time.Hour
}

func (p *BillingPolicy) LifetimeRefundWindow() time.Duration {
	return time.Duration(p.lifetimeRefundWindowDays) * 24 * time.Hour
}

func (p *BillingPolicy) PastDueGraceDays() int           { return p.pastDueGraceDays }
func (p *BillingPolicy) LifetimeRefundWindowDays() int   { return p.lifetimeRefundWindowDays }
func (p *BillingPolicy) PortalConfigurationID() *string  { return p.portalConfigurationID }
func (p *BillingPolicy) UpdatedAt() time.Time            { return p.updatedAt }
```

```go
// internal/billing/domain/price.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Price struct {
	eventRecorder
	id              uuid.UUID
	plan            PlanTier
	interval        Interval
	amount          Money
	providerPriceID string
	lookupKey       string
	active          bool
	createdAt       time.Time
	archivedAt      *time.Time
}

func LookupKey(plan PlanTier, interval Interval) string {
	return plan.String() + "_" + interval.String()
}

func NewPrice(id uuid.UUID, plan PlanTier, interval Interval, amount Money, providerPriceID string, now time.Time, userID string) (*Price, error) {
	if !plan.IsValid() || plan == PlanTierFree {
		return nil, errs.NewInvalid("price plan must be plus or pro")
	}
	if !interval.IsValid() {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid interval %q", interval))
	}
	if amount.Amount <= 0 || amount.Currency == "" {
		return nil, errs.NewInvalid("price amount must be > 0 with a currency")
	}
	p := &Price{
		id: id, plan: plan, interval: interval, amount: amount, providerPriceID: providerPriceID,
		lookupKey: LookupKey(plan, interval), active: true, createdAt: now,
	}
	p.record(&PriceCreatedEvent{BaseEvent: NewBaseEvent(id.String(), userID)})
	return p, nil
}

type UnmarshalPriceParams struct {
	ID              uuid.UUID
	Plan            PlanTier
	Interval        Interval
	Amount          Money
	ProviderPriceID string
	LookupKey       string
	Active          bool
	CreatedAt       time.Time
	ArchivedAt      *time.Time
}

func UnmarshalPrice(p UnmarshalPriceParams) *Price {
	return &Price{
		id: p.ID, plan: p.Plan, interval: p.Interval, amount: p.Amount, providerPriceID: p.ProviderPriceID,
		lookupKey: p.LookupKey, active: p.Active, createdAt: p.CreatedAt, archivedAt: p.ArchivedAt,
	}
}

func (p *Price) Archive(now time.Time, userID string) error {
	if !p.active {
		return nil
	}
	p.active = false
	p.archivedAt = &now
	p.record(&PriceArchivedEvent{BaseEvent: NewBaseEvent(p.id.String(), userID)})
	return nil
}

func (p *Price) ID() uuid.UUID              { return p.id }
func (p *Price) Plan() PlanTier             { return p.plan }
func (p *Price) Interval() Interval         { return p.interval }
func (p *Price) Amount() Money              { return p.amount }
func (p *Price) ProviderPriceID() string    { return p.providerPriceID }
func (p *Price) LookupKey() string          { return p.lookupKey }
func (p *Price) Active() bool               { return p.active }
func (p *Price) CreatedAt() time.Time       { return p.createdAt }
func (p *Price) ArchivedAt() *time.Time     { return p.archivedAt }
```

```go
// internal/billing/domain/billingaccount.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type BillingAccount struct {
	eventRecorder
	workspaceID        uuid.UUID
	providerCustomerID string
	payer              Payer
	payerMissingSince  *time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

func OpenBillingAccount(workspaceID uuid.UUID, providerCustomerID string, payer Payer, now time.Time, userID string) *BillingAccount {
	a := &BillingAccount{workspaceID: workspaceID, providerCustomerID: providerCustomerID, payer: payer, createdAt: now, updatedAt: now}
	a.record(&BillingAccountOpenedEvent{BaseEvent: NewBaseEvent(workspaceID.String(), userID)})
	return a
}

func UnmarshalBillingAccount(workspaceID uuid.UUID, providerCustomerID string, payer Payer, payerMissingSince *time.Time, createdAt, updatedAt time.Time) *BillingAccount {
	return &BillingAccount{workspaceID: workspaceID, providerCustomerID: providerCustomerID, payer: payer, payerMissingSince: payerMissingSince, createdAt: createdAt, updatedAt: updatedAt}
}

func (a *BillingAccount) ChangePayer(payer Payer, now time.Time, userID string) {
	a.payer = payer
	a.payerMissingSince = nil
	a.updatedAt = now
	a.record(&PayerChangedEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID), PayerID: payer.ID})
}

func (a *BillingAccount) MarkPayerMissing(now time.Time, userID string) {
	if a.payerMissingSince != nil {
		return
	}
	a.payerMissingSince = &now
	a.updatedAt = now
	a.record(&PayerMissingEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID)})
}

func (a *BillingAccount) UpdateContact(email, name string, now time.Time, userID string) {
	a.payer.Email, a.payer.Name = email, name
	a.updatedAt = now
	a.record(&BillingContactUpdatedEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID)})
}

func (a *BillingAccount) WorkspaceID() uuid.UUID        { return a.workspaceID }
func (a *BillingAccount) ProviderCustomerID() string    { return a.providerCustomerID }
func (a *BillingAccount) Payer() Payer                  { return a.payer }
func (a *BillingAccount) PayerMissingSince() *time.Time { return a.payerMissingSince }
func (a *BillingAccount) CreatedAt() time.Time          { return a.createdAt }
func (a *BillingAccount) UpdatedAt() time.Time          { return a.updatedAt }
```

- [ ] **Step 4: Subscription + Refund (write the test first)**

```go
// internal/billing/domain/subscription_test.go
package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

func price(t *testing.T, plan domain.PlanTier, interval domain.Interval, amount int64) *domain.Price {
	t.Helper()
	p, err := domain.NewPrice(uuid.New(), plan, interval, domain.Money{Amount: amount, Currency: "vnd"}, "price_x", t0, "u")
	require.NoError(t, err)
	return p
}

func snap(status domain.Status, start, end time.Time) domain.ProviderSnapshot {
	return domain.ProviderSnapshot{ProviderSubscriptionID: "sub_1", Status: status, CurrentPeriodStart: start, CurrentPeriodEnd: end}
}

func eventNames(events []domain.Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, fmtType(e))
	}
	return out
}

func fmtType(e domain.Event) string {
	switch e.(type) {
	case *domain.SubscriptionCreatedEvent:
		return "created"
	case *domain.SubscriptionActivatedEvent:
		return "activated"
	case *domain.SubscriptionRenewedEvent:
		return "renewed"
	case *domain.SubscriptionPastDueEvent:
		return "past_due"
	case *domain.SubscriptionPlanChangedEvent:
		return "plan_changed"
	case *domain.SubscriptionCancelScheduledEvent:
		return "cancel_scheduled"
	case *domain.SubscriptionEndedEvent:
		return "ended"
	default:
		return "other"
	}
}

func TestGrantsAccess(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 3 days grace
	end := t0.AddDate(0, 1, 0)
	cases := []struct {
		name   string
		status domain.Status
		now    time.Time
		want   bool
	}{
		{"active after stored period end (renewal webhook late)", domain.StatusActive, end.Add(2 * time.Hour), true},
		{"past_due within grace", domain.StatusPastDue, end.Add(48 * time.Hour), true},
		{"past_due after grace", domain.StatusPastDue, end.Add(96 * time.Hour), false},
		{"canceled", domain.StatusCanceled, t0, false},
		{"unpaid", domain.StatusUnpaid, t0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(c.status, t0, end), t0, "u")
			require.NoError(t, err)
			assert.Equal(t, c.want, s.GrantsAccess(c.now, policy))
		})
	}
}

func TestApplySnapshotEvents(t *testing.T) {
	plus := price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000)
	pro := price(t, domain.PlanTierPro, domain.IntervalMonth, 99000)
	end := t0.AddDate(0, 1, 0)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), plus, snap(domain.StatusActive, t0, end), t0, "u")
	assert.Equal(t, []string{"created", "activated"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.StatusActive, end, end.AddDate(0, 1, 0)), plus, end, "u")
	assert.Equal(t, []string{"renewed"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.StatusPastDue, end, end.AddDate(0, 1, 0)), pro, end, "u")
	assert.Equal(t, []string{"past_due", "plan_changed"}, eventNames(s.PopEvents()))

	ended := snap(domain.StatusCanceled, end, end.AddDate(0, 1, 0))
	s.ApplySnapshot(ended, pro, end, "u")
	assert.Equal(t, []string{"ended"}, eventNames(s.PopEvents()))
}

func TestQuoteCancelNow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	end := t0.Add(30 * 24 * time.Hour)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.StatusActive, t0, end), t0, "u")
	pay := domain.PaymentInfo{Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_1", PaidAt: t0, CoverageStart: t0, CoverageEnd: end}

	q, err := s.QuoteCancelNow(t0.Add(10*24*time.Hour), policy, []domain.PaymentInfo{pay})
	require.NoError(t, err)
	assert.EqualValues(t, 20000, q.Amount.Amount)

	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	// A second refund against the same payment must not exceed what was paid.
	over := domain.RefundQuote{Amount: domain.Money{Amount: 20000, Currency: "vnd"}, Lines: []domain.RefundLine{{PaymentRef: "pi_1", PaidAmount: pay.Amount, Amount: domain.Money{Amount: 20000, Currency: "vnd"}}}}
	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, over, "u")
	assert.Error(t, err, "non-current subscription / over-refund")
}

// Plus 30000 paid for day 0-30; upgrade on day 10 invoiced 40000 for day 10-30; cancel on day 20.
func TestQuoteCancelNowAfterUpgrade(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	day := func(n int) time.Time { return t0.Add(time.Duration(n) * 24 * time.Hour) }
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalMonth, 90000), snap(domain.StatusActive, t0, day(30)), t0, "u")
	payments := []domain.PaymentInfo{
		{Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_a", PaidAt: t0, CoverageStart: t0, CoverageEnd: day(30)},
		{Amount: domain.Money{Amount: 40000, Currency: "vnd"}, PaymentRef: "pi_b", PaidAt: day(10), CoverageStart: day(10), CoverageEnd: day(30)},
	}

	q, err := s.QuoteCancelNow(day(20), policy, payments)
	require.NoError(t, err)
	require.Len(t, q.Lines, 2)
	assert.EqualValues(t, 10000, q.Lines[0].Amount.Amount)
	assert.EqualValues(t, 20000, q.Lines[1].Amount.Amount)
	assert.EqualValues(t, 30000, q.Amount.Amount)

	refunds, err := s.CancelNow(day(20), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	require.Len(t, refunds, 2)
	assert.Equal(t, "pi_b", refunds[1].PaymentRef())
}

func TestLifetimeRefundWindow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 14 days
	s, _ := domain.NewLifetimeSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalLifetime, 2490000), "pi_l", t0, t0, "u")
	pay := []domain.PaymentInfo{{Amount: domain.Money{Amount: 2490000, Currency: "vnd"}, PaymentRef: "pi_l", PaidAt: t0}}

	q, err := s.QuoteCancelNow(t0.Add(13*24*time.Hour), policy, pay)
	require.NoError(t, err)
	assert.EqualValues(t, 2490000, q.Amount.Amount)

	_, err = s.QuoteCancelNow(t0.Add(15*24*time.Hour), policy, pay)
	assert.Error(t, err, "refundWindowExpired after 14 days")
	assert.True(t, s.GrantsAccess(t0.AddDate(5, 0, 0), policy), "lifetime grants access forever")
}
```

Run: `go test ./internal/billing/domain/...` → FAIL (types missing).

```go
// internal/billing/domain/refund.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Refund struct {
	id               uuid.UUID
	providerRefundID *string
	paymentRef       string
	amount           Money
	reason           RefundReason
	status           RefundStatus
	failureMessage   *string
	createdAt        time.Time
	updatedAt        time.Time
}

type UnmarshalRefundParams struct {
	ID               uuid.UUID
	ProviderRefundID *string
	PaymentRef       string
	Amount           Money
	Reason           RefundReason
	Status           RefundStatus
	FailureMessage   *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func UnmarshalRefund(p UnmarshalRefundParams) *Refund {
	return &Refund{
		id: p.ID, providerRefundID: p.ProviderRefundID, paymentRef: p.PaymentRef, amount: p.Amount,
		reason: p.Reason, status: p.Status, failureMessage: p.FailureMessage, createdAt: p.CreatedAt, updatedAt: p.UpdatedAt,
	}
}

func (r *Refund) ID() uuid.UUID              { return r.id }
func (r *Refund) ProviderRefundID() *string  { return r.providerRefundID }
func (r *Refund) PaymentRef() string         { return r.paymentRef }
func (r *Refund) Amount() Money              { return r.amount }
func (r *Refund) Reason() RefundReason       { return r.reason }
func (r *Refund) Status() RefundStatus       { return r.status }
func (r *Refund) FailureMessage() *string    { return r.failureMessage }
func (r *Refund) CreatedAt() time.Time       { return r.createdAt }
func (r *Refund) UpdatedAt() time.Time       { return r.updatedAt }
```

```go
// internal/billing/domain/subscription.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type ProviderSnapshot struct {
	ProviderSubscriptionID string
	ProviderPriceID        string
	WorkspaceID            uuid.UUID
	Status                 Status
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
}

// PaymentInfo is one paid invoice payment. CoverageStart/End is the service window it paid
// for (invoice line periods); zero for lifetime.
type PaymentInfo struct {
	Amount        Money
	PaymentRef    string
	PaidAt        time.Time
	CoverageStart time.Time
	CoverageEnd   time.Time
}

type RefundLine struct {
	PaymentRef      string
	PaidAmount      Money
	Amount          Money
	CoverageSeconds int64
	UnusedSeconds   int64
}

// RefundQuote.Amount is the sum of Lines; each line becomes one Refund.
type RefundQuote struct {
	Amount        Money
	UnusedSeconds int64
	PeriodSeconds int64
	Lines         []RefundLine
}

// FullRefundQuote refunds every payment in full (duplicate checkout).
func FullRefundQuote(payments ...PaymentInfo) RefundQuote {
	var q RefundQuote
	for _, p := range payments {
		q.Lines = append(q.Lines, RefundLine{PaymentRef: p.PaymentRef, PaidAmount: p.Amount, Amount: p.Amount})
		q.Amount = Money{Amount: q.Amount.Amount + p.Amount.Amount, Currency: p.Amount.Currency}
	}
	return q
}

type Subscription struct {
	eventRecorder
	id                     uuid.UUID
	workspaceID            uuid.UUID
	plan                   PlanTier
	interval               Interval
	priceID                uuid.UUID
	status                 Status
	providerSubscriptionID *string
	providerPaymentID      *string
	currentPeriodStart     *time.Time
	currentPeriodEnd       *time.Time
	cancelAtPeriodEnd      bool
	canceledAt             *time.Time
	paidAt                 *time.Time
	providerSyncedAt       time.Time
	createdAt              time.Time
	updatedAt              time.Time
	refunds                []*Refund
}

func (s *Subscription) base(userID string) BaseEvent { return NewBaseEvent(s.id.String(), userID) }

func NewSubscription(id, workspaceID uuid.UUID, price *Price, snap ProviderSnapshot, now time.Time, userID string) (*Subscription, error) {
	if price.Interval() == IntervalLifetime {
		return nil, errs.NewInvalid("recurring subscription needs a recurring price")
	}
	providerID := snap.ProviderSubscriptionID
	s := &Subscription{
		id: id, workspaceID: workspaceID, plan: price.Plan(), interval: price.Interval(), priceID: price.ID(),
		status: StatusIncomplete, providerSubscriptionID: &providerID, providerSyncedAt: now, createdAt: now, updatedAt: now,
	}
	s.record(&SubscriptionCreatedEvent{BaseEvent: s.base(userID), Plan: s.plan, Interval: s.interval})
	s.ApplySnapshot(snap, price, now, userID)
	return s, nil
}

func NewLifetimeSubscription(id, workspaceID uuid.UUID, price *Price, providerPaymentID string, paidAt, now time.Time, userID string) (*Subscription, error) {
	if price.Interval() != IntervalLifetime {
		return nil, errs.NewInvalid("lifetime subscription needs a lifetime price")
	}
	s := &Subscription{
		id: id, workspaceID: workspaceID, plan: price.Plan(), interval: IntervalLifetime, priceID: price.ID(),
		status: StatusActive, providerPaymentID: &providerPaymentID, paidAt: &paidAt,
		providerSyncedAt: now, createdAt: now, updatedAt: now,
	}
	s.record(&SubscriptionCreatedEvent{BaseEvent: s.base(userID), Plan: s.plan, Interval: s.interval})
	s.record(&SubscriptionActivatedEvent{BaseEvent: s.base(userID)})
	return s, nil
}

type UnmarshalSubscriptionParams struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	Plan                   PlanTier
	Interval               Interval
	PriceID                uuid.UUID
	Status                 Status
	ProviderSubscriptionID *string
	ProviderPaymentID      *string
	CurrentPeriodStart     *time.Time
	CurrentPeriodEnd       *time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
	PaidAt                 *time.Time
	ProviderSyncedAt       time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Refunds                []*Refund
}

func UnmarshalSubscription(p UnmarshalSubscriptionParams) *Subscription {
	return &Subscription{
		id: p.ID, workspaceID: p.WorkspaceID, plan: p.Plan, interval: p.Interval, priceID: p.PriceID, status: p.Status,
		providerSubscriptionID: p.ProviderSubscriptionID, providerPaymentID: p.ProviderPaymentID,
		currentPeriodStart: p.CurrentPeriodStart, currentPeriodEnd: p.CurrentPeriodEnd,
		cancelAtPeriodEnd: p.CancelAtPeriodEnd, canceledAt: p.CanceledAt, paidAt: p.PaidAt,
		providerSyncedAt: p.ProviderSyncedAt, createdAt: p.CreatedAt, updatedAt: p.UpdatedAt, refunds: p.Refunds,
	}
}

func (s *Subscription) ApplySnapshot(snap ProviderSnapshot, price *Price, now time.Time, userID string) {
	prevStatus, prevPlan, prevCancel := s.status, s.plan, s.cancelAtPeriodEnd
	prevEnd := s.currentPeriodEnd
	isNew := prevStatus == StatusIncomplete && prevEnd == nil

	start, end := snap.CurrentPeriodStart, snap.CurrentPeriodEnd
	s.status = snap.Status
	s.plan, s.interval, s.priceID = price.Plan(), price.Interval(), price.ID()
	s.currentPeriodStart, s.currentPeriodEnd = &start, &end
	s.cancelAtPeriodEnd = snap.CancelAtPeriodEnd
	s.canceledAt = snap.CanceledAt
	s.providerSyncedAt, s.updatedAt = now, now

	switch {
	case s.status == StatusActive && prevStatus != StatusActive:
		s.paidAt = &now
		s.record(&SubscriptionActivatedEvent{BaseEvent: s.base(userID)})
	case s.status == StatusActive && prevEnd != nil && end.After(*prevEnd):
		s.paidAt = &now
		s.record(&SubscriptionRenewedEvent{BaseEvent: s.base(userID)})
	}
	if s.status == StatusPastDue && prevStatus != StatusPastDue {
		s.record(&SubscriptionPastDueEvent{BaseEvent: s.base(userID)})
	}
	if !isNew && prevPlan != s.plan {
		s.record(&SubscriptionPlanChangedEvent{BaseEvent: s.base(userID), From: prevPlan, To: s.plan})
	}
	if !prevCancel && s.cancelAtPeriodEnd {
		s.record(&SubscriptionCancelScheduledEvent{BaseEvent: s.base(userID)})
	}
	if prevCancel && !s.cancelAtPeriodEnd && s.IsCurrent() {
		s.record(&SubscriptionCancelResumedEvent{BaseEvent: s.base(userID)})
	}
	if !s.IsCurrent() && (prevStatus != s.status) && !isNew {
		s.record(&SubscriptionEndedEvent{BaseEvent: s.base(userID)})
	}
}

func (s *Subscription) ScheduleCancel(now time.Time, userID string) error {
	if s.interval == IntervalLifetime {
		return errs.NewInvalid("lifetime subscriptions cannot be scheduled for cancellation")
	}
	if s.cancelAtPeriodEnd {
		return nil
	}
	s.cancelAtPeriodEnd, s.updatedAt = true, now
	s.record(&SubscriptionCancelScheduledEvent{BaseEvent: s.base(userID)})
	return nil
}

func (s *Subscription) ResumeCancel(now time.Time, userID string) {
	if !s.cancelAtPeriodEnd {
		return
	}
	s.cancelAtPeriodEnd, s.updatedAt = false, now
	s.record(&SubscriptionCancelResumedEvent{BaseEvent: s.base(userID)})
}

func (s *Subscription) refundedFor(paymentRef string) int64 {
	var total int64
	for _, r := range s.refunds {
		if r.paymentRef == paymentRef && r.status != RefundStatusFailed && r.status != RefundStatusCanceled {
			total += r.amount.Amount
		}
	}
	return total
}

// QuoteCancelNow quotes each paid payment of the current period on its own coverage window,
// so an immediately invoiced upgrade does not hide the earlier payment of the same period.
func (s *Subscription) QuoteCancelNow(now time.Time, policy *BillingPolicy, payments []PaymentInfo) (RefundQuote, error) {
	if !s.IsCurrent() {
		return RefundQuote{}, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	var q RefundQuote
	add := func(p PaymentInfo, amount, coverage, unused int64) {
		amount = max(min(amount, p.Amount.Amount-s.refundedFor(p.PaymentRef)), 0)
		q.Lines = append(q.Lines, RefundLine{
			PaymentRef: p.PaymentRef, PaidAmount: p.Amount, Amount: Money{Amount: amount, Currency: p.Amount.Currency},
			CoverageSeconds: coverage, UnusedSeconds: unused,
		})
		q.Amount = Money{Amount: q.Amount.Amount + amount, Currency: p.Amount.Currency}
	}
	if s.interval == IntervalLifetime {
		if s.paidAt == nil || now.Sub(*s.paidAt) > policy.LifetimeRefundWindow() {
			return RefundQuote{}, errs.New(errs.CodeRefundWindowExpired, "lifetime refund window expired", nil)
		}
		for _, p := range payments {
			add(p, p.Amount.Amount, 0, 0)
		}
		return q, nil
	}
	if s.currentPeriodStart == nil || s.currentPeriodEnd == nil {
		return RefundQuote{}, errs.NewInternal("subscription has no period", nil)
	}
	q.PeriodSeconds = int64(s.currentPeriodEnd.Sub(*s.currentPeriodStart).Seconds())
	q.UnusedSeconds = max(int64(s.currentPeriodEnd.Sub(now).Seconds()), 0)
	for _, p := range payments {
		start, end := latest(p.CoverageStart, *s.currentPeriodStart), earliest(p.CoverageEnd, *s.currentPeriodEnd)
		coverage := int64(end.Sub(start).Seconds())
		if coverage <= 0 {
			continue
		}
		unused := max(int64(end.Sub(latest(now, start)).Seconds()), 0)
		add(p, p.Amount.Amount*unused/coverage, coverage, unused)
	}
	return q, nil
}

func latest(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// CancelNow adds one pending Refund per quote line with a positive amount.
func (s *Subscription) CancelNow(now time.Time, reason RefundReason, quote RefundQuote, userID string) ([]*Refund, error) {
	if !s.IsCurrent() {
		return nil, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	for _, l := range quote.Lines {
		if l.Amount.Amount > 0 && s.refundedFor(l.PaymentRef)+l.Amount.Amount > l.PaidAmount.Amount {
			return nil, errs.NewInvalid(fmt.Sprintf("refund %d on %s exceeds refundable amount", l.Amount.Amount, l.PaymentRef))
		}
	}
	s.status, s.canceledAt, s.cancelAtPeriodEnd, s.updatedAt = StatusCanceled, &now, false, now
	s.record(&SubscriptionCanceledNowEvent{BaseEvent: s.base(userID), Reason: reason})
	var refunds []*Refund
	for _, l := range quote.Lines {
		if l.Amount.Amount <= 0 {
			continue
		}
		r := &Refund{id: uuid.Must(uuid.NewV7()), paymentRef: l.PaymentRef, amount: l.Amount, reason: reason, status: RefundStatusPending, createdAt: now, updatedAt: now}
		s.refunds = append(s.refunds, r)
		s.record(&RefundRequestedEvent{BaseEvent: s.base(userID), RefundID: r.id, Amount: l.Amount})
		refunds = append(refunds, r)
	}
	return refunds, nil
}

func (s *Subscription) UpdateRefund(refundID uuid.UUID, providerRefundID string, status RefundStatus, failureMessage *string, now time.Time, userID string) error {
	for _, r := range s.refunds {
		if r.id != refundID {
			continue
		}
		prev := r.status
		r.providerRefundID, r.status, r.failureMessage, r.updatedAt = &providerRefundID, status, failureMessage, now
		if prev != status {
			switch status {
			case RefundStatusSucceeded:
				s.record(&RefundSucceededEvent{BaseEvent: s.base(userID), RefundID: refundID})
			case RefundStatusFailed, RefundStatusCanceled:
				s.record(&RefundFailedEvent{BaseEvent: s.base(userID), RefundID: refundID})
			}
		}
		return nil
	}
	return errs.NewNotFound(fmt.Sprintf("refund %s not found", refundID))
}

func (s *Subscription) RefundByProviderID(providerRefundID string) *Refund {
	for _, r := range s.refunds {
		if r.providerRefundID != nil && *r.providerRefundID == providerRefundID {
			return r
		}
	}
	return nil
}

func (s *Subscription) RecordExternalRefund(providerRefundID, paymentRef string, amount Money, status RefundStatus, now time.Time, userID string) {
	if r := s.RefundByProviderID(providerRefundID); r != nil {
		_ = s.UpdateRefund(r.id, providerRefundID, status, nil, now, userID)
		return
	}
	id := uuid.Must(uuid.NewV7())
	s.refunds = append(s.refunds, &Refund{id: id, providerRefundID: &providerRefundID, paymentRef: paymentRef, amount: amount, reason: RefundReasonExternal, status: status, createdAt: now, updatedAt: now})
	s.record(&RefundRequestedEvent{BaseEvent: s.base(userID), RefundID: id, Amount: amount})
}

func (s *Subscription) IsCurrent() bool {
	return s.status != StatusCanceled && s.status != StatusIncompleteExpired
}

func (s *Subscription) GrantsAccess(now time.Time, policy *BillingPolicy) bool {
	switch s.status {
	case StatusActive:
		return true
	case StatusPastDue:
		return s.currentPeriodEnd != nil && now.Before(s.currentPeriodEnd.Add(policy.PastDueGrace()))
	default:
		return false
	}
}

func (s *Subscription) AccessUntil(policy *BillingPolicy) *time.Time {
	switch {
	case s.interval == IntervalLifetime || s.currentPeriodEnd == nil:
		return nil
	case s.status == StatusPastDue:
		t := s.currentPeriodEnd.Add(policy.PastDueGrace())
		return &t
	case s.status == StatusActive:
		return s.currentPeriodEnd
	default:
		return nil
	}
}

func (s *Subscription) ID() uuid.UUID                    { return s.id }
func (s *Subscription) WorkspaceID() uuid.UUID           { return s.workspaceID }
func (s *Subscription) Plan() PlanTier                   { return s.plan }
func (s *Subscription) Interval() Interval               { return s.interval }
func (s *Subscription) PriceID() uuid.UUID               { return s.priceID }
func (s *Subscription) Status() Status                   { return s.status }
func (s *Subscription) ProviderSubscriptionID() *string  { return s.providerSubscriptionID }
func (s *Subscription) ProviderPaymentID() *string       { return s.providerPaymentID }
func (s *Subscription) CurrentPeriodStart() *time.Time   { return s.currentPeriodStart }
func (s *Subscription) CurrentPeriodEnd() *time.Time     { return s.currentPeriodEnd }
func (s *Subscription) CancelAtPeriodEnd() bool          { return s.cancelAtPeriodEnd }
func (s *Subscription) CanceledAt() *time.Time           { return s.canceledAt }
func (s *Subscription) PaidAt() *time.Time               { return s.paidAt }
func (s *Subscription) ProviderSyncedAt() time.Time      { return s.providerSyncedAt }
func (s *Subscription) CreatedAt() time.Time             { return s.createdAt }
func (s *Subscription) UpdatedAt() time.Time             { return s.updatedAt }
func (s *Subscription) Refunds() []*Refund               { return s.refunds }
```

- [ ] **Step 5: Entitlements, repositories, unit of work**

```go
// internal/billing/domain/entitlement.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Entitlements struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
	Status      Status
	AccessUntil *time.Time
	Features    map[FeatureKey]bool
	Limits      map[LimitKey]int64
}

func ResolveEntitlements(workspaceID uuid.UUID, sub *Subscription, plans map[PlanTier]*Plan, policy *BillingPolicy, now time.Time) Entitlements {
	tier, status := PlanTierFree, StatusNone
	var until *time.Time
	if sub != nil {
		status = sub.Status()
		if sub.GrantsAccess(now, policy) {
			tier, until = sub.Plan(), sub.AccessUntil(policy)
		}
	}
	plan := plans[tier]
	return Entitlements{WorkspaceID: workspaceID, Plan: tier, Status: status, AccessUntil: until, Features: plan.Features(), Limits: plan.Limits()}
}
```

```go
// internal/billing/domain/repo.go
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Subscription, error)
	GetCurrentByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*Subscription, error) // nil, nil when none
	GetByProviderSubscriptionID(ctx context.Context, id string, forUpdate bool) (*Subscription, error)  // nil, nil when none
	GetByProviderPaymentID(ctx context.Context, id string, forUpdate bool) (*Subscription, error)       // nil, nil when none
	GetByRefundID(ctx context.Context, refundID uuid.UUID, forUpdate bool) (*Subscription, error)        // nil, nil when none
	ListCurrentByWorkspaces(ctx context.Context, workspaceIDs []uuid.UUID) ([]*Subscription, error)
	ListNeedingReconcile(ctx context.Context, now time.Time, limit int) ([]*Subscription, error)
	Save(ctx context.Context, s *Subscription) error
}

type BillingAccountRepo interface {
	GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*BillingAccount, error) // nil, nil when none
	ListByPayer(ctx context.Context, payerID string) ([]*BillingAccount, error)
	Save(ctx context.Context, a *BillingAccount) error
}

type PriceRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Price, error)
	GetByProviderPriceID(ctx context.Context, id string) (*Price, error)
	GetActive(ctx context.Context, plan PlanTier, interval Interval, currency string, forUpdate bool) (*Price, error) // nil, nil when none
	ListActive(ctx context.Context) ([]*Price, error)
	List(ctx context.Context) ([]*Price, error)
	Save(ctx context.Context, p *Price) error
}

type PlanRepo interface {
	Get(ctx context.Context, tier PlanTier, forUpdate bool) (*Plan, error)
	ListMap(ctx context.Context) (map[PlanTier]*Plan, error)
	Save(ctx context.Context, p *Plan) error
}

type BillingPolicyRepo interface {
	Get(ctx context.Context, forUpdate bool) (*BillingPolicy, error)
	Save(ctx context.Context, p *BillingPolicy) error
}

// ProviderEventLog records processed payment-provider events in the same transaction as their effect.
type ProviderEventLog interface {
	MarkProcessed(ctx context.Context, eventID, eventType string) (firstTime bool, err error)
	PruneBefore(ctx context.Context, before time.Time) (int64, error)
}

type RepoRegistry interface {
	Subscription() SubscriptionRepo
	BillingAccount() BillingAccountRepo
	Price() PriceRepo
	Plan() PlanRepo
	BillingPolicy() BillingPolicyRepo
	ProviderEventLog() ProviderEventLog
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(r RepoRegistry) error) error
}
```

`GetByID`/`Get` return `errs.NewNotFound(...)` when missing; the "nil, nil when none" methods are marked above.

`internal/billing/domain/wire.go`:

```go
package domain

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet()
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/billing/domain/...`
Expected: PASS (4 tests).

- [ ] **Step 7: Commit**

```bash
git add internal/billing/domain internal/billing/errs
git commit -m "feat(billing): domain model for plans, prices, accounts and subscriptions"
```

---

### Task 3: Persistence — migration, sqlc, repositories, unit of work

**Files:**

- Create: `internal/billing/infra/persistence/pgmigration/00001_init.sql`
- Create: `internal/billing/infra/persistence/pgsqlc/{plan,policy,price,account,subscription}.sql` (domain) and `{plan,policy,price,account,subscription}_read.sql` (read model), `providerevent.sql` (+ generated `*.go`)
- Create: `internal/billing/infra/persistence/{pg.go,wire.go}`
- Create: `internal/billing/infra/persistence/pgrepo/{uow.go,plan.go,policy.go,price.go,account.go,subscription.go,providerevent.go,common.go}` (write side)
- Create: `internal/billing/infra/persistence/pgreadmodel/billing.go` (read side) and `internal/billing/app/readmodel.go` (port)
- Create: `internal/billing/infra/persistence/pgmapper/{enum.go,row.go}` (shared sqlc ↔ domain mapping), `internal/billing/infra/persistence/pgjsonb/jsonb.go` (typed JSONB columns)
- Create: `cmd/billing/sqlc.yaml`, `cmd/billing/project.json` (targets `gen:sqlc`, `goose*`; the rest in Task 4)

**Interfaces:**

- Consumes: domain repo interfaces (Task 2).
- Produces: `pgrepo.NewUnitOfWork(pool *pgxpool.Pool, queries *pgsqlc.Queries) *UnitOfWork` implementing `domain.UnitOfWork`; `persistence.Pg` with `RunMigrations`, `IsMigrationDone`, `Ping` (same as note).
- Mappers (`pgmapper`, shared by `pgrepo` and `pgreadmodel`): `FromPlanTier`/`ToPlanTier`, `FromInterval`/`ToInterval`, `FromStatus`/`ToStatus`, `FromRefundReason`/`ToRefundReason`, `FromRefundStatus`/`ToRefundStatus` (domain `uint8` enum ↔ sqlc string enum) and `ToPlan`, `ToPlanMap`, `ToPolicy`, `ToPrice(s)`, `ToAccount`, `ToSubscriptions`. The `features`/`limits` JSONB columns are typed via sqlc column overrides to `pgjsonb.Features`/`pgjsonb.Limits` (unknown keys skipped with a warning), so repos never marshal by hand.
- `app.BillingReadModel` (port) implemented by `pgreadmodel.BillingReadModel`; query handlers use it instead of the unit of work.

Schema notes vs spec: the interval column is named `billing_interval` (`interval` is a SQL keyword); `refunds.payment_ref` added (the payment the refund targets).

- [ ] **Step 1: Migration**

```sql
-- internal/billing/infra/persistence/pgmigration/00001_init.sql
-- +goose Up
CREATE TYPE plan_tier AS ENUM ('free', 'plus', 'pro');
CREATE TYPE billing_interval AS ENUM ('month', 'year', 'lifetime');
CREATE TYPE subscription_status AS ENUM ('incomplete', 'active', 'past_due', 'unpaid', 'canceled', 'incomplete_expired');
CREATE TYPE refund_reason AS ENUM ('cancel_now', 'lifetime_window', 'replaced_by_lifetime', 'workspace_deleted', 'duplicate_checkout', 'external');
CREATE TYPE refund_status AS ENUM ('pending', 'succeeded', 'failed', 'canceled');

CREATE TABLE plans (
  tier plan_tier PRIMARY KEY,
  features JSONB NOT NULL,
  limits JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE billing_policy (
  id SMALLINT PRIMARY KEY CHECK (id = 1),
  past_due_grace_days INT NOT NULL,
  lifetime_refund_window_days INT NOT NULL,
  portal_configuration_id TEXT,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE prices (
  id UUID PRIMARY KEY,
  plan plan_tier NOT NULL CHECK (plan <> 'free'),
  billing_interval billing_interval NOT NULL,
  amount BIGINT NOT NULL CHECK (amount > 0),
  currency TEXT NOT NULL,
  provider_price_id TEXT NOT NULL UNIQUE,
  lookup_key TEXT NOT NULL,
  active BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  archived_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_prices_active ON prices(plan, billing_interval, currency) WHERE active;

CREATE TABLE billing_accounts (
  workspace_id UUID PRIMARY KEY,
  provider_customer_id TEXT NOT NULL UNIQUE,
  payer_id TEXT NOT NULL,
  payer_email TEXT NOT NULL,
  payer_name TEXT NOT NULL,
  payer_missing_since TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_billing_accounts_payer ON billing_accounts(payer_id);

CREATE TABLE subscriptions (
  id UUID PRIMARY KEY,
  workspace_id UUID NOT NULL REFERENCES billing_accounts(workspace_id),
  plan plan_tier NOT NULL,
  billing_interval billing_interval NOT NULL,
  price_id UUID NOT NULL REFERENCES prices(id),
  status subscription_status NOT NULL,
  provider_subscription_id TEXT UNIQUE,
  provider_payment_id TEXT UNIQUE,
  current_period_start TIMESTAMPTZ,
  current_period_end TIMESTAMPTZ,
  cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
  canceled_at TIMESTAMPTZ,
  paid_at TIMESTAMPTZ,
  provider_synced_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CHECK ((billing_interval = 'lifetime') = (provider_subscription_id IS NULL))
);
CREATE UNIQUE INDEX uq_subscriptions_current ON subscriptions(workspace_id)
  WHERE status NOT IN ('canceled', 'incomplete_expired');
CREATE INDEX idx_subscriptions_reconcile ON subscriptions(current_period_end, provider_synced_at)
  WHERE status NOT IN ('canceled', 'incomplete_expired') AND billing_interval <> 'lifetime';

CREATE TABLE refunds (
  id UUID PRIMARY KEY,
  subscription_id UUID NOT NULL REFERENCES subscriptions(id),
  provider_refund_id TEXT UNIQUE,
  payment_ref TEXT NOT NULL,
  amount BIGINT NOT NULL CHECK (amount > 0),
  currency TEXT NOT NULL,
  reason refund_reason NOT NULL,
  status refund_status NOT NULL,
  failure_message TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_refunds_subscription ON refunds(subscription_id);

CREATE TABLE processed_provider_events (
  event_id TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  received_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE processed_provider_events;
DROP TABLE refunds;
DROP TABLE subscriptions;
DROP TABLE billing_accounts;
DROP TABLE prices;
DROP TABLE billing_policy;
DROP TABLE plans;
DROP TYPE refund_status;
DROP TYPE refund_reason;
DROP TYPE subscription_status;
DROP TYPE billing_interval;
DROP TYPE plan_tier;
```

- [ ] **Step 2: sqlc config + nx targets**

`cmd/billing/sqlc.yaml`: copy `cmd/note/sqlc.yaml`; replace `internal/note` with `internal/billing`; remove the `temp_tables.sql` schema entry, the `rename: source_ids/target_ids` entries, the `rules:` block and the `trashed_by` overrides. Keep the uuid/timestamptz overrides, `emit_tracing` and `emit_dynamic_filter: true`. Queries use `sqlc.arg('name')` (never `$N`); for runtime-optional filters use `-- :if @param` with `sqlc.narg` as in `internal/note/infra/persistence/pgsqlc/note_read.sql`.

`cmd/billing/project.json`: copy `cmd/note/project.json`, replace `note` with `billing` everywhere (paths, binary, repository name, `dlv` port `23451` → `23453`). Task 4 relies on all targets.

Add column overrides for the JSONB columns (under `overrides.go.overrides`, next to the uuid/timestamptz ones; sqlc maps `jsonb` to `[]byte` otherwise, and pgx JSON-encodes the typed map):

```yaml
- column: plans.features
  go_type:
    import: github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgjsonb
    type: Features
- column: plans.limits
  go_type:
    import: github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgjsonb
    type: Limits
```

`cmd/billing/.env` (Task 4 adds the rest):

```
GOOSE_DBSTRING=postgres://notopiauit:notopiauit@localhost:5435/billing?sslmode=disable
GOOSE_DRIVER=postgres
GOOSE_MIGRATION_DIR=../../internal/billing/infra/persistence/pgmigration/
```

- [ ] **Step 3: Queries**

Two query sets live side by side in `pgsqlc/`, both with `sqlc.arg('name')` (never `$N`), even where they look alike:

- **Domain (write side)**, `<aggregate>.sql`: lock-aware lookups used by repos inside `uow.Execute`, plus upserts. Locking is a runtime flag via the dynamic-filter plugin: `FOR UPDATE -- :if @for_update` is the last line (a lone `;` follows on its own line), so one query serves both `forUpdate=true` and `false`. Generated params carry `ForUpdate bool`.
- **Read model**, `<aggregate>_read.sql`: every query is named `Read…` (as in `internal/note`), never locks, and is used only by `pgreadmodel`.

```sql
-- plan.sql
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
```

```sql
-- policy.sql
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
```

```sql
-- price.sql
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
```

```sql
-- account.sql
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
```

```sql
-- subscription.sql
-- name: GetSubscription :one
SELECT * FROM subscriptions
WHERE id = sqlc.arg('id')
FOR UPDATE -- :if @for_update
;

-- name: GetCurrentSubscription :one
SELECT * FROM subscriptions
WHERE workspace_id = sqlc.arg('workspace_id') AND status NOT IN ('canceled', 'incomplete_expired')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionByProviderSubscriptionID :one
SELECT * FROM subscriptions
WHERE provider_subscription_id = sqlc.arg('provider_subscription_id')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionByProviderPaymentID :one
SELECT * FROM subscriptions
WHERE provider_payment_id = sqlc.arg('provider_payment_id')
FOR UPDATE -- :if @for_update
;

-- name: GetSubscriptionIDByRefundID :one
SELECT subscription_id FROM refunds WHERE id = sqlc.arg('id');

-- name: ListCurrentSubscriptionsByWorkspaces :many
SELECT * FROM subscriptions
WHERE workspace_id = ANY(sqlc.arg('workspace_ids')::uuid[]) AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ListSubscriptionsNeedingReconcile :many
SELECT * FROM subscriptions
WHERE status NOT IN ('canceled', 'incomplete_expired')
  AND billing_interval <> 'lifetime'
  AND (current_period_end + INTERVAL '1 hour' < sqlc.arg('now')::timestamptz
       OR provider_synced_at < sqlc.arg('now')::timestamptz - INTERVAL '24 hours')
ORDER BY provider_synced_at
LIMIT sqlc.arg('row_limit');

-- name: UpsertSubscription :exec
INSERT INTO subscriptions (
  id, workspace_id, plan, billing_interval, price_id, status, provider_subscription_id, provider_payment_id,
  current_period_start, current_period_end, cancel_at_period_end, canceled_at, paid_at, provider_synced_at,
  created_at, updated_at
) VALUES (
  sqlc.arg('id'), sqlc.arg('workspace_id'), sqlc.arg('plan'), sqlc.arg('billing_interval'), sqlc.arg('price_id'),
  sqlc.arg('status'), sqlc.arg('provider_subscription_id'), sqlc.arg('provider_payment_id'),
  sqlc.arg('current_period_start'), sqlc.arg('current_period_end'), sqlc.arg('cancel_at_period_end'),
  sqlc.arg('canceled_at'), sqlc.arg('paid_at'), sqlc.arg('provider_synced_at'),
  sqlc.arg('created_at'), sqlc.arg('updated_at')
)
ON CONFLICT (id) DO UPDATE SET
  plan = EXCLUDED.plan,
  billing_interval = EXCLUDED.billing_interval,
  price_id = EXCLUDED.price_id,
  status = EXCLUDED.status,
  current_period_start = EXCLUDED.current_period_start,
  current_period_end = EXCLUDED.current_period_end,
  cancel_at_period_end = EXCLUDED.cancel_at_period_end,
  canceled_at = EXCLUDED.canceled_at,
  paid_at = EXCLUDED.paid_at,
  provider_synced_at = EXCLUDED.provider_synced_at,
  updated_at = EXCLUDED.updated_at;

-- name: ListRefundsBySubscriptionIDs :many
SELECT * FROM refunds WHERE subscription_id = ANY(sqlc.arg('subscription_ids')::uuid[]) ORDER BY created_at;

-- name: UpsertRefund :exec
INSERT INTO refunds (id, subscription_id, provider_refund_id, payment_ref, amount, currency, reason, status, failure_message, created_at, updated_at)
VALUES (
  sqlc.arg('id'), sqlc.arg('subscription_id'), sqlc.arg('provider_refund_id'), sqlc.arg('payment_ref'), sqlc.arg('amount'),
  sqlc.arg('currency'), sqlc.arg('reason'), sqlc.arg('status'), sqlc.arg('failure_message'), sqlc.arg('created_at'), sqlc.arg('updated_at')
)
ON CONFLICT (id) DO UPDATE SET
  provider_refund_id = EXCLUDED.provider_refund_id,
  status = EXCLUDED.status,
  failure_message = EXCLUDED.failure_message,
  updated_at = EXCLUDED.updated_at;
```

```sql
-- providerevent.sql
-- name: InsertProviderEvent :execrows
INSERT INTO processed_provider_events (event_id, event_type, received_at)
VALUES (sqlc.arg('event_id'), sqlc.arg('event_type'), sqlc.arg('received_at'))
ON CONFLICT (event_id) DO NOTHING;

-- name: DeleteProviderEventsBefore :execrows
DELETE FROM processed_provider_events WHERE received_at < sqlc.arg('received_before');
```

Read-model queries (no locks, no `for_update`):

```sql
-- plan_read.sql
-- name: ReadListPlans :many
SELECT * FROM plans ORDER BY tier;
```

```sql
-- policy_read.sql
-- name: ReadGetPolicy :one
SELECT * FROM billing_policy WHERE id = 1;
```

```sql
-- price_read.sql
-- name: ReadListActivePrices :many
SELECT * FROM prices WHERE active ORDER BY plan, billing_interval;

-- name: ReadListPrices :many
SELECT * FROM prices ORDER BY created_at DESC;
```

```sql
-- account_read.sql
-- name: ReadGetBillingAccount :one
SELECT * FROM billing_accounts WHERE workspace_id = sqlc.arg('workspace_id');
```

```sql
-- subscription_read.sql
-- name: ReadGetCurrentSubscription :one
SELECT * FROM subscriptions
WHERE workspace_id = sqlc.arg('workspace_id') AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ReadListCurrentSubscriptionsByWorkspaces :many
SELECT * FROM subscriptions
WHERE workspace_id = ANY(sqlc.arg('workspace_ids')::uuid[]) AND status NOT IN ('canceled', 'incomplete_expired');

-- name: ReadListRefundsBySubscriptionIDs :many
SELECT * FROM refunds WHERE subscription_id = ANY(sqlc.arg('subscription_ids')::uuid[]) ORDER BY created_at;
```

Run: `nx run billing:gen:sqlc`. Generated model types: `pgsqlc.Plan`, `BillingPolicy`, `Price`, `BillingAccount`, `Subscription`, `Refund`; string enums `PlanTier`, `BillingInterval`, `SubscriptionStatus`, `RefundReason`, `RefundStatus` (consts `pgsqlc.PlanTierFree`, `pgsqlc.SubscriptionStatusPastDue`, …); `Plan.Features`/`Plan.Limits` are `pgjsonb.Features`/`pgjsonb.Limits` (column overrides, Step 2). The `Read…` queries return the same model structs. The code below uses those names; adjust if sqlc names a field differently (check `models.go`), and if a lock-only query (`GetPolicy`, which has no other param) generates a bare `forUpdate bool` argument instead of a `Params` struct, pass it positionally. These sqlc enums never leave `pgmapper`/`pgrepo`/`pgreadmodel`: `pgmapper/enum.go` maps them to the domain `uint8` enums.

- [ ] **Step 4: Pg + wire**

`internal/billing/infra/persistence/pg.go`: copy `internal/note/infra/persistence/pg.go` verbatim, replacing the `pgsqlc` import path with `internal/billing/infra/persistence/pgsqlc`.

```go
// internal/billing/infra/persistence/wire.go
package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgreadmodel"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgrepo"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

var ProviderSet = wire.NewSet(
	ProvideGooseProvider,
	ProvidePg,
	ProvidePgPool,
	ProvideSQLCQueries,
	ProvidePgxPoolStdlib,
	pgrepo.ProvideUnitOfWork,
	pgreadmodel.ProvideBillingReadModel,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
	wire.Bind(new(domain.UnitOfWork), new(*pgrepo.UnitOfWork)),
	wire.Bind(new(app.BillingReadModel), new(*pgreadmodel.BillingReadModel)),
)
```

- [ ] **Step 5: Unit of work**

```go
// internal/billing/infra/persistence/pgrepo/uow.go
package pgrepo

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type UnitOfWork struct {
	pool    *pgxpool.Pool
	queries *pgsqlc.Queries
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(pool *pgxpool.Pool, queries *pgsqlc.Queries) *UnitOfWork {
	return &UnitOfWork{pool: pool, queries: queries}
}

var ProvideUnitOfWork = NewUnitOfWork

func (u *UnitOfWork) Execute(ctx context.Context, fn func(r domain.RepoRegistry) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return errs.NewInternal("failed to begin transaction", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rollback transaction", slog.Any("error", err))
		}
	}()
	if err := fn(&registry{q: u.queries.WithTx(tx)}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.NewInternal("failed to commit transaction", err)
	}
	return nil
}

type registry struct{ q *pgsqlc.Queries }

func (r *registry) Subscription() domain.SubscriptionRepo     { return &subscriptionRepo{q: r.q} }
func (r *registry) BillingAccount() domain.BillingAccountRepo { return &accountRepo{q: r.q} }
func (r *registry) Price() domain.PriceRepo                   { return &priceRepo{q: r.q} }
func (r *registry) Plan() domain.PlanRepo                     { return &planRepo{q: r.q} }
func (r *registry) BillingPolicy() domain.BillingPolicyRepo   { return &policyRepo{q: r.q} }
func (r *registry) ProviderEventLog() domain.ProviderEventLog { return &providerEventLog{q: r.q} }
```

```go
// internal/billing/infra/persistence/pgrepo/common.go
package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func notFound(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func wrap(msg string, err error) error {
	if err == nil {
		return nil
	}
	return errs.NewInternal(msg, err)
}

func drain(ctx context.Context, events []domain.Event) {
	for _, e := range events {
		slog.DebugContext(ctx, "billing domain event", slog.String("type", typeName(e)), slog.String("aggregate_id", e.GetAggregateID()), slog.String("user_id", e.GetUserID()))
	}
}

func typeName(v any) string {
	return fmt.Sprintf("%T", v)
}
```

`fromX` (in `pgmapper`) returns `""` for an Unspecified/unknown value, which the Postgres enum rejects on write. `toX` returns the Unspecified zero for an unknown value; the column's `CREATE TYPE ... AS ENUM` makes that unreachable.

The `features`/`limits` JSONB columns need no hand-written marshal calls: sqlc maps `jsonb` to `[]byte` by default (`sqlc-gen-go` `postgresql_type.go`, pgx/v5), but a column `go_type` override swaps in a typed map, and pgx's JSONB codec then runs `json.Marshal`/`json.Unmarshal` on it. The types live in their own package (not `pgmapper`) because the generated `pgsqlc` imports them and `pgmapper` imports `pgsqlc`. `UnmarshalJSON` skips keys the code no longer knows (warning log) instead of failing, so a removed feature/limit key never breaks entitlement reads; the next `Save` drops it.

```go
// internal/billing/infra/persistence/pgjsonb/jsonb.go
package pgjsonb

import (
	"encoding/json"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type Features map[domain.FeatureKey]bool

func (f Features) MarshalJSON() ([]byte, error) { return marshalKeyed(f) }

func (f *Features) UnmarshalJSON(raw []byte) (err error) {
	*f, err = unmarshalKeyed[domain.FeatureKey, bool](raw, domain.ParseFeatureKey)
	return err
}

type Limits map[domain.LimitKey]int64

func (l Limits) MarshalJSON() ([]byte, error) { return marshalKeyed(l) }

func (l *Limits) UnmarshalJSON(raw []byte) (err error) {
	*l, err = unmarshalKeyed[domain.LimitKey, int64](raw, domain.ParseLimitKey)
	return err
}

type stringKey interface {
	comparable
	String() string
}

func marshalKeyed[K stringKey, V any](m map[K]V) ([]byte, error) {
	out := make(map[string]V, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return json.Marshal(out)
}

func unmarshalKeyed[K comparable, V any](raw []byte, parse func(string) (K, error)) (map[K]V, error) {
	var in map[string]V
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	out := make(map[K]V, len(in))
	for s, v := range in {
		k, err := parse(s)
		if err != nil {
			slog.Warn("skipping unknown plan key", slog.String("key", s))
			continue
		}
		out[k] = v
	}
	return out, nil
}
```

`pgmapper` is shared by `pgrepo` (write side) and `pgreadmodel` so the sqlc-row → domain mapping exists once. `ToSubscriptions` takes the already-fetched refund rows; each side fetches them with its own query (`ListRefundsBySubscriptionIDs` vs `ReadListRefundsBySubscriptionIDs`).

```go
// internal/billing/infra/persistence/pgmapper/enum.go
package pgmapper

import (
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

func FromPlanTier(t domain.PlanTier) pgsqlc.PlanTier {
	switch t {
	case domain.PlanTierFree:
		return pgsqlc.PlanTierFree
	case domain.PlanTierPlus:
		return pgsqlc.PlanTierPlus
	case domain.PlanTierPro:
		return pgsqlc.PlanTierPro
	default:
		return ""
	}
}

func ToPlanTier(t pgsqlc.PlanTier) domain.PlanTier {
	switch t {
	case pgsqlc.PlanTierFree:
		return domain.PlanTierFree
	case pgsqlc.PlanTierPlus:
		return domain.PlanTierPlus
	case pgsqlc.PlanTierPro:
		return domain.PlanTierPro
	default:
		return domain.PlanTierUnspecified
	}
}

func FromInterval(i domain.Interval) pgsqlc.BillingInterval {
	switch i {
	case domain.IntervalMonth:
		return pgsqlc.BillingIntervalMonth
	case domain.IntervalYear:
		return pgsqlc.BillingIntervalYear
	case domain.IntervalLifetime:
		return pgsqlc.BillingIntervalLifetime
	default:
		return ""
	}
}

func ToInterval(i pgsqlc.BillingInterval) domain.Interval {
	switch i {
	case pgsqlc.BillingIntervalMonth:
		return domain.IntervalMonth
	case pgsqlc.BillingIntervalYear:
		return domain.IntervalYear
	case pgsqlc.BillingIntervalLifetime:
		return domain.IntervalLifetime
	default:
		return domain.IntervalUnspecified
	}
}

func FromStatus(s domain.Status) pgsqlc.SubscriptionStatus {
	switch s {
	case domain.StatusIncomplete:
		return pgsqlc.SubscriptionStatusIncomplete
	case domain.StatusActive:
		return pgsqlc.SubscriptionStatusActive
	case domain.StatusPastDue:
		return pgsqlc.SubscriptionStatusPastDue
	case domain.StatusUnpaid:
		return pgsqlc.SubscriptionStatusUnpaid
	case domain.StatusCanceled:
		return pgsqlc.SubscriptionStatusCanceled
	case domain.StatusIncompleteExpired:
		return pgsqlc.SubscriptionStatusIncompleteExpired
	default:
		return ""
	}
}

func ToStatus(s pgsqlc.SubscriptionStatus) domain.Status {
	switch s {
	case pgsqlc.SubscriptionStatusIncomplete:
		return domain.StatusIncomplete
	case pgsqlc.SubscriptionStatusActive:
		return domain.StatusActive
	case pgsqlc.SubscriptionStatusPastDue:
		return domain.StatusPastDue
	case pgsqlc.SubscriptionStatusUnpaid:
		return domain.StatusUnpaid
	case pgsqlc.SubscriptionStatusCanceled:
		return domain.StatusCanceled
	case pgsqlc.SubscriptionStatusIncompleteExpired:
		return domain.StatusIncompleteExpired
	default:
		return domain.StatusNone
	}
}

func FromRefundReason(r domain.RefundReason) pgsqlc.RefundReason {
	switch r {
	case domain.RefundReasonCancelNow:
		return pgsqlc.RefundReasonCancelNow
	case domain.RefundReasonLifetimeWindow:
		return pgsqlc.RefundReasonLifetimeWindow
	case domain.RefundReasonReplacedByLifetime:
		return pgsqlc.RefundReasonReplacedByLifetime
	case domain.RefundReasonWorkspaceDeleted:
		return pgsqlc.RefundReasonWorkspaceDeleted
	case domain.RefundReasonDuplicateCheckout:
		return pgsqlc.RefundReasonDuplicateCheckout
	case domain.RefundReasonExternal:
		return pgsqlc.RefundReasonExternal
	default:
		return ""
	}
}

func ToRefundReason(r pgsqlc.RefundReason) domain.RefundReason {
	switch r {
	case pgsqlc.RefundReasonCancelNow:
		return domain.RefundReasonCancelNow
	case pgsqlc.RefundReasonLifetimeWindow:
		return domain.RefundReasonLifetimeWindow
	case pgsqlc.RefundReasonReplacedByLifetime:
		return domain.RefundReasonReplacedByLifetime
	case pgsqlc.RefundReasonWorkspaceDeleted:
		return domain.RefundReasonWorkspaceDeleted
	case pgsqlc.RefundReasonDuplicateCheckout:
		return domain.RefundReasonDuplicateCheckout
	case pgsqlc.RefundReasonExternal:
		return domain.RefundReasonExternal
	default:
		return domain.RefundReasonUnspecified
	}
}

func FromRefundStatus(s domain.RefundStatus) pgsqlc.RefundStatus {
	switch s {
	case domain.RefundStatusPending:
		return pgsqlc.RefundStatusPending
	case domain.RefundStatusSucceeded:
		return pgsqlc.RefundStatusSucceeded
	case domain.RefundStatusFailed:
		return pgsqlc.RefundStatusFailed
	case domain.RefundStatusCanceled:
		return pgsqlc.RefundStatusCanceled
	default:
		return ""
	}
}

func ToRefundStatus(s pgsqlc.RefundStatus) domain.RefundStatus {
	switch s {
	case pgsqlc.RefundStatusPending:
		return domain.RefundStatusPending
	case pgsqlc.RefundStatusSucceeded:
		return domain.RefundStatusSucceeded
	case pgsqlc.RefundStatusFailed:
		return domain.RefundStatusFailed
	case pgsqlc.RefundStatusCanceled:
		return domain.RefundStatusCanceled
	default:
		return domain.RefundStatusUnspecified
	}
}
```

```go
// internal/billing/infra/persistence/pgmapper/row.go
package pgmapper

import (
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

func ToPlan(row *pgsqlc.Plan) *domain.Plan {
	return domain.UnmarshalPlan(ToPlanTier(row.Tier), row.Features, row.Limits, row.UpdatedAt)
}

func ToPlanMap(rows []*pgsqlc.Plan) map[domain.PlanTier]*domain.Plan {
	out := make(map[domain.PlanTier]*domain.Plan, len(rows))
	for _, row := range rows {
		p := ToPlan(row)
		out[p.Tier()] = p
	}
	return out
}

func ToPolicy(row *pgsqlc.BillingPolicy) *domain.BillingPolicy {
	return domain.UnmarshalBillingPolicy(int(row.PastDueGraceDays), int(row.LifetimeRefundWindowDays), row.PortalConfigurationID, row.UpdatedAt)
}

func ToPrice(row *pgsqlc.Price) *domain.Price {
	return domain.UnmarshalPrice(domain.UnmarshalPriceParams{
		ID: row.ID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
		Amount: domain.Money{Amount: row.Amount, Currency: row.Currency}, ProviderPriceID: row.ProviderPriceID,
		LookupKey: row.LookupKey, Active: row.Active, CreatedAt: row.CreatedAt, ArchivedAt: row.ArchivedAt,
	})
}

func ToPrices(rows []*pgsqlc.Price) []*domain.Price {
	out := make([]*domain.Price, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToPrice(row))
	}
	return out
}

func ToAccount(row *pgsqlc.BillingAccount) *domain.BillingAccount {
	return domain.UnmarshalBillingAccount(row.WorkspaceID, row.ProviderCustomerID,
		domain.Payer{ID: row.PayerID, Email: row.PayerEmail, Name: row.PayerName},
		row.PayerMissingSince, row.CreatedAt, row.UpdatedAt)
}

func ToSubscriptions(rows []*pgsqlc.Subscription, refundRows []*pgsqlc.Refund) []*domain.Subscription {
	refunds := map[uuid.UUID][]*domain.Refund{}
	for _, rr := range refundRows {
		refunds[rr.SubscriptionID] = append(refunds[rr.SubscriptionID], domain.UnmarshalRefund(domain.UnmarshalRefundParams{
			ID: rr.ID, ProviderRefundID: rr.ProviderRefundID, PaymentRef: rr.PaymentRef,
			Amount: domain.Money{Amount: rr.Amount, Currency: rr.Currency}, Reason: ToRefundReason(rr.Reason),
			Status: ToRefundStatus(rr.Status), FailureMessage: rr.FailureMessage, CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
		}))
	}
	out := make([]*domain.Subscription, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.UnmarshalSubscription(domain.UnmarshalSubscriptionParams{
			ID: row.ID, WorkspaceID: row.WorkspaceID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
			PriceID: row.PriceID, Status: ToStatus(row.Status), ProviderSubscriptionID: row.ProviderSubscriptionID,
			ProviderPaymentID: row.ProviderPaymentID, CurrentPeriodStart: row.CurrentPeriodStart, CurrentPeriodEnd: row.CurrentPeriodEnd,
			CancelAtPeriodEnd: row.CancelAtPeriodEnd, CanceledAt: row.CanceledAt, PaidAt: row.PaidAt,
			ProviderSyncedAt: row.ProviderSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Refunds: refunds[row.ID],
		}))
	}
	return out
}
```

- [ ] **Step 6: Repositories**

```go
// internal/billing/infra/persistence/pgrepo/plan.go
package pgrepo

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgjsonb"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type planRepo struct{ q *pgsqlc.Queries }

func (r *planRepo) Get(ctx context.Context, tier domain.PlanTier, forUpdate bool) (*domain.Plan, error) {
	row, err := r.q.GetPlan(ctx, &pgsqlc.GetPlanParams{Tier: pgmapper.FromPlanTier(tier), ForUpdate: forUpdate})
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("plan %s not found", tier))
	}
	if err != nil {
		return nil, wrap("get plan", err)
	}
	return pgmapper.ToPlan(row), nil
}

func (r *planRepo) ListMap(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	rows, err := r.q.ListPlans(ctx)
	if err != nil {
		return nil, wrap("list plans", err)
	}
	return pgmapper.ToPlanMap(rows), nil
}

func (r *planRepo) Save(ctx context.Context, p *domain.Plan) error {
	err := r.q.UpsertPlan(ctx, &pgsqlc.UpsertPlanParams{
		Tier: pgmapper.FromPlanTier(p.Tier()), Features: pgjsonb.Features(p.Features()),
		Limits: pgjsonb.Limits(p.Limits()), UpdatedAt: p.UpdatedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save plan", err)
}
```

```go
// internal/billing/infra/persistence/pgrepo/policy.go
package pgrepo

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type policyRepo struct{ q *pgsqlc.Queries }

func (r *policyRepo) Get(ctx context.Context, forUpdate bool) (*domain.BillingPolicy, error) {
	row, err := r.q.GetPolicy(ctx, forUpdate)
	if notFound(err) {
		return nil, errs.NewNotFound("billing policy not seeded")
	}
	if err != nil {
		return nil, wrap("get policy", err)
	}
	return pgmapper.ToPolicy(row), nil
}

func (r *policyRepo) Save(ctx context.Context, p *domain.BillingPolicy) error {
	err := r.q.UpsertPolicy(ctx, &pgsqlc.UpsertPolicyParams{
		PastDueGraceDays:         int32(p.PastDueGraceDays()),
		LifetimeRefundWindowDays: int32(p.LifetimeRefundWindowDays()),
		PortalConfigurationID:    p.PortalConfigurationID(),
		UpdatedAt:                p.UpdatedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save policy", err)
}
```

```go
// internal/billing/infra/persistence/pgrepo/price.go
package pgrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type priceRepo struct{ q *pgsqlc.Queries }

func (r *priceRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetPrice(ctx, &pgsqlc.GetPriceParams{ID: id, ForUpdate: forUpdate})
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("price %s not found", id))
	}
	if err != nil {
		return nil, wrap("get price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetByProviderPriceID(ctx context.Context, id string) (*domain.Price, error) {
	row, err := r.q.GetPriceByProviderPriceID(ctx, id)
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("provider price %s not found", id))
	}
	if err != nil {
		return nil, wrap("get price by provider id", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetActive(ctx context.Context, plan domain.PlanTier, interval domain.Interval, currency string, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetActivePrice(ctx, &pgsqlc.GetActivePriceParams{
		Plan: pgmapper.FromPlanTier(plan), BillingInterval: pgmapper.FromInterval(interval), Currency: currency, ForUpdate: forUpdate,
	})
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, wrap("get active price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) ListActive(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListActivePrices(ctx)
	return pgmapper.ToPrices(rows), wrap("list active prices", err)
}

func (r *priceRepo) List(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListPrices(ctx)
	return pgmapper.ToPrices(rows), wrap("list prices", err)
}

func (r *priceRepo) Save(ctx context.Context, p *domain.Price) error {
	err := r.q.UpsertPrice(ctx, &pgsqlc.UpsertPriceParams{
		ID: p.ID(), Plan: pgmapper.FromPlanTier(p.Plan()), BillingInterval: pgmapper.FromInterval(p.Interval()),
		Amount: p.Amount().Amount, Currency: p.Amount().Currency, ProviderPriceID: p.ProviderPriceID(),
		LookupKey: p.LookupKey(), Active: p.Active(), CreatedAt: p.CreatedAt(), ArchivedAt: p.ArchivedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save price", err)
}
```

```go
// internal/billing/infra/persistence/pgrepo/account.go
package pgrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type accountRepo struct{ q *pgsqlc.Queries }

func (r *accountRepo) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*domain.BillingAccount, error) {
	row, err := r.q.GetBillingAccount(ctx, &pgsqlc.GetBillingAccountParams{WorkspaceID: workspaceID, ForUpdate: forUpdate})
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, wrap("get billing account", err)
	}
	return pgmapper.ToAccount(row), nil
}

func (r *accountRepo) ListByPayer(ctx context.Context, payerID string) ([]*domain.BillingAccount, error) {
	rows, err := r.q.ListBillingAccountsByPayer(ctx, payerID)
	if err != nil {
		return nil, wrap("list billing accounts by payer", err)
	}
	out := make([]*domain.BillingAccount, 0, len(rows))
	for _, row := range rows {
		out = append(out, pgmapper.ToAccount(row))
	}
	return out, nil
}

func (r *accountRepo) Save(ctx context.Context, a *domain.BillingAccount) error {
	p := a.Payer()
	err := r.q.UpsertBillingAccount(ctx, &pgsqlc.UpsertBillingAccountParams{
		WorkspaceID: a.WorkspaceID(), ProviderCustomerID: a.ProviderCustomerID(),
		PayerID: p.ID, PayerEmail: p.Email, PayerName: p.Name, PayerMissingSince: a.PayerMissingSince(),
		CreatedAt: a.CreatedAt(), UpdatedAt: a.UpdatedAt(),
	})
	drain(ctx, a.PopEvents())
	return wrap("save billing account", err)
}
```

```go
// internal/billing/infra/persistence/pgrepo/subscription.go
package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type subscriptionRepo struct{ q *pgsqlc.Queries }

func (r *subscriptionRepo) hydrate(ctx context.Context, rows []*pgsqlc.Subscription) ([]*domain.Subscription, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	refundRows, err := r.q.ListRefundsBySubscriptionIDs(ctx, ids)
	if err != nil {
		return nil, wrap("list refunds", err)
	}
	return pgmapper.ToSubscriptions(rows, refundRows), nil
}

func (r *subscriptionRepo) one(ctx context.Context, row *pgsqlc.Subscription, err error, missingIsNil bool, what string) (*domain.Subscription, error) {
	if notFound(err) {
		if missingIsNil {
			return nil, nil
		}
		return nil, errs.NewNotFound(fmt.Sprintf("subscription %s not found", what))
	}
	if err != nil {
		return nil, wrap("get subscription", err)
	}
	subs, err := r.hydrate(ctx, []*pgsqlc.Subscription{row})
	if err != nil {
		return nil, err
	}
	return subs[0], nil
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscription(ctx, &pgsqlc.GetSubscriptionParams{ID: id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, false, id.String())
}

func (r *subscriptionRepo) GetCurrentByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetCurrentSubscription(ctx, &pgsqlc.GetCurrentSubscriptionParams{WorkspaceID: workspaceID, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByProviderSubscriptionID(ctx context.Context, id string, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscriptionByProviderSubscriptionID(ctx, &pgsqlc.GetSubscriptionByProviderSubscriptionIDParams{ProviderSubscriptionID: &id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByProviderPaymentID(ctx context.Context, id string, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscriptionByProviderPaymentID(ctx, &pgsqlc.GetSubscriptionByProviderPaymentIDParams{ProviderPaymentID: &id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByRefundID(ctx context.Context, refundID uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	subID, err := r.q.GetSubscriptionIDByRefundID(ctx, refundID)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, wrap("get subscription id by refund", err)
	}
	return r.GetByID(ctx, subID, forUpdate)
}

func (r *subscriptionRepo) ListCurrentByWorkspaces(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error) {
	rows, err := r.q.ListCurrentSubscriptionsByWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, wrap("list current subscriptions", err)
	}
	return r.hydrate(ctx, rows)
}

func (r *subscriptionRepo) ListNeedingReconcile(ctx context.Context, now time.Time, limit int) ([]*domain.Subscription, error) {
	rows, err := r.q.ListSubscriptionsNeedingReconcile(ctx, &pgsqlc.ListSubscriptionsNeedingReconcileParams{Now: now, RowLimit: int32(limit)})
	if err != nil {
		return nil, wrap("list subscriptions needing reconcile", err)
	}
	return r.hydrate(ctx, rows)
}

func (r *subscriptionRepo) Save(ctx context.Context, s *domain.Subscription) error {
	err := r.q.UpsertSubscription(ctx, &pgsqlc.UpsertSubscriptionParams{
		ID: s.ID(), WorkspaceID: s.WorkspaceID(), Plan: pgmapper.FromPlanTier(s.Plan()), BillingInterval: pgmapper.FromInterval(s.Interval()),
		PriceID: s.PriceID(), Status: pgmapper.FromStatus(s.Status()), ProviderSubscriptionID: s.ProviderSubscriptionID(),
		ProviderPaymentID: s.ProviderPaymentID(), CurrentPeriodStart: s.CurrentPeriodStart(), CurrentPeriodEnd: s.CurrentPeriodEnd(),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd(), CanceledAt: s.CanceledAt(), PaidAt: s.PaidAt(),
		ProviderSyncedAt: s.ProviderSyncedAt(), CreatedAt: s.CreatedAt(), UpdatedAt: s.UpdatedAt(),
	})
	if err != nil {
		return wrap("save subscription", err)
	}
	for _, rf := range s.Refunds() {
		if err := r.q.UpsertRefund(ctx, &pgsqlc.UpsertRefundParams{
			ID: rf.ID(), SubscriptionID: s.ID(), ProviderRefundID: rf.ProviderRefundID(), PaymentRef: rf.PaymentRef(),
			Amount: rf.Amount().Amount, Currency: rf.Amount().Currency, Reason: pgmapper.FromRefundReason(rf.Reason()),
			Status: pgmapper.FromRefundStatus(rf.Status()), FailureMessage: rf.FailureMessage(), CreatedAt: rf.CreatedAt(), UpdatedAt: rf.UpdatedAt(),
		}); err != nil {
			return wrap("save refund", err)
		}
	}
	drain(ctx, s.PopEvents())
	return nil
}
```

```go
// internal/billing/infra/persistence/pgrepo/providerevent.go
package pgrepo

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type providerEventLog struct{ q *pgsqlc.Queries }

func (l *providerEventLog) MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	n, err := l.q.InsertProviderEvent(ctx, &pgsqlc.InsertProviderEventParams{EventID: eventID, EventType: eventType, ReceivedAt: time.Now()})
	if err != nil {
		return false, wrap("mark provider event processed", err)
	}
	return n == 1, nil
}

func (l *providerEventLog) PruneBefore(ctx context.Context, before time.Time) (int64, error) {
	n, err := l.q.DeleteProviderEventsBefore(ctx, before)
	return n, wrap("prune provider events", err)
}
```

- [ ] **Step 7: Read model**

Query handlers (`q_*`) never open a transaction or take locks: they depend on one `app.BillingReadModel` port, implemented by `pgreadmodel` on top of the `Read…` sqlc queries. Mapping rows to domain objects reuses `pgmapper`; `Unmarshal…` bypasses validation and events, so nothing here can mutate state.

```go
// internal/billing/app/readmodel.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type BillingReadModel interface {
	CurrentSubscription(ctx context.Context, workspaceID uuid.UUID) (*domain.Subscription, error) // nil, nil when none
	CurrentSubscriptions(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error)
	Account(ctx context.Context, workspaceID uuid.UUID) (*domain.BillingAccount, error) // nil, nil when none
	Plans(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error)
	Policy(ctx context.Context) (*domain.BillingPolicy, error)
	ActivePrices(ctx context.Context) ([]*domain.Price, error)
	Prices(ctx context.Context) ([]*domain.Price, error)
}
```

```go
// internal/billing/infra/persistence/pgreadmodel/billing.go
package pgreadmodel

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type BillingReadModel struct{ q *pgsqlc.Queries }

var _ app.BillingReadModel = (*BillingReadModel)(nil)

func NewBillingReadModel(q *pgsqlc.Queries) *BillingReadModel { return &BillingReadModel{q: q} }

var ProvideBillingReadModel = NewBillingReadModel

func toErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return errs.NewInternal(msg, err)
}

func (m *BillingReadModel) subscriptions(ctx context.Context, rows []*pgsqlc.Subscription) ([]*domain.Subscription, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	refundRows, err := m.q.ReadListRefundsBySubscriptionIDs(ctx, ids)
	if err != nil {
		return nil, toErr("read refunds", err)
	}
	return pgmapper.ToSubscriptions(rows, refundRows), nil
}

func (m *BillingReadModel) CurrentSubscription(ctx context.Context, workspaceID uuid.UUID) (*domain.Subscription, error) {
	row, err := m.q.ReadGetCurrentSubscription(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, toErr("read current subscription", err)
	}
	subs, err := m.subscriptions(ctx, []*pgsqlc.Subscription{row})
	if err != nil {
		return nil, err
	}
	return subs[0], nil
}

func (m *BillingReadModel) CurrentSubscriptions(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error) {
	rows, err := m.q.ReadListCurrentSubscriptionsByWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, toErr("read current subscriptions", err)
	}
	return m.subscriptions(ctx, rows)
}

func (m *BillingReadModel) Account(ctx context.Context, workspaceID uuid.UUID) (*domain.BillingAccount, error) {
	row, err := m.q.ReadGetBillingAccount(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, toErr("read billing account", err)
	}
	return pgmapper.ToAccount(row), nil
}

func (m *BillingReadModel) Plans(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	rows, err := m.q.ReadListPlans(ctx)
	if err != nil {
		return nil, toErr("read plans", err)
	}
	return pgmapper.ToPlanMap(rows), nil
}

func (m *BillingReadModel) Policy(ctx context.Context) (*domain.BillingPolicy, error) {
	row, err := m.q.ReadGetPolicy(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound("billing policy not seeded")
	}
	if err != nil {
		return nil, toErr("read policy", err)
	}
	return pgmapper.ToPolicy(row), nil
}

func (m *BillingReadModel) ActivePrices(ctx context.Context) ([]*domain.Price, error) {
	rows, err := m.q.ReadListActivePrices(ctx)
	if err != nil {
		return nil, toErr("read active prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}

func (m *BillingReadModel) Prices(ctx context.Context) ([]*domain.Price, error) {
	rows, err := m.q.ReadListPrices(ctx)
	if err != nil {
		return nil, toErr("read prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}
```

The `Read…` queries each take a single param, so sqlc generates positional arguments (as used above).

- [ ] **Step 8: Build, validate migration, commit**

```bash
go build ./internal/billing/...
docker compose -f deploy/compose/compose.yaml up -d billing_postgresql   # added in Task 4 Step 6; run Task 4 Step 6 first if absent
nx run billing:goose:validate
git add internal/billing/infra/persistence cmd/billing/sqlc.yaml cmd/billing/project.json cmd/billing/.env
git commit -m "feat(billing): persistence with goose migration, sqlc and repositories"
```

---

### Task 4: Service runtime — config, app skeleton, HTTP/gRPC/health, notifier, compose, Traefik

**Files:**

- Create: `internal/billing/config/{config.go,viper.go,wire.go}`
- Create: `internal/billing/component/{validator.go,eventtopic.go,wire.go}`
- Create: `internal/billing/app/{app.go,wire.go,svc_authorization.go,svc_notifier.go}`
- Create: `internal/billing/infra/{wire.go,redis/redis.go,notifier/kafka.go,notifier/map.go,service/authorization.go}`
- Create: `internal/billing/controller/{wire.go,http/{http.go,err.go,user.go,unimplemented.go},grpc/{grpc.go,err.go,service.go},health/health.go}`
- Create: `internal/billing/{server.go,wire.go,AGENTS.md}`, `internal/billing/CLAUDE.md` → symlink to `AGENTS.md`
- Create: `cmd/billing/{main.go,wire.go,metadata.go,Dockerfile}`; extend `cmd/billing/.env`
- Create: `deploy/compose/services/billing.yaml`, `deploy/compose/infrastructure/traefik/dynamic/billing.yaml`
- Modify: `deploy/compose/compose.yaml` (include billing), `.gitignore` if `cmd/*/` binaries are listed per service

**Interfaces:**

- Produces: `app.Server{Cmds *Cmds; Queries *Queries}` (fields added by later tasks), `app.AuthorizationSvc`, `app.WorkspaceRole` (`uint8`, zero `WorkspaceRoleNone`), `app.Notifier`; mappers `service.toAppWorkspaceRole` (pb → app) and `notifier.toSharePlan`/`toShareStatus` (domain → `api/share`); `http.StrictHandler{App *app.Server; Identity *commonconfig.Identity}`, `grpc.ServiceServer{app *app.Server}`, `billing.Server.Run(ctx)`.

- [ ] **Step 1: Config**

```go
// internal/billing/config/config.go
package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type Server struct {
	URL    string                     `json:"url"    mapstructure:"url"    validate:"required,url" yaml:"url"`
	Health commonconfig.ServerAddress `json:"health" mapstructure:"health" validate:"required"     yaml:"health"`
	HTTP   commonconfig.ServerAddress `json:"http"   mapstructure:"http"   validate:"required"     yaml:"http"`
	GRPC   commonconfig.ServerAddress `json:"grpc"   mapstructure:"grpc"   validate:"required"     yaml:"grpc"`
}

type Services struct {
	Authorization commonconfig.Service `json:"authorization" mapstructure:"authorization" validate:"required" yaml:"authorization"`
}

type Stripe struct {
	SecretKey     string `default:"-" json:"-" mapstructure:"secret_key"     validate:"required" yaml:"secret_key"`
	WebhookSecret string `default:"-" json:"-" mapstructure:"webhook_secret" validate:"required" yaml:"webhook_secret"`
}

type Billing struct {
	Currency   string `default:"vnd" json:"currency"   mapstructure:"currency"     validate:"required"     yaml:"currency"`
	WebBaseURL string `default:"-"   json:"webBaseUrl" mapstructure:"web_base_url" validate:"required,url" yaml:"web_base_url"`
}

type Config struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Server   Server                `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Kafka    commonconfig.Kafka    `json:"kafka"    mapstructure:"kafka"    validate:"required"  yaml:"kafka"`
	Redis    commonconfig.Redis    `json:"redis"    mapstructure:"redis"    validate:"required"  yaml:"redis"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
	Services Services              `json:"services" mapstructure:"services" validate:"required"  yaml:"services"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Billing  Billing               `json:"billing"  mapstructure:"billing"  validate:"required"  yaml:"billing"`
}

func New(validate *validator.Validate, viper *viper.Viper) (*Config, error) {
	viper.SetEnvPrefix("notopia_billing")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("billing.notopia.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", viper.ConfigFileUsed()))
	}

	cfg := &Config{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("cannot set default config: %w", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from env or config file: %w", err)
	}
	if cfg.Server.HTTP.Port == 0 {
		cfg.Server.HTTP.Port = 8084
	}
	if cfg.Server.GRPC.Port == 0 {
		cfg.Server.GRPC.Port = 18084
	}
	if cfg.Server.Health.Port == 0 {
		cfg.Server.Health.Port = 28084
	}
	if cfg.Kafka.ConsumerGroup == "" {
		cfg.Kafka.ConsumerGroup = "billing-service"
	}
	slog.Info("configuration", slog.Any("config", cfg))
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

var Provide = New
```

`viper.go`: copy `internal/note/config/viper.go` (package `config`).

```go
// internal/billing/config/wire.go
package config

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(new(*Config), "Billing", "Database", "General", "Identity", "Kafka", "Log", "Redis", "Server", "Services", "Stripe"),
)
```

`component/validator.go` and `component/wire.go`: copy note's `validator.go`; `ProviderSet = wire.NewSet(ProvideValidate)`.

`component/eventtopic.go` is the single place that declares every Kafka topic the service touches, split into published and consumed. The notifier (publish) and `controller/event` (subscribe) import these constants and never inline a topic string. Note keeps its outbound topics in `component/eventtopic.go` but inlines inbound ones in `controller/event`; billing puts both in one file so a topic rename is a one-file diff. Payload types stay in `api/share`.

```go
// internal/billing/component/eventtopic.go
package component

const (
	TopicSubscriptionChanged = "events.integration.billing.subscription.changed"
)

const (
	TopicWorkspaceDeleted = "events.integration.note.workspace.deleted"
	TopicRoleUpdated      = "events.integration.authorization.user_workspace_role_updated"
	TopicMemberRemoved    = "events.integration.authorization.workspace_member_removed"
)
```

- [ ] **Step 2: App skeleton + ports**

```go
// internal/billing/app/app.go
package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(traceProvider trace.TracerProvider, logger *slog.Logger) *HandlerProvider {
	return (*HandlerProvider)(commonhandler.NewHandlerProvider(
		commonhandler.WithTracer(traceProvider.Tracer("billing-app")),
		commonhandler.WithLogger(logger),
	))
}

var ProvideHandlerProvider = NewHandlerProvider

type Cmds struct{}

func NewCmds(handlerProvider *HandlerProvider) *Cmds {
	return &Cmds{}
}

var ProvideCmds = NewCmds

type Queries struct{}

func NewQueries(handlerProvider *HandlerProvider) *Queries {
	return &Queries{}
}

var ProvideQueries = NewQueries

type Server struct {
	Cmds    *Cmds
	Queries *Queries
}
```

Later tasks add fields to `Cmds`/`Queries` and constructor params exactly like `internal/note/app/app.go` does (`commonhandler.DecorateCmd(hp, h)` / `DecorateQuery(hp, h)` with `hp := (*commonhandler.HandlerProvider)(handlerProvider)`), and add `ProvideXHandler` to `wire.go`.

```go
// internal/billing/app/wire.go
package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,
	wire.Struct(new(Server), "*"),
)
```

```go
// internal/billing/app/svc_authorization.go
package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type WorkspacePermission uint8

const (
	WorkspacePermissionUnspecified WorkspacePermission = iota
	WorkspacePermissionBillingRead
	WorkspacePermissionBillingManage
)

func (p WorkspacePermission) IsValid() bool {
	switch p {
	case WorkspacePermissionBillingRead, WorkspacePermissionBillingManage:
		return true
	default:
		return false
	}
}

func (p WorkspacePermission) String() string {
	switch p {
	case WorkspacePermissionUnspecified:
		return "unspecified"
	case WorkspacePermissionBillingRead:
		return "billing_read"
	case WorkspacePermissionBillingManage:
		return "billing_manage"
	default:
		return "unknown"
	}
}

type WorkspaceRole uint8

const (
	WorkspaceRoleNone WorkspaceRole = iota
	WorkspaceRoleOwner
	WorkspaceRoleEditor
	WorkspaceRoleViewer
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case WorkspaceRoleNone, WorkspaceRoleOwner, WorkspaceRoleEditor, WorkspaceRoleViewer:
		return true
	default:
		return false
	}
}

func (r WorkspaceRole) String() string {
	switch r {
	case WorkspaceRoleNone:
		return "none"
	case WorkspaceRoleOwner:
		return "owner"
	case WorkspaceRoleEditor:
		return "editor"
	case WorkspaceRoleViewer:
		return "viewer"
	default:
		return "unknown"
	}
}

type AuthorizationSvc interface {
	HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission WorkspacePermission) (bool, error)
	// GetMemberRole returns WorkspaceRoleNone when memberID is not a member. callerID must be a member.
	GetMemberRole(ctx context.Context, callerID string, workspaceID uuid.UUID, memberID string) (WorkspaceRole, error)
}

func requirePermission(ctx context.Context, authz AuthorizationSvc, userID string, workspaceID uuid.UUID, p WorkspacePermission) error {
	ok, err := authz.HasWorkspacePermission(ctx, userID, workspaceID, p)
	if err != nil {
		return err
	}
	if !ok {
		return errs.NewForbidden(fmt.Sprintf("user %s lacks billing permission on workspace %s", userID, workspaceID))
	}
	return nil
}
```

```go
// internal/billing/app/svc_notifier.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

// Notifier publishes best-effort notifications after commit. Implementations log errors and never fail.
type Notifier interface {
	SubscriptionChanged(ctx context.Context, workspaceID uuid.UUID, plan domain.PlanTier, status domain.Status)
}

func notifySubscription(ctx context.Context, n Notifier, s *domain.Subscription) {
	if s == nil {
		return
	}
	n.SubscriptionChanged(ctx, s.WorkspaceID(), s.Plan(), s.Status())
}
```

- [ ] **Step 3: Infra adapters (authorization client, Redis, notifier)**

`internal/billing/infra/service/authorization.go`: copy the client setup from `internal/note/infra/service/authorization.go` (otelgrpc stats handler, logging interceptor, insecure creds, `servicesCfg.Authorization.URL`, error interceptor wrapping in `errs.New(errs.CodeAuthorizationService, "authorization service error", err)`), then implement:

```go
func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, p app.WorkspacePermission) (bool, error) {
	perm, err := toPBWorkspacePermission(p)
	if err != nil {
		return false, err
	}
	res, err := a.client.HasWorkspacePermission(ctx, &pb.HasWorkspacePermissionRequest{
		MemberId: userID, WorkspaceId: workspaceID.String(), Permission: perm,
	})
	if err != nil {
		return false, err
	}
	return res.HasPermission, nil
}

func (a *Authorization) GetMemberRole(ctx context.Context, callerID string, workspaceID uuid.UUID, memberID string) (app.WorkspaceRole, error) {
	res, err := a.client.GetWorkspaceMembers(ctx, &pb.GetWorkspaceMembersRequest{UserId: callerID, WorkspaceId: workspaceID.String()})
	if err != nil {
		return app.WorkspaceRoleNone, err
	}
	for _, m := range res.Members {
		if m.Id == memberID {
			return toAppWorkspaceRole(m.Role), nil
		}
	}
	return app.WorkspaceRoleNone, nil
}

func toPBWorkspacePermission(p app.WorkspacePermission) (pb.WorkspacePermission, error) {
	switch p {
	case app.WorkspacePermissionBillingRead:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_BILLING_READ, nil
	case app.WorkspacePermissionBillingManage:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_BILLING_MANAGE, nil
	default:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace permission: %v", p), nil)
	}
}

func toAppWorkspaceRole(role pb.WorkspaceRole) app.WorkspaceRole {
	switch role {
	case pb.WorkspaceRole_WORKSPACE_ROLE_OWNER:
		return app.WorkspaceRoleOwner
	case pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR:
		return app.WorkspaceRoleEditor
	case pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER:
		return app.WorkspaceRoleViewer
	default:
		return app.WorkspaceRoleNone
	}
}
```

`var ProvideAuthorization = NewAuthorization` and `var _ app.AuthorizationSvc = (*Authorization)(nil)`.

```go
// internal/billing/infra/redis/redis.go
package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func NewClient(ctx context.Context, cfg *commonconfig.Redis) (*goredis.Client, func(), error) {
	client := goredis.NewClient(&goredis.Options{Addr: cfg.Addr})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return client, func() { _ = client.Close() }, nil
}

var ProvideClient = NewClient
```

```go
// internal/billing/infra/notifier/kafka.go
package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

type Kafka struct {
	publisher message.Publisher
}

var _ app.Notifier = (*Kafka)(nil)

func NewKafka(cfg *commonconfig.Kafka, logger watermill.LoggerAdapter, tracer kafka.SaramaTracer, serviceName metadata.ServiceName) (*Kafka, func(), error) {
	publisher, err := kafka.NewPublisher(kafka.PublisherConfig{Brokers: cfg.Brokers, Tracer: tracer}, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}
	p := wotel.NewNamedPublisherDecorator(serviceName.String(), publisher)
	return &Kafka{publisher: p}, func() { _ = p.Close() }, nil
}

var ProvideKafka = NewKafka

func (k *Kafka) publish(ctx context.Context, topic string, payload any) {
	body, err := json.Marshal(payload)
	if err == nil {
		err = k.publisher.Publish(topic, message.NewMessageWithContext(ctx, watermill.NewUUID(), body))
	}
	if err != nil {
		slog.WarnContext(ctx, "billing notification not published", slog.String("topic", topic), slog.Any("error", err))
	}
}

func (k *Kafka) SubscriptionChanged(ctx context.Context, workspaceID uuid.UUID, plan domain.PlanTier, status domain.Status) {
	sharePlan, err := toSharePlan(plan)
	if err == nil {
		var shareStatus string
		if shareStatus, err = toShareStatus(status); err == nil {
			k.publish(ctx, component.TopicSubscriptionChanged, &share.SubscriptionChangedEvent{WorkspaceId: workspaceID, Plan: sharePlan, Status: shareStatus})
			return
		}
	}
	slog.WarnContext(ctx, "billing notification not published", slog.String("topic", component.TopicSubscriptionChanged), slog.Any("error", err))
}
```

```go
// internal/billing/infra/notifier/map.go
package notifier

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func toSharePlan(t domain.PlanTier) (share.SubscriptionChangedEventPlan, error) {
	switch t {
	case domain.PlanTierFree:
		return share.SubscriptionChangedEventPlanFree, nil
	case domain.PlanTierPlus:
		return share.SubscriptionChangedEventPlanPlus, nil
	case domain.PlanTierPro:
		return share.SubscriptionChangedEventPlanPro, nil
	default:
		return "", fmt.Errorf("invalid plan tier: %v", t)
	}
}

func toShareStatus(s domain.Status) (string, error) {
	if !s.IsValid() {
		return "", fmt.Errorf("invalid subscription status: %v", s)
	}
	return s.String(), nil
}
```

(Plan 02 Task 5b names the `plan` consts with `x-enum-varnames`, so they are `share.SubscriptionChangedEventPlanFree`/`Plus`/`Pro` of type `share.SubscriptionChangedEventPlan`.)

```go
// internal/billing/infra/wire.go
package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/infra/notifier"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billing/infra/redis"
	"github.com/notopia-uit/notopia/internal/billing/infra/service"
)

var ProviderSet = wire.NewSet(
	persistence.ProviderSet,
	redis.ProvideClient,
	notifier.ProvideKafka,
	service.ProvideAuthorization,
	wire.Bind(new(app.Notifier), new(*notifier.Kafka)),
	wire.Bind(new(app.AuthorizationSvc), new(*service.Authorization)),
)
```

(Task 6 adds the Stripe adapter here.)

- [ ] **Step 4: Controllers**

HTTP — `internal/billing/controller/http/http.go`: copy `internal/note/controller/http/http.go` and change:

- imports `internal/billing/app`, `internal/billing/config`, `pkg/api/billing` (alias `billing`), plus `commonconfig`;
- `IHandler = billing.ServerInterface`, `IStrictHandler = billing.StrictServerInterface`;
- `StrictHandler` → `type StrictHandler struct { App *app.Server; Identity *commonconfig.Identity }`, constructor `NewStrictHandler(app *app.Server, identity *commonconfig.Identity) *StrictHandler`;
- `api.GetSpec(nil, api.BillingSpec)`;
- `RegisterRoutes`: `api.Use(commonhttp.GatewayUserAuth()); api.Use(captureRawBody()); api.Use(validateHandler); billing.RegisterHandlers(api, handler)`; ping route `/billing/ping`;
- `New(...)` uses `cfg.HTTP.Address()` as in note.

Add to `http.go`:

```go
const rawBodyKey = "billing.rawBody"

func captureRawBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/billing/webhooks/stripe" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Set(rawBodyKey, body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		c.Next()
	}
}

func rawBody(ctx context.Context) []byte {
	if c, ok := ctx.(*gin.Context); ok {
		if v, ok := c.Get(rawBodyKey); ok {
			if b, ok := v.([]byte); ok {
				return b
			}
		}
	}
	return nil
}
```

```go
// internal/billing/controller/http/user.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func currentUser(ctx context.Context) (*commonhttp.User, error) {
	u, ok := commonhttp.UserFromContext(ctx)
	if !ok || u.ID == "" {
		return nil, errs.New(errs.CodeUnauthorized, "unauthorized", nil)
	}
	return u, nil
}

func currentPayer(ctx context.Context) (app.Payer, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return app.Payer{}, err
	}
	name := u.Name
	if name == "" {
		name = u.Email
	}
	return app.Payer{ID: u.ID, Email: u.Email, Name: name}, nil
}

func (h *StrictHandler) requireAdmin(ctx context.Context) (*commonhttp.User, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !u.HasGroup(h.Identity.AdminGroup) {
		return nil, errs.NewForbidden("admin only")
	}
	return u, nil
}
```

```go
// internal/billing/controller/http/err.go
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func statusOf(code errs.Code) int {
	switch code {
	case errs.CodeUnauthorized:
		return http.StatusUnauthorized
	case errs.CodeForbidden, errs.CodeNotOwner:
		return http.StatusForbidden
	case errs.CodeInvalid, errs.CodeWebhookSignature:
		return http.StatusBadRequest
	case errs.CodeNotFound, errs.CodeNoBillingAccount:
		return http.StatusNotFound
	case errs.CodeSubscriptionAlreadyActive, errs.CodeNoActiveSubscription, errs.CodeRefundWindowExpired,
		errs.CodeAlreadyPayer, errs.CodePriceNotActive:
		return http.StatusConflict
	case errs.CodePaymentProvider:
		return http.StatusBadGateway
	case errs.CodeAuthorizationService:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func writeErr(c *gin.Context, status int, code, message string) {
	c.JSON(status, billing.Error{Code: code, Message: message})
}

func ginMiddlewareErrorHandler(c *gin.Context, message string, statusCode int) {
	code := errs.CodeInternal
	switch statusCode {
	case http.StatusUnauthorized:
		code = errs.CodeUnauthorized
	case http.StatusBadRequest:
		code = errs.CodeInvalid
	case http.StatusNotFound:
		code = errs.CodeNotFound
	}
	writeErr(c, statusCode, code.String(), message)
}

var _ ginmiddleware.ErrorHandler = ginMiddlewareErrorHandler

func strictHandlerRequestErrorHandler(c *gin.Context, err error) {
	writeErr(c, http.StatusBadRequest, errs.CodeInvalid.String(), err.Error())
}

func strictHandlerResponseErrorHandler(c *gin.Context, err error) {
	writeErr(c, http.StatusInternalServerError, errs.CodeInternal.String(), err.Error())
}

func strictHandlerError(c *gin.Context, err error) {
	var e errs.Error
	if errors.As(err, &e) {
		status := statusOf(e.Code())
		message := e.Message()
		if status == http.StatusInternalServerError {
			message = "internal error"
		}
		writeErr(c, status, e.Code().String(), message)
		return
	}
	writeErr(c, http.StatusInternalServerError, errs.CodeInternal.String(), "internal error")
}
```

Match the signatures of `strictHandler*` to what `billing.StrictGinServerOptions` expects (same as note). If `billing.Error` is named differently in the generated models (the common `Error` schema), use that name.

`internal/billing/controller/http/unimplemented.go`: one method per operation of `billing.StrictServerInterface`, each `return nil, errs.New(errs.CodeInternal, "not implemented", nil)`. Example:

```go
func (h *StrictHandler) ListPlans(ctx context.Context, _ billing.ListPlansRequestObject) (billing.ListPlansResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}
```

Operations: `ListPlans`, `GetWorkspaceEntitlements`, `GetWorkspaceBilling`, `CreateCheckoutSession`, `CreatePortalSession`, `PreviewCancelNow`, `CancelNow`, `ClaimPayer`, `AdminListPlans`, `AdminUpdatePlan`, `AdminGetPolicy`, `AdminUpdatePolicy`, `AdminListPrices`, `AdminCreatePrice`, `AdminArchivePrice`, `StripeWebhook`. **Later tasks move each method into a real file and delete its stub.**

`internal/billing/controller/http/wire.go`: `ProviderSet = wire.NewSet(ProvideStrictHandler, ProvideHandler, Provide, wire.Bind(new(IStrictHandler), new(*StrictHandler)))` — mirror note's `controller/http/wire.go`.

gRPC — `controller/grpc/grpc.go`: copy note's `grpc.go`; `ServiceServer{app *app.Server}`, `var _ pb.BillingServiceServer = (*ServiceServer)(nil)`, register with `pb.RegisterBillingServiceServer`. `controller/grpc/err.go`: copy note's `unaryErrorInterceptor`, mapping `errs.Error` codes to gRPC: notFound → `codes.NotFound`, forbidden → `PermissionDenied`, invalid → `InvalidArgument`, others → `Internal`. `controller/grpc/service.go` methods are added in Task 5 (until then embed `pb.UnimplementedBillingServiceServer` in `ServiceServer`; remove it in Task 5).

Health — `controller/health/health.go`: copy note's, with checks: `persistenceMigration` (startup); `persistenceConnection`, `http` (`http://{HTTP.Address()}/billing/ping`), `grpc` (dial), `redis` (`client.Ping`), `authorizationService` (`servicesCfg.Authorization.LiveURL`), `kafka` (broker dial) for readiness; routes `/billing/health/startup|ready|live`. Constructor params: `*persistence.Pg`, `*config.Server`, `*commonconfig.Kafka`, `*goredis.Client`, `*config.Services`.

`controller/wire.go`: `ProviderSet = wire.NewSet(grpc.ProviderSet, http.ProviderSet, health.ProviderSet)` (each sub-package has `ProviderSet` with its `Provide*` funcs, as in note).

- [ ] **Step 5: Server, wiring, entrypoint**

```go
// internal/billing/server.go
package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/controller/grpc"
	"github.com/notopia-uit/notopia/internal/billing/controller/health"
	"github.com/notopia-uit/notopia/internal/billing/controller/http"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/otel"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	persistence *persistence.Pg
	http        *http.HTTP
	grpc        *grpc.GRPC
	health      *health.Health
	logger      *slog.Logger
}

func NewServer(persistence *persistence.Pg, http *http.HTTP, grpc *grpc.GRPC, health *health.Health, logger *slog.Logger, _ otel.Global) *Server {
	slog.SetDefault(logger)
	return &Server{persistence: persistence, http: http, grpc: grpc, health: health, logger: logger}
}

var ProvideServer = NewServer

func (s *Server) Run(ctx context.Context) error {
	if err := s.persistence.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.http.Shutdown(c)
		}()
		if err := s.http.Run(); err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		go func() {
			<-ctx.Done()
			s.grpc.GracefulStop()
		}()
		if err := s.grpc.Run(); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		go func() {
			<-ctx.Done()
			c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.health.Shutdown(c)
		}()
		if err := s.health.Run(); err != nil {
			return fmt.Errorf("health server: %w", err)
		}
		return nil
	})
	return g.Wait()
}
```

(`http.Server.Shutdown` makes `ListenAndServe` return `http.ErrServerClosed`; treat it as nil like note does if note filters it — check note's `HTTP.Run`.)

```go
// internal/billing/wire.go
package billing

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/controller"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	app.ProviderSet,
	component.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	domain.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
```

`cmd/billing/main.go`, `wire.go`, `metadata.go`, `Dockerfile`: copy note's; replace `note` → `billing`, `NOTOPIA_NOTE_LOG_LEVEL` → `NOTOPIA_BILLING_LOG_LEVEL`, `ServiceName = "notopia-billing"`, Dockerfile `EXPOSE 8084 18084 28084`.

`cmd/billing/.env` (append to Task 3's goose lines):

```
NOTOPIA_BILLING_GENERAL_APP_ENV="development"
NOTOPIA_BILLING_LOG_LEVEL=debug
NOTOPIA_BILLING_LOG_PRETTY=true
NOTOPIA_BILLING_SERVER_URL=http://api.notopia.localhost
NOTOPIA_BILLING_DATABASE_URL=postgres://notopiauit:notopiauit@localhost:5435/billing?sslmode=disable
NOTOPIA_BILLING_KAFKA_BROKERS=localhost:19092
NOTOPIA_BILLING_KAFKA_CONSUMER_GROUP=billing-service
NOTOPIA_BILLING_REDIS_ADDR=localhost:6375
NOTOPIA_BILLING_SERVICES_AUTHORIZATION_URL=localhost:18089
NOTOPIA_BILLING_SERVICES_AUTHORIZATION_LIVE_URL=http://localhost:28089/authorization/health/live
NOTOPIA_BILLING_IDENTITY_SERVICE_ACCOUNT_USER_ID=1003
NOTOPIA_BILLING_BILLING_WEB_BASE_URL=http://web.notopia.localhost
OTEL_SERVICE_NAME="billing"
```

`cmd/billing/.env.local` (gitignored, not committed): `NOTOPIA_BILLING_STRIPE_SECRET_KEY=sk_test_...`, `NOTOPIA_BILLING_STRIPE_WEBHOOK_SECRET=whsec_...` (from `stripe listen`). Check how note's `run` target loads `.env`/`.env.local` (nx env loading) and mirror it.

- [ ] **Step 6: Compose + Traefik**

```yaml
# deploy/compose/services/billing.yaml
# vim:ft=yaml.docker-compose:
# yaml-language-server: $schema=https://raw.githubusercontent.com/compose-spec/compose-go/master/schema/compose-spec.json
services:
  billing_postgresql:
    container_name: notopia-billing_postgresql
    image: postgres:18.6-alpine3.23
    environment:
      POSTGRES_DB: ${BILLING_POSTGRESQL_DB:-billing}
      POSTGRES_USER: ${BILLING_POSTGRESQL_USER:-notopiauit}
      POSTGRES_PASSWORD: ${BILLING_POSTGRESQL_PASSWORD:-notopiauit}
    ports:
      - ${BILLING_POSTGRESQL_PORT:-5435}:5432
    volumes:
      - billing_postgresql_data:/var/lib/postgresql
    healthcheck:
      test:
        - CMD-SHELL
        - pg_isready -d $${POSTGRES_DB} -U $${POSTGRES_USER}
      start_period: 10s
      start_interval: 1s
    restart: unless-stopped

  billing_redis:
    container_name: notopia-billing_redis
    image: redis:8.4.3-alpine3.22
    command: ['redis-server', '--appendonly', 'yes', '--maxmemory-policy', 'noeviction']
    ports:
      - 6375:6379

volumes:
  billing_postgresql_data:
```

`deploy/compose/compose.yaml`: add `- ./services/billing.yaml` after `authorization.yaml`.

```yaml
# deploy/compose/infrastructure/traefik/dynamic/billing.yaml
# yaml-language-server: $schema=https://www.schemastore.org/traefik-v3-file-provider.json
http:
  routers:
    billing:
      entryPoints:
        - web
      rule: Host(`api.notopia.localhost`) && PathPrefix(`/billing`) && !PathPrefix(`/billing/health`) && !Path(`/billing/webhooks/stripe`) && !(Path(`/billing/plans`) && Method(`GET`))
      service: billing
      middlewares:
        - cors
        - jwt
    billing-public:
      entryPoints:
        - web
      rule: Host(`api.notopia.localhost`) && (Path(`/billing/webhooks/stripe`) || (Path(`/billing/plans`) && Method(`GET`)))
      service: billing
      middlewares:
        - cors
    billing-health:
      entryPoints:
        - web
      rule: Host(`api.notopia.localhost`) && PathPrefix(`/billing/health`)
      service: billing-health

  services:
    billing:
      loadBalancer:
        servers:
          - url: http://host.docker.internal:8084
    billing-health:
      loadBalancer:
        servers:
          - url: http://host.docker.internal:28084
```

- [ ] **Step 7: Generate, build, run**

```bash
nx run billing:gen:wire && nx build billing && nx lint billing
docker compose -f deploy/compose/compose.yaml up -d billing_postgresql billing_redis
nx run billing:run   # separate terminal; needs authorization + redpanda up
curl -s http://localhost:28084/billing/health/ready
curl -s http://api.notopia.localhost/billing/health/live
```

Expected: ready returns all checks `up`; migrations applied (`\dt` in `billing` DB lists the 7 tables + `goose_db_version`); traces for the health/HTTP calls appear in Tempo (Grafana).

- [ ] **Step 8: AGENTS.md + commit**

`internal/billing/AGENTS.md`: purpose, layer table (copy note's, minus outbox; keep pgreadmodel), Stripe notes (sandbox only, fetch-and-sync webhooks, idempotency keys), crons, "Notifications are best-effort", TODO list (free trial; live payments via a VN gateway behind `PaymentProvider`). `ln -s AGENTS.md internal/billing/CLAUDE.md`. Add `billing/` lines to the root `CLAUDE.md` structure block and the scoped-docs table.

```bash
git add internal/billing cmd/billing deploy/compose CLAUDE.md
git commit -m "feat(billing): service runtime, config, health, compose and traefik routes"
```

---

### Task 5: Default catalog, entitlement queries, gRPC, read endpoints

**Files:**

- Create: `internal/billing/app/{model.go,mapper.go,common.go,cmd_ensuredefaults.go,q_getentitlements.go,q_batchgetentitlements.go,q_listplans.go,q_getworkspacebilling.go,q_getworkspacepayer.go}`
- Modify: `internal/billing/app/app.go`, `internal/billing/app/wire.go`, `internal/billing/server.go`
- Create: `internal/billing/controller/grpc/{service.go,mapout.go}`
- Create: `internal/billing/controller/http/{mapout.go,plan.go,workspace.go}`; modify `unimplemented.go` (delete moved stubs)

**Interfaces:**

- Consumes: `domain.UnitOfWork` (commands), `app.BillingReadModel` (queries), `app.AuthorizationSvc`.
- Produces: `Cmds.EnsureDefaults`, `Queries.GetEntitlements` (`GetEntitlements{WorkspaceID uuid.UUID; UserID string; Internal bool}` → `app.Entitlements`), `Queries.BatchGetEntitlements` (`[]uuid.UUID` → `map[uuid.UUID]app.Entitlements`), `Queries.ListPlans` (→ `[]PlanView`), `Queries.GetWorkspaceBilling` (→ `WorkspaceBillingView`), `Queries.GetWorkspacePayer` (→ `PayerView`).
- App models (`app/model.go`) are the app's own response shapes and enums (no `domain` import); `app/mapper.go` converts domain values into them. Controllers only see `app.*` types.
- Mappers: gRPC `controller/grpc/mapout.go` (`toPBPlanTier`, `toPBStatus`, `toPBEntitlements`); HTTP `controller/http/mapout.go` (`toPlanTierDTO`, `toBillingIntervalDTO`, `toSubscriptionStatusDTO`, `toRefundReasonDTO`, `toRefundStatusDTO`, `toFeatures`, `toLimits`, and `toEntitlements`/`toPrice`/`toPlan`/`toWorkspaceBilling`, which now return `(T, error)`).

Default plans and policy are inserted on startup (no Stripe needed). Prices need Stripe and come from the seeder (Task 7).

- [ ] **Step 1: App models + helpers**

```go
// internal/billing/app/model.go
package app

import (
	"time"

	"github.com/google/uuid"
)

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

func (t PlanTier) IsValid() bool { return t >= PlanTierFree && t <= PlanTierPro }

func (t PlanTier) String() string {
	switch t {
	case PlanTierFree:
		return "free"
	case PlanTierPlus:
		return "plus"
	case PlanTierPro:
		return "pro"
	default:
		return "unspecified"
	}
}

type BillingInterval uint8

const (
	BillingIntervalUnspecified BillingInterval = iota
	BillingIntervalMonth
	BillingIntervalYear
	BillingIntervalLifetime
)

func (i BillingInterval) IsValid() bool {
	return i >= BillingIntervalMonth && i <= BillingIntervalLifetime
}

func (i BillingInterval) String() string {
	switch i {
	case BillingIntervalMonth:
		return "month"
	case BillingIntervalYear:
		return "year"
	case BillingIntervalLifetime:
		return "lifetime"
	default:
		return "unspecified"
	}
}

// SubscriptionStatusNone is the zero value and a real state: the workspace has no subscription.
type SubscriptionStatus uint8

const (
	SubscriptionStatusNone SubscriptionStatus = iota
	SubscriptionStatusIncomplete
	SubscriptionStatusActive
	SubscriptionStatusPastDue
	SubscriptionStatusUnpaid
	SubscriptionStatusCanceled
	SubscriptionStatusIncompleteExpired
)

func (s SubscriptionStatus) IsValid() bool { return s <= SubscriptionStatusIncompleteExpired }

func (s SubscriptionStatus) String() string {
	switch s {
	case SubscriptionStatusNone:
		return "none"
	case SubscriptionStatusIncomplete:
		return "incomplete"
	case SubscriptionStatusActive:
		return "active"
	case SubscriptionStatusPastDue:
		return "past_due"
	case SubscriptionStatusUnpaid:
		return "unpaid"
	case SubscriptionStatusCanceled:
		return "canceled"
	case SubscriptionStatusIncompleteExpired:
		return "incomplete_expired"
	default:
		return "unknown"
	}
}

type RefundReason uint8

const (
	RefundReasonUnspecified RefundReason = iota
	RefundReasonCancelNow
	RefundReasonLifetimeWindow
	RefundReasonReplacedByLifetime
	RefundReasonWorkspaceDeleted
	RefundReasonDuplicateCheckout
	RefundReasonExternal
)

func (r RefundReason) IsValid() bool { return r >= RefundReasonCancelNow && r <= RefundReasonExternal }

func (r RefundReason) String() string {
	switch r {
	case RefundReasonCancelNow:
		return "cancel_now"
	case RefundReasonLifetimeWindow:
		return "lifetime_window"
	case RefundReasonReplacedByLifetime:
		return "replaced_by_lifetime"
	case RefundReasonWorkspaceDeleted:
		return "workspace_deleted"
	case RefundReasonDuplicateCheckout:
		return "duplicate_checkout"
	case RefundReasonExternal:
		return "external"
	default:
		return "unspecified"
	}
}

type RefundStatus uint8

const (
	RefundStatusUnspecified RefundStatus = iota
	RefundStatusPending
	RefundStatusSucceeded
	RefundStatusFailed
	RefundStatusCanceled
)

func (s RefundStatus) IsValid() bool { return s >= RefundStatusPending && s <= RefundStatusCanceled }

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusPending:
		return "pending"
	case RefundStatusSucceeded:
		return "succeeded"
	case RefundStatusFailed:
		return "failed"
	case RefundStatusCanceled:
		return "canceled"
	default:
		return "unspecified"
	}
}

// Feature and limit keys are plain strings here: both wires (proto map, OpenAPI object) key them by string.
type Plan struct {
	Tier     PlanTier
	Features map[string]bool
	Limits   map[string]int64
}

type Price struct {
	ID         uuid.UUID
	Plan       PlanTier
	Interval   BillingInterval
	Amount     int64
	Currency   string
	LookupKey  string
	Active     bool
	CreatedAt  time.Time
	ArchivedAt *time.Time
}

type PlanView struct {
	Plan   Plan
	Prices []Price
}

type BillingPolicy struct {
	PastDueGraceDays         int
	LifetimeRefundWindowDays int
}

type Entitlements struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
	Status      SubscriptionStatus
	AccessUntil *time.Time
	Features    map[string]bool
	Limits      map[string]int64
}

type Refund struct {
	ID        uuid.UUID
	Amount    int64
	Currency  string
	Reason    RefundReason
	Status    RefundStatus
	CreatedAt time.Time
}

type Subscription struct {
	ID                 uuid.UUID
	Plan               PlanTier
	Interval           BillingInterval
	Status             SubscriptionStatus
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	AccessUntil        *time.Time
	Refunds            []Refund
}

type Payer struct {
	ID           string
	Email        string
	Name         string
	MissingSince *time.Time
}

type WorkspaceBillingView struct {
	Subscription *Subscription
	Payer        *Payer
}

type PayerView struct {
	PayerID       *string
	RequiresPayer bool
}

type RefundQuoteLine struct {
	PaidAmount      int64
	Amount          int64
	CoverageSeconds int64
	UnusedSeconds   int64
}

type RefundQuote struct {
	Amount        int64
	Currency      string
	UnusedSeconds int64
	PeriodSeconds int64
	Lines         []RefundQuoteLine
}
```

`app/model.go` imports nothing from `domain`: these are the application's own read/response shapes, so they can diverge from the aggregates. `app/mapper.go` is the only place that converts a domain value into them (explicit switches per enum, like every other port):

```go
// internal/billing/app/mapper.go
package app

import "github.com/notopia-uit/notopia/internal/billing/domain"

func newPlanTier(t domain.PlanTier) PlanTier {
	switch t {
	case domain.PlanTierFree:
		return PlanTierFree
	case domain.PlanTierPlus:
		return PlanTierPlus
	case domain.PlanTierPro:
		return PlanTierPro
	default:
		return PlanTierUnspecified
	}
}

func newBillingInterval(i domain.Interval) BillingInterval {
	switch i {
	case domain.IntervalMonth:
		return BillingIntervalMonth
	case domain.IntervalYear:
		return BillingIntervalYear
	case domain.IntervalLifetime:
		return BillingIntervalLifetime
	default:
		return BillingIntervalUnspecified
	}
}

func newSubscriptionStatus(s domain.Status) SubscriptionStatus {
	switch s {
	case domain.StatusIncomplete:
		return SubscriptionStatusIncomplete
	case domain.StatusActive:
		return SubscriptionStatusActive
	case domain.StatusPastDue:
		return SubscriptionStatusPastDue
	case domain.StatusUnpaid:
		return SubscriptionStatusUnpaid
	case domain.StatusCanceled:
		return SubscriptionStatusCanceled
	case domain.StatusIncompleteExpired:
		return SubscriptionStatusIncompleteExpired
	default:
		return SubscriptionStatusNone
	}
}

func newRefundReason(r domain.RefundReason) RefundReason {
	switch r {
	case domain.RefundReasonCancelNow:
		return RefundReasonCancelNow
	case domain.RefundReasonLifetimeWindow:
		return RefundReasonLifetimeWindow
	case domain.RefundReasonReplacedByLifetime:
		return RefundReasonReplacedByLifetime
	case domain.RefundReasonWorkspaceDeleted:
		return RefundReasonWorkspaceDeleted
	case domain.RefundReasonDuplicateCheckout:
		return RefundReasonDuplicateCheckout
	case domain.RefundReasonExternal:
		return RefundReasonExternal
	default:
		return RefundReasonUnspecified
	}
}

func newRefundStatus(s domain.RefundStatus) RefundStatus {
	switch s {
	case domain.RefundStatusPending:
		return RefundStatusPending
	case domain.RefundStatusSucceeded:
		return RefundStatusSucceeded
	case domain.RefundStatusFailed:
		return RefundStatusFailed
	case domain.RefundStatusCanceled:
		return RefundStatusCanceled
	default:
		return RefundStatusUnspecified
	}
}

func newFeatures(m map[domain.FeatureKey]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return out
}

func newLimits(m map[domain.LimitKey]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return out
}

func newPlan(p *domain.Plan) Plan {
	return Plan{Tier: newPlanTier(p.Tier()), Features: newFeatures(p.Features()), Limits: newLimits(p.Limits())}
}

func newPrice(p *domain.Price) Price {
	return Price{
		ID: p.ID(), Plan: newPlanTier(p.Plan()), Interval: newBillingInterval(p.Interval()),
		Amount: p.Amount().Amount, Currency: p.Amount().Currency, LookupKey: p.LookupKey(), Active: p.Active(),
		CreatedAt: p.CreatedAt(), ArchivedAt: p.ArchivedAt(),
	}
}

func newPrices(ps []*domain.Price) []Price {
	out := make([]Price, 0, len(ps))
	for _, p := range ps {
		out = append(out, newPrice(p))
	}
	return out
}

func newBillingPolicy(p *domain.BillingPolicy) BillingPolicy {
	return BillingPolicy{PastDueGraceDays: p.PastDueGraceDays(), LifetimeRefundWindowDays: p.LifetimeRefundWindowDays()}
}

func newEntitlements(e domain.Entitlements) Entitlements {
	return Entitlements{
		WorkspaceID: e.WorkspaceID, Plan: newPlanTier(e.Plan), Status: newSubscriptionStatus(e.Status),
		AccessUntil: e.AccessUntil, Features: newFeatures(e.Features), Limits: newLimits(e.Limits),
	}
}

func newSubscription(s *domain.Subscription, policy *domain.BillingPolicy) Subscription {
	refunds := make([]Refund, 0, len(s.Refunds()))
	for _, r := range s.Refunds() {
		refunds = append(refunds, Refund{
			ID: r.ID(), Amount: r.Amount().Amount, Currency: r.Amount().Currency,
			Reason: newRefundReason(r.Reason()), Status: newRefundStatus(r.Status()), CreatedAt: r.CreatedAt(),
		})
	}
	return Subscription{
		ID: s.ID(), Plan: newPlanTier(s.Plan()), Interval: newBillingInterval(s.Interval()), Status: newSubscriptionStatus(s.Status()),
		CurrentPeriodStart: s.CurrentPeriodStart(), CurrentPeriodEnd: s.CurrentPeriodEnd(), CancelAtPeriodEnd: s.CancelAtPeriodEnd(),
		CanceledAt: s.CanceledAt(), AccessUntil: s.AccessUntil(policy), Refunds: refunds,
	}
}

func newPayer(a *domain.BillingAccount) Payer {
	p := a.Payer()
	return Payer{ID: p.ID, Email: p.Email, Name: p.Name, MissingSince: a.PayerMissingSince()}
}

func newRefundQuote(q domain.RefundQuote) RefundQuote {
	lines := make([]RefundQuoteLine, 0, len(q.Lines))
	for _, l := range q.Lines {
		lines = append(lines, RefundQuoteLine{PaidAmount: l.PaidAmount.Amount, Amount: l.Amount.Amount, CoverageSeconds: l.CoverageSeconds, UnusedSeconds: l.UnusedSeconds})
	}
	return RefundQuote{Amount: q.Amount, Currency: q.Currency, UnusedSeconds: q.UnusedSeconds, PeriodSeconds: q.PeriodSeconds, Lines: lines}
}

func toDomainPlanTier(t PlanTier) domain.PlanTier {
	switch t {
	case PlanTierFree:
		return domain.PlanTierFree
	case PlanTierPlus:
		return domain.PlanTierPlus
	case PlanTierPro:
		return domain.PlanTierPro
	default:
		return domain.PlanTierUnspecified
	}
}

func toDomainInterval(i BillingInterval) domain.Interval {
	switch i {
	case BillingIntervalMonth:
		return domain.IntervalMonth
	case BillingIntervalYear:
		return domain.IntervalYear
	case BillingIntervalLifetime:
		return domain.IntervalLifetime
	default:
		return domain.IntervalUnspecified
	}
}

func toDomainPayer(p Payer) domain.Payer {
	return domain.Payer{ID: p.ID, Email: p.Email, Name: p.Name}
}

// parseFeatures/parseLimits reject keys the domain does not know (errs.NewInvalid → 400).
func parseFeatures(in map[string]bool) (map[domain.FeatureKey]bool, error) {
	out := make(map[domain.FeatureKey]bool, len(in))
	for s, v := range in {
		k, err := domain.ParseFeatureKey(s)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

func parseLimits(in map[string]int64) (map[domain.LimitKey]int64, error) {
	out := make(map[domain.LimitKey]int64, len(in))
	for s, v := range in {
		k, err := domain.ParseLimitKey(s)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}
```

```go
// internal/billing/app/common.go
package app

import (
	"errors"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func hasCode(err error, code errs.Code) bool {
	var e errs.Error
	return errors.As(err, &e) && e.Code() == code
}
```

- [ ] **Step 2: Handlers**

```go
// internal/billing/app/cmd_ensuredefaults.go
package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type EnsureDefaults struct{}

type EnsureDefaultsHandler struct{ uow domain.UnitOfWork }

func NewEnsureDefaultsHandler(uow domain.UnitOfWork) *EnsureDefaultsHandler {
	return &EnsureDefaultsHandler{uow: uow}
}

var ProvideEnsureDefaultsHandler = NewEnsureDefaultsHandler

type EnsureDefaultsCmd commonhandler.Cmd[EnsureDefaults]

var _ EnsureDefaultsCmd = (*EnsureDefaultsHandler)(nil)

func (h *EnsureDefaultsHandler) Handle(ctx context.Context, _ *EnsureDefaults) error {
	now := time.Now()
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		plans, err := r.Plan().ListMap(ctx)
		if err != nil {
			return err
		}
		for _, p := range domain.DefaultPlans(now) {
			if _, ok := plans[p.Tier()]; ok {
				continue
			}
			if err := r.Plan().Save(ctx, p); err != nil {
				return err
			}
		}
		if _, err := r.BillingPolicy().Get(ctx, false); err != nil {
			if !hasCode(err, errs.CodeNotFound) {
				return err
			}
			return r.BillingPolicy().Save(ctx, domain.DefaultBillingPolicy(now))
		}
		return nil
	})
}
```

```go
// internal/billing/app/q_getentitlements.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetEntitlements struct {
	WorkspaceID uuid.UUID
	UserID      string
	Internal    bool // gRPC from another service: skip the permission check
}

type GetEntitlementsHandler struct {
	read  BillingReadModel
	authz AuthorizationSvc
}

func NewGetEntitlementsHandler(read BillingReadModel, authz AuthorizationSvc) *GetEntitlementsHandler {
	return &GetEntitlementsHandler{read: read, authz: authz}
}

var ProvideGetEntitlementsHandler = NewGetEntitlementsHandler

type GetEntitlementsQuery commonhandler.Query[GetEntitlements, Entitlements]

var _ GetEntitlementsQuery = (*GetEntitlementsHandler)(nil)

func (h *GetEntitlementsHandler) Handle(ctx context.Context, q *GetEntitlements) (Entitlements, error) {
	if !q.Internal {
		if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingRead); err != nil {
			return Entitlements{}, err
		}
	}
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return Entitlements{}, err
	}
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return Entitlements{}, err
	}
	policy, err := h.read.Policy(ctx)
	if err != nil {
		return Entitlements{}, err
	}
	return newEntitlements(domain.ResolveEntitlements(q.WorkspaceID, sub, plans, policy, time.Now())), nil
}
```

```go
// internal/billing/app/q_batchgetentitlements.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type BatchGetEntitlements struct {
	WorkspaceIDs []uuid.UUID
}

type BatchGetEntitlementsHandler struct{ read BillingReadModel }

func NewBatchGetEntitlementsHandler(read BillingReadModel) *BatchGetEntitlementsHandler {
	return &BatchGetEntitlementsHandler{read: read}
}

var ProvideBatchGetEntitlementsHandler = NewBatchGetEntitlementsHandler

type BatchGetEntitlementsQuery commonhandler.Query[BatchGetEntitlements, map[uuid.UUID]Entitlements]

var _ BatchGetEntitlementsQuery = (*BatchGetEntitlementsHandler)(nil)

func (h *BatchGetEntitlementsHandler) Handle(ctx context.Context, q *BatchGetEntitlements) (map[uuid.UUID]Entitlements, error) {
	subs, err := h.read.CurrentSubscriptions(ctx, q.WorkspaceIDs)
	if err != nil {
		return nil, err
	}
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := h.read.Policy(ctx)
	if err != nil {
		return nil, err
	}
	byWorkspace := make(map[uuid.UUID]*domain.Subscription, len(subs))
	for _, s := range subs {
		byWorkspace[s.WorkspaceID()] = s
	}
	now := time.Now()
	out := make(map[uuid.UUID]Entitlements, len(q.WorkspaceIDs))
	for _, id := range q.WorkspaceIDs {
		out[id] = newEntitlements(domain.ResolveEntitlements(id, byWorkspace[id], plans, policy, now))
	}
	return out, nil
}
```

```go
// internal/billing/app/q_listplans.go
package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ListPlans struct{}

type ListPlansHandler struct{ read BillingReadModel }

func NewListPlansHandler(read BillingReadModel) *ListPlansHandler { return &ListPlansHandler{read: read} }

var ProvideListPlansHandler = NewListPlansHandler

type ListPlansQuery commonhandler.Query[ListPlans, []PlanView]

var _ ListPlansQuery = (*ListPlansHandler)(nil)

func (h *ListPlansHandler) Handle(ctx context.Context, _ *ListPlans) ([]PlanView, error) {
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return nil, err
	}
	prices, err := h.read.ActivePrices(ctx)
	if err != nil {
		return nil, err
	}
	var views []PlanView
	for _, tier := range domain.PlanTiers {
		p, ok := plans[tier]
		if !ok {
			continue
		}
		v := PlanView{Plan: newPlan(p)}
		for _, pr := range prices {
			if pr.Plan() == tier {
				v.Prices = append(v.Prices, newPrice(pr))
			}
		}
		views = append(views, v)
	}
	return views, nil
}
```

```go
// internal/billing/app/q_getworkspacebilling.go
package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspaceBilling struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type GetWorkspaceBillingHandler struct {
	read  BillingReadModel
	authz AuthorizationSvc
}

func NewGetWorkspaceBillingHandler(read BillingReadModel, authz AuthorizationSvc) *GetWorkspaceBillingHandler {
	return &GetWorkspaceBillingHandler{read: read, authz: authz}
}

var ProvideGetWorkspaceBillingHandler = NewGetWorkspaceBillingHandler

type GetWorkspaceBillingQuery commonhandler.Query[GetWorkspaceBilling, WorkspaceBillingView]

var _ GetWorkspaceBillingQuery = (*GetWorkspaceBillingHandler)(nil)

func (h *GetWorkspaceBillingHandler) Handle(ctx context.Context, q *GetWorkspaceBilling) (WorkspaceBillingView, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return WorkspaceBillingView{}, err
	}
	var v WorkspaceBillingView
	account, err := h.read.Account(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	if account != nil {
		payer := newPayer(account)
		v.Payer = &payer
	}
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	if sub != nil {
		policy, err := h.read.Policy(ctx)
		if err != nil {
			return v, err
		}
		view := newSubscription(sub, policy)
		v.Subscription = &view
	}
	return v, nil
}
```

```go
// internal/billing/app/q_getworkspacepayer.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspacePayer struct {
	WorkspaceID uuid.UUID
}

type GetWorkspacePayerHandler struct{ read BillingReadModel }

func NewGetWorkspacePayerHandler(read BillingReadModel) *GetWorkspacePayerHandler {
	return &GetWorkspacePayerHandler{read: read}
}

var ProvideGetWorkspacePayerHandler = NewGetWorkspacePayerHandler

type GetWorkspacePayerQuery commonhandler.Query[GetWorkspacePayer, PayerView]

var _ GetWorkspacePayerQuery = (*GetWorkspacePayerHandler)(nil)

func (h *GetWorkspacePayerHandler) Handle(ctx context.Context, q *GetWorkspacePayer) (PayerView, error) {
	var v PayerView
	account, err := h.read.Account(ctx, q.WorkspaceID)
	if err != nil || account == nil {
		return v, err
	}
	id := account.Payer().ID
	v.PayerID = &id
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	v.RequiresPayer = sub != nil && sub.Interval() != domain.IntervalLifetime && !sub.CancelAtPeriodEnd()
	return v, nil
}
```

Register all six in `app.go` (`Cmds.EnsureDefaults EnsureDefaultsCmd`; `Queries.GetEntitlements`, `BatchGetEntitlements`, `ListPlans`, `GetWorkspaceBilling`, `GetWorkspacePayer`) and their `Provide*Handler` in `wire.go`.

`server.go`: inject `app *app.Server` into `NewServer` and, right after `RunMigrations`, call
`if err := s.app.Cmds.EnsureDefaults.Handle(ctx, &app.EnsureDefaults{}); err != nil { return fmt.Errorf("failed to ensure billing defaults: %w", err) }`.

- [ ] **Step 3: gRPC service**

`plan`/`status`/`tier` are proto enums, mapped by explicit switches; an Unspecified domain value is an internal error (the proto rule `not_in: [0]` forbids sending it). Feature/limit map keys go out as `k.String()`. No billing request carries an enum (all take workspace ids), so there is no gRPC `mapin.go`; add one with `pgmapper.ToPlanTier(pb.PlanTier) (domain.PlanTier, error)` if a request ever gains an enum field.

```go
// internal/billing/controller/grpc/mapout.go
package grpc

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toPBPlanTier(t app.PlanTier) (pb.PlanTier, error) {
	switch t {
	case app.PlanTierFree:
		return pb.PlanTier_PLAN_TIER_FREE, nil
	case app.PlanTierPlus:
		return pb.PlanTier_PLAN_TIER_PLUS, nil
	case app.PlanTierPro:
		return pb.PlanTier_PLAN_TIER_PRO, nil
	default:
		return pb.PlanTier_PLAN_TIER_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid plan tier: %v", t), nil)
	}
}

func toPBStatus(s app.SubscriptionStatus) (pb.SubscriptionStatus, error) {
	switch s {
	case app.SubscriptionStatusNone:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_NONE, nil
	case app.SubscriptionStatusIncomplete:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE, nil
	case app.SubscriptionStatusActive:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE, nil
	case app.SubscriptionStatusPastDue:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE, nil
	case app.SubscriptionStatusUnpaid:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_UNPAID, nil
	case app.SubscriptionStatusCanceled:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED, nil
	case app.SubscriptionStatusIncompleteExpired:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE_EXPIRED, nil
	default:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid subscription status: %v", s), nil)
	}
}

func toPBEntitlements(e app.Entitlements) (*pb.Entitlements, error) {
	plan, err := toPBPlanTier(e.Plan)
	if err != nil {
		return nil, err
	}
	status, err := toPBStatus(e.Status)
	if err != nil {
		return nil, err
	}
	out := &pb.Entitlements{
		WorkspaceId: e.WorkspaceID.String(), Plan: plan, Status: status,
		Features: e.Features, Limits: e.Limits,
	}
	if e.AccessUntil != nil {
		out.AccessUntil = timestamppb.New(*e.AccessUntil)
	}
	return out, nil
}
```

```go
// internal/billing/controller/grpc/service.go
package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errs.NewInvalid("invalid uuid: " + s)
	}
	return id, nil
}

func (s *ServiceServer) GetWorkspaceEntitlements(ctx context.Context, req *pb.GetWorkspaceEntitlementsRequest) (*pb.GetWorkspaceEntitlementsResponse, error) {
	id, err := parseUUID(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	ent, err := s.app.Queries.GetEntitlements.Handle(ctx, &app.GetEntitlements{WorkspaceID: id, Internal: true})
	if err != nil {
		return nil, err
	}
	out, err := toPBEntitlements(ent)
	if err != nil {
		return nil, err
	}
	return &pb.GetWorkspaceEntitlementsResponse{Entitlements: out}, nil
}

func (s *ServiceServer) BatchGetWorkspaceEntitlements(ctx context.Context, req *pb.BatchGetWorkspaceEntitlementsRequest) (*pb.BatchGetWorkspaceEntitlementsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.WorkspaceIds))
	for _, raw := range req.WorkspaceIds {
		id, err := parseUUID(raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	res, err := s.app.Queries.BatchGetEntitlements.Handle(ctx, &app.BatchGetEntitlements{WorkspaceIDs: ids})
	if err != nil {
		return nil, err
	}
	out := make(map[string]*pb.Entitlements, len(res))
	for id, e := range res {
		if out[id.String()], err = toPBEntitlements(e); err != nil {
			return nil, err
		}
	}
	return &pb.BatchGetWorkspaceEntitlementsResponse{Entitlements: out}, nil
}

func (s *ServiceServer) GetWorkspacePayer(ctx context.Context, req *pb.GetWorkspacePayerRequest) (*pb.GetWorkspacePayerResponse, error) {
	id, err := parseUUID(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	v, err := s.app.Queries.GetWorkspacePayer.Handle(ctx, &app.GetWorkspacePayer{WorkspaceID: id})
	if err != nil {
		return nil, err
	}
	return &pb.GetWorkspacePayerResponse{PayerId: v.PayerID, RequiresPayer: v.RequiresPayer}, nil
}

func (s *ServiceServer) ListPlans(ctx context.Context, _ *pb.ListPlansRequest) (*pb.ListPlansResponse, error) {
	views, err := s.app.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out := &pb.ListPlansResponse{}
	for _, v := range views {
		tier, err := toPBPlanTier(v.Plan.Tier)
		if err != nil {
			return nil, err
		}
		out.Plans = append(out.Plans, &pb.BillingPlan{Tier: tier, Features: v.Plan.Features, Limits: v.Plan.Limits})
	}
	return out, nil
}
```

Remove `pb.UnimplementedBillingServiceServer` from `ServiceServer`.

- [ ] **Step 4: HTTP read endpoints**

```go
// internal/billing/controller/http/mapout.go
package http

import (
	"fmt"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/oapi-codegen/nullable"
)

func nullableTime(t *time.Time) nullable.Nullable[time.Time] {
	if t == nil {
		return nullable.NewNullNullable[time.Time]()
	}
	return nullable.NewNullableWithValue(*t)
}

func toPlanTierDTO(t app.PlanTier) (billing.PlanTier, error) {
	switch t {
	case app.PlanTierFree:
		return billing.PlanTierFree, nil
	case app.PlanTierPlus:
		return billing.PlanTierPlus, nil
	case app.PlanTierPro:
		return billing.PlanTierPro, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid plan tier: %v", t), nil)
	}
}

func toBillingIntervalDTO(i app.BillingInterval) (billing.BillingInterval, error) {
	switch i {
	case app.BillingIntervalMonth:
		return billing.BillingIntervalMonth, nil
	case app.BillingIntervalYear:
		return billing.BillingIntervalYear, nil
	case app.BillingIntervalLifetime:
		return billing.BillingIntervalLifetime, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid billing interval: %v", i), nil)
	}
}

func toSubscriptionStatusDTO(s app.SubscriptionStatus) (billing.SubscriptionStatus, error) {
	switch s {
	case app.SubscriptionStatusNone:
		return billing.SubscriptionStatusNone, nil
	case app.SubscriptionStatusIncomplete:
		return billing.SubscriptionStatusIncomplete, nil
	case app.SubscriptionStatusActive:
		return billing.SubscriptionStatusActive, nil
	case app.SubscriptionStatusPastDue:
		return billing.SubscriptionStatusPastDue, nil
	case app.SubscriptionStatusUnpaid:
		return billing.SubscriptionStatusUnpaid, nil
	case app.SubscriptionStatusCanceled:
		return billing.SubscriptionStatusCanceled, nil
	case app.SubscriptionStatusIncompleteExpired:
		return billing.SubscriptionStatusIncompleteExpired, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid subscription status: %v", s), nil)
	}
}

func toRefundReasonDTO(r app.RefundReason) (billing.RefundReason, error) {
	switch r {
	case app.RefundReasonCancelNow:
		return billing.RefundReasonCancelNow, nil
	case app.RefundReasonLifetimeWindow:
		return billing.RefundReasonLifetimeWindow, nil
	case app.RefundReasonReplacedByLifetime:
		return billing.RefundReasonReplacedByLifetime, nil
	case app.RefundReasonWorkspaceDeleted:
		return billing.RefundReasonWorkspaceDeleted, nil
	case app.RefundReasonDuplicateCheckout:
		return billing.RefundReasonDuplicateCheckout, nil
	case app.RefundReasonExternal:
		return billing.RefundReasonExternal, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid refund reason: %v", r), nil)
	}
}

func toRefundStatusDTO(s app.RefundStatus) (billing.RefundStatus, error) {
	switch s {
	case app.RefundStatusPending:
		return billing.RefundStatusPending, nil
	case app.RefundStatusSucceeded:
		return billing.RefundStatusSucceeded, nil
	case app.RefundStatusFailed:
		return billing.RefundStatusFailed, nil
	case app.RefundStatusCanceled:
		return billing.RefundStatusCanceled, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid refund status: %v", s), nil)
	}
}

func toFeatures(m map[string]bool) billing.Features {
	out := billing.Features{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func toLimits(m map[string]int64) billing.Limits {
	out := billing.Limits{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func toEntitlements(e app.Entitlements) (billing.Entitlements, error) {
	plan, err := toPlanTierDTO(e.Plan)
	if err != nil {
		return billing.Entitlements{}, err
	}
	status, err := toSubscriptionStatusDTO(e.Status)
	if err != nil {
		return billing.Entitlements{}, err
	}
	return billing.Entitlements{
		WorkspaceId: e.WorkspaceID, Plan: plan, Status: status,
		AccessUntil: nullableTime(e.AccessUntil), Features: toFeatures(e.Features), Limits: toLimits(e.Limits),
	}, nil
}

func toPrice(p app.Price) (billing.Price, error) {
	plan, err := toPlanTierDTO(p.Plan)
	if err != nil {
		return billing.Price{}, err
	}
	interval, err := toBillingIntervalDTO(p.Interval)
	if err != nil {
		return billing.Price{}, err
	}
	return billing.Price{
		Id: p.ID, Plan: plan, Interval: interval,
		Amount: p.Amount, Currency: p.Currency, LookupKey: p.LookupKey, Active: p.Active,
		CreatedAt: p.CreatedAt, ArchivedAt: nullableTime(p.ArchivedAt),
	}, nil
}

func toPrices(ps []app.Price) ([]billing.Price, error) {
	out := make([]billing.Price, 0, len(ps))
	for _, p := range ps {
		dto, err := toPrice(p)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func toPlan(v app.PlanView) (billing.Plan, error) {
	tier, err := toPlanTierDTO(v.Plan.Tier)
	if err != nil {
		return billing.Plan{}, err
	}
	prices, err := toPrices(v.Prices)
	if err != nil {
		return billing.Plan{}, err
	}
	return billing.Plan{Tier: tier, Features: toFeatures(v.Plan.Features), Limits: toLimits(v.Plan.Limits), Prices: prices}, nil
}

func toPlans(views []app.PlanView) ([]billing.Plan, error) {
	out := make([]billing.Plan, 0, len(views))
	for _, v := range views {
		dto, err := toPlan(v)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func toRefund(r app.Refund) (billing.Refund, error) {
	reason, err := toRefundReasonDTO(r.Reason)
	if err != nil {
		return billing.Refund{}, err
	}
	status, err := toRefundStatusDTO(r.Status)
	if err != nil {
		return billing.Refund{}, err
	}
	return billing.Refund{
		Id: r.ID, Amount: r.Amount, Currency: r.Currency,
		Reason: reason, Status: status, CreatedAt: r.CreatedAt,
	}, nil
}

func toSubscription(s app.Subscription) (billing.Subscription, error) {
	plan, err := toPlanTierDTO(s.Plan)
	if err != nil {
		return billing.Subscription{}, err
	}
	interval, err := toBillingIntervalDTO(s.Interval)
	if err != nil {
		return billing.Subscription{}, err
	}
	status, err := toSubscriptionStatusDTO(s.Status)
	if err != nil {
		return billing.Subscription{}, err
	}
	refunds := make([]billing.Refund, 0, len(s.Refunds))
	for _, r := range s.Refunds {
		dto, err := toRefund(r)
		if err != nil {
			return billing.Subscription{}, err
		}
		refunds = append(refunds, dto)
	}
	return billing.Subscription{
		Id: s.ID, Plan: plan, Interval: interval, Status: status,
		CurrentPeriodStart: nullableTime(s.CurrentPeriodStart), CurrentPeriodEnd: nullableTime(s.CurrentPeriodEnd),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd, CanceledAt: nullableTime(s.CanceledAt),
		AccessUntil: nullableTime(s.AccessUntil), Refunds: refunds,
	}, nil
}

func toWorkspaceBilling(v app.WorkspaceBillingView) (billing.WorkspaceBilling, error) {
	out := billing.WorkspaceBilling{}
	if v.Payer != nil {
		p := v.Payer
		out.Payer = nullable.NewNullableWithValue(billing.Payer{Id: p.ID, Email: p.Email, Name: p.Name, MissingSince: nullableTime(p.MissingSince)})
	}
	if v.Subscription != nil {
		sub, err := toSubscription(*v.Subscription)
		if err != nil {
			return billing.WorkspaceBilling{}, err
		}
		out.Subscription = nullable.NewNullableWithValue(sub)
	}
	return out, nil
}
```

Generated field/type names come from the schemas in Task 1 (`Id`, `WorkspaceId`; the inline `Refund.reason`/`Refund.status` enums generate as `billing.RefundReason`/`billing.RefundStatus`). Enum consts are type-prefixed because of `always-prefix-enum-values` (Task 1 Step 3) — adjust to `pkg/api/billing/model.gen.go` if a name differs.

```go
// internal/billing/controller/http/plan.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) ListPlans(ctx context.Context, _ billing.ListPlansRequestObject) (billing.ListPlansResponseObject, error) {
	views, err := h.App.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out, err := toPlans(views)
	if err != nil {
		return nil, err
	}
	return billing.ListPlans200JSONResponse(out), nil
}
```

```go
// internal/billing/controller/http/workspace.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) GetWorkspaceEntitlements(ctx context.Context, req billing.GetWorkspaceEntitlementsRequestObject) (billing.GetWorkspaceEntitlementsResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	ent, err := h.App.Queries.GetEntitlements.Handle(ctx, &app.GetEntitlements{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	out, err := toEntitlements(ent)
	if err != nil {
		return nil, err
	}
	return billing.GetWorkspaceEntitlements200JSONResponse(out), nil
}

func (h *StrictHandler) GetWorkspaceBilling(ctx context.Context, req billing.GetWorkspaceBillingRequestObject) (billing.GetWorkspaceBillingResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	v, err := h.App.Queries.GetWorkspaceBilling.Handle(ctx, &app.GetWorkspaceBilling{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	out, err := toWorkspaceBilling(v)
	if err != nil {
		return nil, err
	}
	return billing.GetWorkspaceBilling200JSONResponse(out), nil
}
```

Delete the three stubs from `unimplemented.go`.

- [ ] **Step 5: Verify**

```bash
nx run billing:gen:wire && nx build billing && nx lint billing
nx run billing:run
curl -s http://api.notopia.localhost/billing/plans | jq '.[].tier'          # "free" "plus" "pro", prices []
grpcurl -plaintext -d '{"workspace_id":"00000000-0000-4000-8000-000000000110"}' localhost:18084 billing.BillingService/GetWorkspaceEntitlements
```

Expected: entitlements `plan: "PLAN_TIER_FREE"`, `status: "SUBSCRIPTION_STATUS_NONE"` (grpcurl prints enum names), free limits (members 3). (If gRPC reflection is not registered, skip grpcurl; plan 04 exercises it.)

- [ ] **Step 6: Commit**

```bash
git add internal/billing cmd/billing/wire_gen.go
git commit -m "feat(billing): default catalog, entitlement queries, gRPC and read endpoints"
```

---

### Task 6: Payment provider port + Stripe adapter

**Files:**

- Create: `internal/billing/app/svc_paymentprovider.go`
- Create: `internal/billing/infra/payment/stripe/{provider.go,checkout.go,subscription.go,refund.go,catalog.go,map.go}`
- Create: `internal/billing/infra/payment/stripe/map_test.go`, `internal/billing/infra/payment/stripe/testdata/subscription.json`
- Modify: `internal/billing/infra/wire.go`, `go.mod`, `go.sum`

**Interfaces:**

- Produces: `app.PaymentProvider` (below) implemented by `*stripe.Provider` (`stripeadapter` import alias recommended to avoid clashing with `github.com/stripe/stripe-go/v87`); `uint8` enum `app.CheckoutPurpose` (zero `CheckoutPurposeUnspecified`; values keep the full type name as prefix: `CheckoutPurposeSubscribe`, `CheckoutPurposeLifetime`, `CheckoutPurposePayerChange`). Webhook verification and parsing are **not** part of the port: they live in the inbound adapter `controller/http` (Task 8).
- Tracing: no hand-written spans. The Stripe client gets an `otelhttp` transport (same as `internal/note/infra/identity/authentik.go`); stripe-go passes `ctx` to every request, so HTTP spans nest under the use case span from `commonhandler`. Methods drop the named results that existed only for a deferred span.
- Mappers (`infra/payment/stripe/map.go`): `toDomainStatus` (Stripe subscription status → `domain.Status`), `toDomainRefundStatus`, `toStripeTier` (`domain.PlanTier` → `"plus"`/`"pro"`, used for product id, name, metadata, idempotency key), `toStripeRecurringInterval` (`domain.Interval` → `month`/`year`; lifetime = not recurring), `toStripePurpose` / `toAppCheckoutPurpose` (checkout `purpose` metadata).

- [x] **Step 1: Port**

```go
// internal/billing/app/svc_paymentprovider.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type CheckoutPurpose uint8

const (
	CheckoutPurposeUnspecified CheckoutPurpose = iota
	CheckoutPurposeSubscribe
	CheckoutPurposeLifetime
	CheckoutPurposePayerChange
)

func (p CheckoutPurpose) IsValid() bool {
	switch p {
	case CheckoutPurposeSubscribe, CheckoutPurposeLifetime, CheckoutPurposePayerChange:
		return true
	default:
		return false
	}
}

func (p CheckoutPurpose) String() string {
	switch p {
	case CheckoutPurposeUnspecified:
		return "unspecified"
	case CheckoutPurposeSubscribe:
		return "subscribe"
	case CheckoutPurposeLifetime:
		return "lifetime"
	case CheckoutPurposePayerChange:
		return "payer_change"
	default:
		return "unknown"
	}
}

type CheckoutInput struct {
	Purpose         CheckoutPurpose
	WorkspaceID     uuid.UUID
	CustomerID      string
	Payer           domain.Payer
	PriceID         uuid.UUID // zero for payer_change
	ProviderPriceID string    // empty for payer_change
	SuccessURL      string
	CancelURL       string
}

type CheckoutSessionResult struct {
	ID                     string
	Purpose                CheckoutPurpose
	Complete               bool
	WorkspaceID            uuid.UUID
	Payer                  domain.Payer
	PriceID                uuid.UUID
	CustomerID             string
	ProviderSubscriptionID string
	PaymentMethodID        string
	Payment                domain.PaymentInfo // lifetime only
}

type RefundResult struct {
	ProviderRefundID string
	RefundID         *uuid.UUID // from metadata, nil for dashboard refunds
	PaymentRef       string
	Amount           domain.Money
	Status           domain.RefundStatus
	FailureMessage   *string
}

type PriceInput struct {
	PriceID  uuid.UUID
	Plan     domain.PlanTier
	Interval domain.Interval
	Amount   domain.Money
}

type PaymentProvider interface {
	CreateCustomer(ctx context.Context, workspaceID uuid.UUID, payer domain.Payer) (string, error)
	UpdateCustomerContact(ctx context.Context, customerID string, payer domain.Payer) error
	CreateCheckoutSession(ctx context.Context, in CheckoutInput) (url string, err error)
	GetCheckoutSession(ctx context.Context, id string) (CheckoutSessionResult, error)
	CreatePortalSession(ctx context.Context, customerID, configurationID, returnURL string) (url string, err error)
	GetSubscriptionSnapshot(ctx context.Context, providerSubscriptionID string) (domain.ProviderSnapshot, error)
	SetCancelAtPeriodEnd(ctx context.Context, providerSubscriptionID string, cancel bool) error
	CancelSubscriptionNow(ctx context.Context, providerSubscriptionID string) error
	// PeriodPayments lists every paid invoice payment whose invoice lines overlap [periodStart, periodEnd).
	PeriodPayments(ctx context.Context, providerSubscriptionID string, periodStart, periodEnd time.Time) ([]domain.PaymentInfo, error)
	PaymentForIntent(ctx context.Context, paymentIntentID string) (domain.PaymentInfo, error)
	CreateRefund(ctx context.Context, refundID uuid.UUID, paymentRef string, amount domain.Money) (RefundResult, error)
	GetRefund(ctx context.Context, providerRefundID string) (RefundResult, error)
	SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string, providerSubscriptionID *string) error
	DetachPaymentMethods(ctx context.Context, customerID, keepPaymentMethodID string) error // keep "" = detach all
	CreatePrice(ctx context.Context, in PriceInput) (providerPriceID string, err error)
	ArchivePrice(ctx context.Context, providerPriceID string) error
	CreatePortalConfiguration(ctx context.Context, activePrices []*domain.Price, returnURL string) (string, error)
}
```

- [x] **Step 2: Dependency**

```bash
go get github.com/stripe/stripe-go/v87@v87.0.0 && go mod tidy
```

- [x] **Step 3: Write the failing mapping tests**

`testdata/subscription.json` — a minimal subscription in the v87 API shape:

```json
{
  "id": "sub_123",
  "object": "subscription",
  "status": "active",
  "cancel_at_period_end": false,
  "cancel_at": null,
  "canceled_at": null,
  "metadata": { "workspace_id": "00000000-0000-4000-8000-000000000110" },
  "items": {
    "object": "list",
    "data": [
      {
        "id": "si_1",
        "object": "subscription_item",
        "current_period_start": 1759276800,
        "current_period_end": 1761955200,
        "price": { "id": "price_plus_month", "object": "price" }
      }
    ]
  }
}
```

```go
// internal/billing/infra/payment/stripe/map_test.go
package stripe

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripego "github.com/stripe/stripe-go/v87"
)

func TestSnapshotFromSubscription(t *testing.T) {
	raw, err := os.ReadFile("testdata/subscription.json")
	require.NoError(t, err)
	var s stripego.Subscription
	require.NoError(t, json.Unmarshal(raw, &s))
	snap, err := snapshotFromSubscription(&s)
	require.NoError(t, err)
	assert.Equal(t, "price_plus_month", snap.ProviderPriceID)
	assert.Equal(t, domain.StatusActive, snap.Status)
	assert.True(t, snap.CurrentPeriodEnd.Equal(time.Unix(1761955200, 0)))
	assert.Equal(t, "00000000-0000-4000-8000-000000000110", snap.WorkspaceID.String())
}
```

Run: `go test ./internal/billing/infra/payment/stripe/...` → FAIL (package empty).

- [x] **Step 4: Adapter**

```go
// internal/billing/infra/payment/stripe/provider.go
package stripe

import (
	"errors"
	"net/http"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

type Provider struct {
	sc       *stripego.Client
	currency string
}

var _ app.PaymentProvider = (*Provider)(nil)

// Spans come from the HTTP transport (same as internal/note/infra/identity/authentik.go);
// stripe-go forwards ctx to every request, so they nest under the use case span.
func New(stripeCfg *config.Stripe, billingCfg *config.Billing) *Provider {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanOptions(trace.WithAttributes(
				semconv.ServicePeerName("stripe"),
			)),
		),
	}
	return &Provider{
		sc: stripego.NewClient(
			stripeCfg.SecretKey,
			stripego.WithBackends(stripego.NewBackends(httpClient)),
		),
		currency: billingCfg.Currency,
	}
}

var Provide = New

func providerErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return errs.New(errs.CodePaymentProvider, "stripe "+op+" failed", err)
}

func isStripeCode(err error, code stripego.ErrorCode) bool {
	var se *stripego.Error
	return errors.As(err, &se) && se.Code == code
}

func productID(tier string) string { return "notopia_" + tier }
```

```go
// internal/billing/infra/payment/stripe/map.go
package stripe

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
)

func toStripeTier(t domain.PlanTier) (string, error) {
	switch t {
	case domain.PlanTierPlus:
		return "plus", nil
	case domain.PlanTierPro:
		return "pro", nil
	default:
		return "", errs.NewInvalid(fmt.Sprintf("plan %v cannot be sold", t))
	}
}

func toStripeRecurringInterval(i domain.Interval) (string, bool, error) {
	switch i {
	case domain.IntervalMonth:
		return string(stripego.PriceRecurringIntervalMonth), true, nil
	case domain.IntervalYear:
		return string(stripego.PriceRecurringIntervalYear), true, nil
	case domain.IntervalLifetime:
		return "", false, nil
	default:
		return "", false, errs.NewInvalid(fmt.Sprintf("invalid interval %v", i))
	}
}

func toStripePurpose(p app.CheckoutPurpose) (string, error) {
	switch p {
	case app.CheckoutPurposeSubscribe:
		return "subscribe", nil
	case app.CheckoutPurposeLifetime:
		return "lifetime", nil
	case app.CheckoutPurposePayerChange:
		return "payer_change", nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid checkout purpose %v", p), nil)
	}
}

// Sessions not created by billing have no purpose; CompleteCheckout ignores CheckoutPurposeUnspecified.
func toAppCheckoutPurpose(s string) app.CheckoutPurpose {
	switch s {
	case "subscribe":
		return app.CheckoutPurposeSubscribe
	case "lifetime":
		return app.CheckoutPurposeLifetime
	case "payer_change":
		return app.CheckoutPurposePayerChange
	default:
		return app.CheckoutPurposeUnspecified
	}
}

func toDomainStatus(s stripego.SubscriptionStatus) domain.Status {
	switch s {
	case stripego.SubscriptionStatusActive, stripego.SubscriptionStatusTrialing:
		return domain.StatusActive
	case stripego.SubscriptionStatusPastDue:
		return domain.StatusPastDue
	case stripego.SubscriptionStatusIncomplete:
		return domain.StatusIncomplete
	case stripego.SubscriptionStatusIncompleteExpired:
		return domain.StatusIncompleteExpired
	case stripego.SubscriptionStatusCanceled:
		return domain.StatusCanceled
	default: // unpaid, paused
		return domain.StatusUnpaid
	}
}

func snapshotFromSubscription(s *stripego.Subscription) (domain.ProviderSnapshot, error) {
	if s.Items == nil || len(s.Items.Data) == 0 || s.Items.Data[0].Price == nil {
		return domain.ProviderSnapshot{}, errs.NewInternal("stripe subscription "+s.ID+" has no items", nil)
	}
	item := s.Items.Data[0]
	ws, err := uuid.Parse(s.Metadata["workspace_id"])
	if err != nil {
		return domain.ProviderSnapshot{}, errs.NewInternal("stripe subscription "+s.ID+" has no workspace_id metadata", err)
	}
	snap := domain.ProviderSnapshot{
		ProviderSubscriptionID: s.ID,
		ProviderPriceID:        item.Price.ID,
		WorkspaceID:            ws,
		Status:                 toDomainStatus(s.Status),
		CurrentPeriodStart:     time.Unix(item.CurrentPeriodStart, 0),
		CurrentPeriodEnd:       time.Unix(item.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:      s.CancelAtPeriodEnd || s.CancelAt > 0,
	}
	if s.CanceledAt > 0 {
		t := time.Unix(s.CanceledAt, 0)
		snap.CanceledAt = &t
	}
	return snap, nil
}

func toDomainRefundStatus(s stripego.RefundStatus) domain.RefundStatus {
	switch s {
	case stripego.RefundStatusSucceeded:
		return domain.RefundStatusSucceeded
	case stripego.RefundStatusFailed:
		return domain.RefundStatusFailed
	case stripego.RefundStatusCanceled:
		return domain.RefundStatusCanceled
	default:
		return domain.RefundStatusPending
	}
}
```

```go
// internal/billing/infra/payment/stripe/checkout.go
package stripe

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) CreateCustomer(ctx context.Context, workspaceID uuid.UUID, payer domain.Payer) (id string, err error) {
	params := &stripego.CustomerCreateParams{
		Email: stripego.String(payer.Email), Name: stripego.String(payer.Name),
		Metadata: map[string]string{"workspace_id": workspaceID.String(), "payer_id": payer.ID},
	}
	params.SetIdempotencyKey("customer:" + workspaceID.String())
	c, err := p.sc.V1Customers.Create(ctx, params)
	if err != nil {
		return "", providerErr("create customer", err)
	}
	return c.ID, nil
}

func (p *Provider) UpdateCustomerContact(ctx context.Context, customerID string, payer domain.Payer) (err error) {
	_, err = p.sc.V1Customers.Update(ctx, customerID, &stripego.CustomerUpdateParams{
		Email: stripego.String(payer.Email), Name: stripego.String(payer.Name),
		Metadata: map[string]string{"payer_id": payer.ID},
	})
	return providerErr("update customer", err)
}

func (p *Provider) CreateCheckoutSession(ctx context.Context, in app.CheckoutInput) (url string, err error) {
	purpose, err := toStripePurpose(in.Purpose)
	if err != nil {
		return "", err
	}
	meta := map[string]string{
		"workspace_id": in.WorkspaceID.String(), "payer_id": in.Payer.ID, "payer_email": in.Payer.Email,
		"payer_name": in.Payer.Name, "purpose": purpose,
	}
	if in.PriceID != uuid.Nil {
		meta["price_id"] = in.PriceID.String()
	}
	params := &stripego.CheckoutSessionCreateParams{
		Customer: stripego.String(in.CustomerID), ClientReferenceID: stripego.String(in.WorkspaceID.String()),
		SuccessURL: stripego.String(in.SuccessURL), CancelURL: stripego.String(in.CancelURL), Metadata: meta,
	}
	lineItems := []*stripego.CheckoutSessionCreateLineItemParams{{Price: stripego.String(in.ProviderPriceID), Quantity: stripego.Int64(1)}}
	switch in.Purpose {
	case app.CheckoutPurposeSubscribe:
		params.Mode = stripego.String("subscription")
		params.LineItems = lineItems
		params.SubscriptionData = &stripego.CheckoutSessionCreateSubscriptionDataParams{Metadata: meta}
	case app.CheckoutPurposeLifetime:
		params.Mode = stripego.String("payment")
		params.LineItems = lineItems
		params.PaymentIntentData = &stripego.CheckoutSessionCreatePaymentIntentDataParams{Metadata: meta}
	case app.CheckoutPurposePayerChange:
		params.Mode = stripego.String("setup")
		params.Currency = stripego.String(p.currency)
		params.SetupIntentData = &stripego.CheckoutSessionCreateSetupIntentDataParams{Metadata: meta}
	}
	s, err := p.sc.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return "", providerErr("create checkout session", err)
	}
	return s.URL, nil
}

func (p *Provider) GetCheckoutSession(ctx context.Context, id string) (res app.CheckoutSessionResult, err error) {
	params := &stripego.CheckoutSessionRetrieveParams{}
	params.AddExpand("setup_intent")
	params.AddExpand("payment_intent")
	s, err := p.sc.V1CheckoutSessions.Retrieve(ctx, id, params)
	if err != nil {
		return res, providerErr("retrieve checkout session", err)
	}
	ws, _ := uuid.Parse(s.Metadata["workspace_id"])
	priceID, _ := uuid.Parse(s.Metadata["price_id"])
	res = app.CheckoutSessionResult{
		ID: s.ID, Purpose: toAppCheckoutPurpose(s.Metadata["purpose"]), WorkspaceID: ws, PriceID: priceID,
		Payer: domain.Payer{ID: s.Metadata["payer_id"], Email: s.Metadata["payer_email"], Name: s.Metadata["payer_name"]},
		Complete: s.Status == stripego.CheckoutSessionStatusComplete &&
			(s.PaymentStatus == stripego.CheckoutSessionPaymentStatusPaid || s.PaymentStatus == stripego.CheckoutSessionPaymentStatusNoPaymentRequired),
	}
	if s.Customer != nil {
		res.CustomerID = s.Customer.ID
	}
	if s.Subscription != nil {
		res.ProviderSubscriptionID = s.Subscription.ID
	}
	if s.SetupIntent != nil && s.SetupIntent.PaymentMethod != nil {
		res.PaymentMethodID = s.SetupIntent.PaymentMethod.ID
	}
	if pi := s.PaymentIntent; pi != nil {
		res.Payment = domain.PaymentInfo{
			Amount: domain.Money{Amount: pi.AmountReceived, Currency: string(pi.Currency)}, PaymentRef: pi.ID, PaidAt: time.Unix(pi.Created, 0),
		}
	}
	return res, nil
}

func (p *Provider) CreatePortalSession(ctx context.Context, customerID, configurationID, returnURL string) (url string, err error) {
	s, err := p.sc.V1BillingPortalSessions.Create(ctx, &stripego.BillingPortalSessionCreateParams{
		Customer: stripego.String(customerID), Configuration: stripego.String(configurationID), ReturnURL: stripego.String(returnURL),
	})
	if err != nil {
		return "", providerErr("create portal session", err)
	}
	return s.URL, nil
}

func (p *Provider) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string, providerSubscriptionID *string) (err error) {
	if _, err = p.sc.V1Customers.Update(ctx, customerID, &stripego.CustomerUpdateParams{
		InvoiceSettings: &stripego.CustomerUpdateInvoiceSettingsParams{DefaultPaymentMethod: stripego.String(paymentMethodID)},
	}); err != nil {
		return providerErr("set customer default payment method", err)
	}
	if providerSubscriptionID != nil {
		_, err = p.sc.V1Subscriptions.Update(ctx, *providerSubscriptionID, &stripego.SubscriptionUpdateParams{DefaultPaymentMethod: stripego.String(paymentMethodID)})
	}
	return providerErr("set subscription default payment method", err)
}

func (p *Provider) DetachPaymentMethods(ctx context.Context, customerID, keep string) (err error) {
	list := p.sc.V1Customers.ListPaymentMethods(ctx, &stripego.CustomerListPaymentMethodsParams{Customer: stripego.String(customerID)})
	for pm, lerr := range list.All(ctx) {
		if lerr != nil {
			return providerErr("list payment methods", lerr)
		}
		if pm.ID == keep {
			continue
		}
		if _, err = p.sc.V1PaymentMethods.Detach(ctx, pm.ID, nil); err != nil {
			return providerErr("detach payment method", err)
		}
	}
	return nil
}
```

```go
// internal/billing/infra/payment/stripe/subscription.go
package stripe

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) GetSubscriptionSnapshot(ctx context.Context, id string) (snap domain.ProviderSnapshot, err error) {
	s, err := p.sc.V1Subscriptions.Retrieve(ctx, id, nil)
	if err != nil {
		return snap, providerErr("retrieve subscription", err)
	}
	return snapshotFromSubscription(s)
}

func (p *Provider) SetCancelAtPeriodEnd(ctx context.Context, id string, cancel bool) (err error) {
	_, err = p.sc.V1Subscriptions.Update(ctx, id, &stripego.SubscriptionUpdateParams{CancelAtPeriodEnd: stripego.Bool(cancel)})
	return providerErr("update subscription cancel_at_period_end", err)
}

func (p *Provider) CancelSubscriptionNow(ctx context.Context, id string) (err error) {
	_, err = p.sc.V1Subscriptions.Cancel(ctx, id, &stripego.SubscriptionCancelParams{Prorate: stripego.Bool(false), InvoiceNow: stripego.Bool(false)})
	if isStripeCode(err, stripego.ErrorCodeResourceMissing) {
		return nil
	}
	return providerErr("cancel subscription", err)
}

// PeriodPayments: a period can hold the renewal invoice plus proration invoices from upgrades
// (always_invoice). Coverage = the invoice's line periods clipped to the current period.
func (p *Provider) PeriodPayments(ctx context.Context, id string, periodStart, periodEnd time.Time) (out []domain.PaymentInfo, err error) {
	params := &stripego.InvoiceListParams{Subscription: stripego.String(id), Status: stripego.String("paid")}
	params.CreatedRange = &stripego.RangeQueryParams{GreaterThanOrEqual: periodStart.Add(-24 * time.Hour).Unix()}
	for inv, lerr := range p.sc.V1Invoices.List(ctx, params).All(ctx) {
		if lerr != nil {
			return nil, providerErr("list invoices", lerr)
		}
		var covStart, covEnd time.Time
		for _, l := range inv.Lines.Data {
			if l.Period == nil {
				continue
			}
			ls, le := time.Unix(l.Period.Start, 0), time.Unix(l.Period.End, 0)
			if covStart.IsZero() || ls.Before(covStart) {
				covStart = ls
			}
			if le.After(covEnd) {
				covEnd = le
			}
		}
		if !covEnd.After(periodStart) || !covStart.Before(periodEnd) {
			continue
		}
		payments := p.sc.V1InvoicePayments.List(ctx, &stripego.InvoicePaymentListParams{Invoice: stripego.String(inv.ID), Status: stripego.String("paid")})
		for ip, perr := range payments.All(ctx) {
			if perr != nil {
				return nil, providerErr("list invoice payments", perr)
			}
			if ip.Payment == nil || ip.Payment.PaymentIntent == nil {
				continue
			}
			paidAt := time.Unix(inv.Created, 0)
			if ip.StatusTransitions != nil && ip.StatusTransitions.PaidAt > 0 {
				paidAt = time.Unix(ip.StatusTransitions.PaidAt, 0)
			}
			out = append(out, domain.PaymentInfo{
				Amount: domain.Money{Amount: ip.AmountPaid, Currency: string(ip.Currency)}, PaymentRef: ip.Payment.PaymentIntent.ID,
				PaidAt: paidAt, CoverageStart: covStart, CoverageEnd: covEnd,
			})
		}
	}
	if len(out) == 0 {
		return nil, errs.NewNotFound("no paid payment in the current period")
	}
	return out, nil
}

func (p *Provider) PaymentForIntent(ctx context.Context, paymentIntentID string) (info domain.PaymentInfo, err error) {
	pi, err := p.sc.V1PaymentIntents.Retrieve(ctx, paymentIntentID, nil)
	if err != nil {
		return info, providerErr("retrieve payment intent", err)
	}
	return domain.PaymentInfo{Amount: domain.Money{Amount: pi.AmountReceived, Currency: string(pi.Currency)}, PaymentRef: pi.ID, PaidAt: time.Unix(pi.Created, 0)}, nil
}
```

```go
// internal/billing/infra/payment/stripe/refund.go
package stripe

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	stripego "github.com/stripe/stripe-go/v87"
)

func toRefundResult(r *stripego.Refund) app.RefundResult {
	res := app.RefundResult{
		ProviderRefundID: r.ID, Amount: domain.Money{Amount: r.Amount, Currency: string(r.Currency)}, Status: toDomainRefundStatus(r.Status),
	}
	if r.PaymentIntent != nil {
		res.PaymentRef = r.PaymentIntent.ID
	}
	if id, err := uuid.Parse(r.Metadata["refund_id"]); err == nil {
		res.RefundID = &id
	}
	if r.FailureReason != "" {
		msg := string(r.FailureReason)
		res.FailureMessage = &msg
	}
	return res
}

func (p *Provider) CreateRefund(ctx context.Context, refundID uuid.UUID, paymentRef string, amount domain.Money) (res app.RefundResult, err error) {
	params := &stripego.RefundCreateParams{
		PaymentIntent: stripego.String(paymentRef), Amount: stripego.Int64(amount.Amount),
		Metadata: map[string]string{"refund_id": refundID.String()},
	}
	params.SetIdempotencyKey("refund:" + refundID.String())
	r, err := p.sc.V1Refunds.Create(ctx, params)
	if err != nil {
		return res, providerErr("create refund", err)
	}
	return toRefundResult(r), nil
}

func (p *Provider) GetRefund(ctx context.Context, id string) (res app.RefundResult, err error) {
	r, err := p.sc.V1Refunds.Retrieve(ctx, id, nil)
	if err != nil {
		return res, providerErr("retrieve refund", err)
	}
	return toRefundResult(r), nil
}
```

```go
// internal/billing/infra/payment/stripe/catalog.go
package stripe

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) ensureProduct(ctx context.Context, tier string) error {
	params := &stripego.ProductCreateParams{
		ID: stripego.String(productID(tier)), Name: stripego.String("Notopia " + tier),
		Metadata: map[string]string{"tier": tier},
	}
	params.SetIdempotencyKey("product:" + tier)
	_, err := p.sc.V1Products.Create(ctx, params)
	if err != nil && !isStripeCode(err, stripego.ErrorCodeResourceAlreadyExists) {
		return providerErr("create product", err)
	}
	return nil
}

func (p *Provider) CreatePrice(ctx context.Context, in app.PriceInput) (id string, err error) {
	tier, err := toStripeTier(in.Plan)
	if err != nil {
		return "", err
	}
	interval, recurring, err := toStripeRecurringInterval(in.Interval)
	if err != nil {
		return "", err
	}
	if err = p.ensureProduct(ctx, tier); err != nil {
		return "", err
	}
	params := &stripego.PriceCreateParams{
		Product: stripego.String(productID(tier)), Currency: stripego.String(in.Amount.Currency),
		UnitAmount: stripego.Int64(in.Amount.Amount), LookupKey: stripego.String(domain.LookupKey(in.Plan, in.Interval)),
		TransferLookupKey: stripego.Bool(true), Metadata: map[string]string{"price_id": in.PriceID.String()},
	}
	if recurring {
		params.Recurring = &stripego.PriceCreateRecurringParams{Interval: stripego.String(interval)}
	}
	params.SetIdempotencyKey("price:" + in.PriceID.String())
	pr, err := p.sc.V1Prices.Create(ctx, params)
	if err != nil {
		return "", providerErr("create price", err)
	}
	return pr.ID, nil
}

func (p *Provider) ArchivePrice(ctx context.Context, providerPriceID string) (err error) {
	_, err = p.sc.V1Prices.Update(ctx, providerPriceID, &stripego.PriceUpdateParams{Active: stripego.Bool(false)})
	return providerErr("archive price", err)
}

func (p *Provider) CreatePortalConfiguration(ctx context.Context, activePrices []*domain.Price, returnURL string) (id string, err error) {
	byProduct := map[string][]*string{}
	for _, pr := range activePrices {
		if pr.Interval() == domain.IntervalLifetime {
			continue
		}
		tier, err := toStripeTier(pr.Plan())
		if err != nil {
			return "", err
		}
		pid := productID(tier)
		byProduct[pid] = append(byProduct[pid], stripego.String(pr.ProviderPriceID()))
	}
	products := make([]*stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateProductParams, 0, len(byProduct))
	for pid, prices := range byProduct {
		products = append(products, &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateProductParams{Product: stripego.String(pid), Prices: prices})
	}
	features := &stripego.BillingPortalConfigurationCreateFeaturesParams{
		InvoiceHistory:      &stripego.BillingPortalConfigurationCreateFeaturesInvoiceHistoryParams{Enabled: stripego.Bool(true)},
		PaymentMethodUpdate: &stripego.BillingPortalConfigurationCreateFeaturesPaymentMethodUpdateParams{Enabled: stripego.Bool(true)},
		SubscriptionCancel: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionCancelParams{
			Enabled: stripego.Bool(true), Mode: stripego.String("at_period_end"),
		},
		SubscriptionUpdate: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateParams{
			Enabled: stripego.Bool(len(products) > 0), DefaultAllowedUpdates: []*string{stripego.String("price")},
			Products: products, ProrationBehavior: stripego.String("always_invoice"),
			ScheduleAtPeriodEnd: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateScheduleAtPeriodEndParams{
				Conditions: []*stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateScheduleAtPeriodEndConditionParams{
					{Type: stripego.String("decreasing_item_amount")},
					{Type: stripego.String("shortening_interval")},
				},
			},
		},
	}
	cfg, err := p.sc.V1BillingPortalConfigurations.Create(ctx, &stripego.BillingPortalConfigurationCreateParams{
		Features: features, DefaultReturnURL: stripego.String(returnURL),
	})
	if err != nil {
		return "", providerErr("create portal configuration", err)
	}
	return cfg.ID, nil
}
```

`infra/wire.go`: add `stripeadapter.Provide` and `wire.Bind(new(app.PaymentProvider), new(*stripeadapter.Provider))` (import `stripeadapter "github.com/notopia-uit/notopia/internal/billing/infra/payment/stripe"`).

- [x] **Step 5: Run tests, build, commit**

```bash
go test ./internal/billing/infra/payment/stripe/...
nx run billing:gen:wire && nx build billing && nx lint billing
git add internal/billing go.mod go.sum cmd/billing/wire_gen.go
git commit -m "feat(billing): stripe payment provider adapter"
```

---

### Task 7: Catalog management — admin APIs, seeder, portal configuration

**Files:**

- Create: `internal/billing/app/{settings.go,catalog.go,cmd_createprice.go,cmd_archiveprice.go,cmd_updateplan.go,cmd_updatepolicy.go,cmd_ensureportalconfiguration.go,q_listprices.go,q_getprice.go,q_getplan.go,q_getpolicy.go}`
- Create: `internal/billing/settings.go` (maps config to `app.Settings`; composition root, so `config` never imports `app`)
- Create: `internal/billing/controller/http/{admin.go,mapin.go}`; modify `unimplemented.go`
- Create: `internal/billingseed/{config/{config.go,wire.go},seed.go,wire.go}`, `internal/billingfixtures/{config/{config.go,wire.go},fixtures.go,wire.go}`
- Create: `cmd/billingseed/{main.go,wire.go,project.json,.env,.gitignore}`, `cmd/billingfixtures/{main.go,wire.go,project.json,.env,.gitignore}` (+ generated `wire_gen.go`)
- Modify: `justfile`

**Interfaces:**

- Produces: `app.Settings{WebBaseURL, ServiceAccountUserID string}`; `Cmds.CreatePrice`, `Cmds.ArchivePrice`, `Cmds.UpdatePlan`, `Cmds.UpdatePolicy`, `Cmds.EnsurePortalConfiguration` (`EnsurePortalConfiguration{UserID string}`: creates the Stripe portal configuration only when `billing_policy` has none); `Queries.ListPrices`, `Queries.GetPrice`, `Queries.GetPlan`, `Queries.GetPolicy`. Commands return only `error` and never carry results: the controller generates ids up front (`CreatePrice.ID`) and, after a command, answers with a query (`GetPrice`, `GetPlan`, `GetPolicy`, `GetWorkspaceBilling`). Command inputs use app types (`app.PlanTier`, `app.BillingInterval`, string-keyed feature/limit maps); handlers map them to domain values in `app/mapper.go`.
- Mappers (`controller/http/mapin.go`): `toPlanTier` (`billing.PlanTier` path param → `app.PlanTier`), `toCreatePricePlan` (`billing.CreatePriceRequestPlan`), `toInterval` (`billing.BillingInterval` → `app.BillingInterval`), `fromFeatures`/`fromLimits` (request maps, passed through as string-keyed maps). Unknown enum values return `errs.NewInvalid` → 400; unknown feature/limit keys are rejected by the handler (`parseFeatures`/`parseLimits`).

- [x] **Step 1: Settings**

```go
// internal/billing/app/settings.go
package app

type Settings struct {
	WebBaseURL           string
	ServiceAccountUserID string
}
```

`app` declares what it needs and never imports `config`. The composition root (`internal/billing`) does the mapping, so `config` stays a leaf package like `internal/note/config`. Viper and the env vars do not change.

```go
// internal/billing/settings.go
package billing

import (
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func NewAppSettings(billingCfg *config.Billing, identity *commonconfig.Identity) *app.Settings {
	return &app.Settings{
		WebBaseURL:           billingCfg.WebBaseURL,
		ServiceAccountUserID: identity.ServiceAccountUserID,
	}
}

var ProvideAppSettings = NewAppSettings
```

Add `ProvideAppSettings` to `ProviderSet` in `internal/billing/wire.go`. `cmd/billingseed` uses it too (Step 4).

- [x] **Step 2: Catalog helpers and commands**

```go
// internal/billing/app/catalog.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type catalog struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func newCatalog(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *catalog {
	return &catalog{uow: uow, provider: provider, settings: settings}
}

// createPrice creates the Stripe price first (idempotent by our price id), then swaps the active row.
func (c *catalog) createPrice(ctx context.Context, id uuid.UUID, plan domain.PlanTier, interval domain.Interval, amount domain.Money, userID string) (*domain.Price, error) {
	providerPriceID, err := c.provider.CreatePrice(ctx, PriceInput{PriceID: id, Plan: plan, Interval: interval, Amount: amount})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var created, previous *domain.Price
	err = c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if previous, err = r.Price().GetActive(ctx, plan, interval, amount.Currency, true); err != nil {
			return err
		}
		if previous != nil {
			_ = previous.Archive(now, userID)
			if err := r.Price().Save(ctx, previous); err != nil {
				return err
			}
		}
		if created, err = domain.NewPrice(id, plan, interval, amount, providerPriceID, now, userID); err != nil {
			return err
		}
		return r.Price().Save(ctx, created)
	})
	if err != nil {
		return nil, err
	}
	if previous != nil {
		if err := c.provider.ArchivePrice(ctx, previous.ProviderPriceID()); err != nil {
			return nil, err
		}
	}
	return created, c.refreshPortal(ctx, userID)
}

func (c *catalog) refreshPortal(ctx context.Context, userID string) error {
	var prices []*domain.Price
	if err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		prices, err = r.Price().ListActive(ctx)
		return err
	}); err != nil {
		return err
	}
	id, err := c.provider.CreatePortalConfiguration(ctx, prices, c.settings.WebBaseURL)
	if err != nil {
		return err
	}
	return c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		policy, err := r.BillingPolicy().Get(ctx, true)
		if err != nil {
			return err
		}
		policy.SetPortalConfiguration(id, time.Now(), userID)
		return r.BillingPolicy().Save(ctx, policy)
	})
}
```

```go
// internal/billing/app/cmd_createprice.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CreatePrice struct {
	ID       uuid.UUID
	Plan     PlanTier
	Interval BillingInterval
	Amount   int64
	Currency string
	UserID   string
}

type CreatePriceHandler struct{ catalog *catalog }

func NewCreatePriceHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *CreatePriceHandler {
	return &CreatePriceHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideCreatePriceHandler = NewCreatePriceHandler

type CreatePriceCmd commonhandler.Cmd[CreatePrice]

var _ CreatePriceCmd = (*CreatePriceHandler)(nil)

func (h *CreatePriceHandler) Handle(ctx context.Context, cmd *CreatePrice) error {
	_, err := h.catalog.createPrice(ctx, cmd.ID, toDomainPlanTier(cmd.Plan), toDomainInterval(cmd.Interval),
		domain.Money{Amount: cmd.Amount, Currency: cmd.Currency}, cmd.UserID)
	return err
}
```

```go
// internal/billing/app/cmd_archiveprice.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ArchivePrice struct {
	PriceID uuid.UUID
	UserID  string
}

type ArchivePriceHandler struct{ catalog *catalog }

func NewArchivePriceHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *ArchivePriceHandler {
	return &ArchivePriceHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideArchivePriceHandler = NewArchivePriceHandler

type ArchivePriceCmd commonhandler.Cmd[ArchivePrice]

var _ ArchivePriceCmd = (*ArchivePriceHandler)(nil)

func (h *ArchivePriceHandler) Handle(ctx context.Context, cmd *ArchivePrice) error {
	var price *domain.Price
	if err := h.catalog.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if price, err = r.Price().GetByID(ctx, cmd.PriceID, true); err != nil {
			return err
		}
		_ = price.Archive(time.Now(), cmd.UserID)
		return r.Price().Save(ctx, price)
	}); err != nil {
		return err
	}
	if err := h.catalog.provider.ArchivePrice(ctx, price.ProviderPriceID()); err != nil {
		return err
	}
	return h.catalog.refreshPortal(ctx, cmd.UserID)
}
```

```go
// internal/billing/app/cmd_updateplan.go
package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdatePlan struct {
	Tier     PlanTier
	Features map[string]bool
	Limits   map[string]int64
	UserID   string
}

type UpdatePlanHandler struct{ uow domain.UnitOfWork }

func NewUpdatePlanHandler(uow domain.UnitOfWork) *UpdatePlanHandler {
	return &UpdatePlanHandler{uow: uow}
}

var ProvideUpdatePlanHandler = NewUpdatePlanHandler

type UpdatePlanCmd commonhandler.Cmd[UpdatePlan]

var _ UpdatePlanCmd = (*UpdatePlanHandler)(nil)

func (h *UpdatePlanHandler) Handle(ctx context.Context, cmd *UpdatePlan) error {
	features, err := parseFeatures(cmd.Features)
	if err != nil {
		return err
	}
	limits, err := parseLimits(cmd.Limits)
	if err != nil {
		return err
	}
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		plan, err := r.Plan().Get(ctx, toDomainPlanTier(cmd.Tier), true)
		if err != nil {
			return err
		}
		if err := plan.Update(features, limits, time.Now(), cmd.UserID); err != nil {
			return err
		}
		return r.Plan().Save(ctx, plan)
	})
}
```

```go
// internal/billing/app/cmd_updatepolicy.go
package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdatePolicy struct {
	PastDueGraceDays         *int
	LifetimeRefundWindowDays *int
	UserID                   string
}

type UpdatePolicyHandler struct{ uow domain.UnitOfWork }

func NewUpdatePolicyHandler(uow domain.UnitOfWork) *UpdatePolicyHandler { return &UpdatePolicyHandler{uow: uow} }

var ProvideUpdatePolicyHandler = NewUpdatePolicyHandler

type UpdatePolicyCmd commonhandler.Cmd[UpdatePolicy]

var _ UpdatePolicyCmd = (*UpdatePolicyHandler)(nil)

func (h *UpdatePolicyHandler) Handle(ctx context.Context, cmd *UpdatePolicy) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		policy, err := r.BillingPolicy().Get(ctx, true)
		if err != nil {
			return err
		}
		if err := policy.Update(cmd.PastDueGraceDays, cmd.LifetimeRefundWindowDays, time.Now(), cmd.UserID); err != nil {
			return err
		}
		return r.BillingPolicy().Save(ctx, policy)
	})
}
```

```go
// internal/billing/app/cmd_ensureportalconfiguration.go
package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type EnsurePortalConfiguration struct {
	UserID string
}

type EnsurePortalConfigurationHandler struct{ catalog *catalog }

func NewEnsurePortalConfigurationHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *EnsurePortalConfigurationHandler {
	return &EnsurePortalConfigurationHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideEnsurePortalConfigurationHandler = NewEnsurePortalConfigurationHandler

type EnsurePortalConfigurationCmd commonhandler.Cmd[EnsurePortalConfiguration]

var _ EnsurePortalConfigurationCmd = (*EnsurePortalConfigurationHandler)(nil)

// Handle creates the portal configuration only when billing_policy has none;
// CreatePrice/ArchivePrice already refresh it when the price set changes.
func (h *EnsurePortalConfigurationHandler) Handle(ctx context.Context, cmd *EnsurePortalConfiguration) error {
	var policy *domain.BillingPolicy
	if err := h.catalog.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	}); err != nil {
		return err
	}
	if policy.PortalConfigurationID() != nil {
		return nil
	}
	return h.catalog.refreshPortal(ctx, cmd.UserID)
}
```

`q_getprice.go` (`GetPrice{ID uuid.UUID}` → `Price`; scans `read.Prices` and returns `errs.NewNotFound` when absent, the admin catalog is tiny) and `q_getplan.go` (`GetPlan{Tier PlanTier}` → `PlanView`, same assembly as `q_listplans.go` filtered to one tier via `toDomainPlanTier`). `q_listprices.go` (`ListPrices{}` → `[]Price` via `read.Prices` + `newPrices`) and `q_getpolicy.go` (`GetPolicy{}` → `BillingPolicy` via `read.Policy` + `newBillingPolicy`): same handler shape as `q_listplans.go` (depend on `BillingReadModel`, no unit of work).

Register all in `app.go` / `wire.go`.

- [x] **Step 3: Admin HTTP**

```go
// internal/billing/controller/http/admin.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) AdminListPlans(ctx context.Context, _ billing.AdminListPlansRequestObject) (billing.AdminListPlansResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	views, err := h.App.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out, err := toPlans(views)
	if err != nil {
		return nil, err
	}
	return billing.AdminListPlans200JSONResponse(out), nil
}

func (h *StrictHandler) AdminUpdatePlan(ctx context.Context, req billing.AdminUpdatePlanRequestObject) (billing.AdminUpdatePlanResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	tier, err := toPlanTier(req.Tier)
	if err != nil {
		return nil, err
	}
	cmd := &app.UpdatePlan{Tier: tier, UserID: u.ID, Features: fromFeatures(req.Body.Features), Limits: fromLimits(req.Body.Limits)}
	if err := h.App.Cmds.UpdatePlan.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	v, err := h.App.Queries.GetPlan.Handle(ctx, &app.GetPlan{Tier: cmd.Tier})
	if err != nil {
		return nil, err
	}
	out, err := toPlan(v)
	if err != nil {
		return nil, err
	}
	return billing.AdminUpdatePlan200JSONResponse(out), nil
}

func toPolicy(p app.BillingPolicy) billing.BillingPolicy {
	return billing.BillingPolicy{PastDueGraceDays: p.PastDueGraceDays, LifetimeRefundWindowDays: p.LifetimeRefundWindowDays}
}

func (h *StrictHandler) AdminGetPolicy(ctx context.Context, _ billing.AdminGetPolicyRequestObject) (billing.AdminGetPolicyResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	p, err := h.App.Queries.GetPolicy.Handle(ctx, &app.GetPolicy{})
	if err != nil {
		return nil, err
	}
	return billing.AdminGetPolicy200JSONResponse(toPolicy(p)), nil
}

func (h *StrictHandler) AdminUpdatePolicy(ctx context.Context, req billing.AdminUpdatePolicyRequestObject) (billing.AdminUpdatePolicyResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	cmd := &app.UpdatePolicy{PastDueGraceDays: req.Body.PastDueGraceDays, LifetimeRefundWindowDays: req.Body.LifetimeRefundWindowDays, UserID: u.ID}
	if err := h.App.Cmds.UpdatePolicy.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	p, err := h.App.Queries.GetPolicy.Handle(ctx, &app.GetPolicy{})
	if err != nil {
		return nil, err
	}
	return billing.AdminUpdatePolicy200JSONResponse(toPolicy(p)), nil
}

func (h *StrictHandler) AdminListPrices(ctx context.Context, _ billing.AdminListPricesRequestObject) (billing.AdminListPricesResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	prices, err := h.App.Queries.ListPrices.Handle(ctx, &app.ListPrices{})
	if err != nil {
		return nil, err
	}
	out, err := toPrices(prices)
	if err != nil {
		return nil, err
	}
	return billing.AdminListPrices200JSONResponse(out), nil
}

func (h *StrictHandler) AdminCreatePrice(ctx context.Context, req billing.AdminCreatePriceRequestObject) (billing.AdminCreatePriceResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := toCreatePricePlan(req.Body.Plan)
	if err != nil {
		return nil, err
	}
	interval, err := toInterval(req.Body.Interval)
	if err != nil {
		return nil, err
	}
	cmd := &app.CreatePrice{
		ID: uuid.Must(uuid.NewV7()), Plan: plan, Interval: interval,
		Amount: req.Body.Amount, Currency: req.Body.Currency, UserID: u.ID,
	}
	if err := h.App.Cmds.CreatePrice.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	price, err := h.App.Queries.GetPrice.Handle(ctx, &app.GetPrice{ID: cmd.ID})
	if err != nil {
		return nil, err
	}
	out, err := toPrice(price)
	if err != nil {
		return nil, err
	}
	return billing.AdminCreatePrice201JSONResponse(out), nil
}

func (h *StrictHandler) AdminArchivePrice(ctx context.Context, req billing.AdminArchivePriceRequestObject) (billing.AdminArchivePriceResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	cmd := &app.ArchivePrice{PriceID: req.PriceId, UserID: u.ID}
	if err := h.App.Cmds.ArchivePrice.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	price, err := h.App.Queries.GetPrice.Handle(ctx, &app.GetPrice{ID: req.PriceId})
	if err != nil {
		return nil, err
	}
	out, err := toPrice(price)
	if err != nil {
		return nil, err
	}
	return billing.AdminArchivePrice200JSONResponse(out), nil
}
```

```go
// internal/billing/controller/http/mapin.go
package http

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func toPlanTier(t billing.PlanTier) (app.PlanTier, error) {
	switch t {
	case billing.PlanTierFree:
		return app.PlanTierFree, nil
	case billing.PlanTierPlus:
		return app.PlanTierPlus, nil
	case billing.PlanTierPro:
		return app.PlanTierPro, nil
	default:
		return app.PlanTierUnspecified, errs.NewInvalid(fmt.Sprintf("invalid plan tier: %q", t))
	}
}

func toCreatePricePlan(p billing.CreatePriceRequestPlan) (app.PlanTier, error) {
	switch p {
	case billing.CreatePriceRequestPlanPlus:
		return app.PlanTierPlus, nil
	case billing.CreatePriceRequestPlanPro:
		return app.PlanTierPro, nil
	default:
		return app.PlanTierUnspecified, errs.NewInvalid(fmt.Sprintf("invalid price plan: %q", p))
	}
}

func toInterval(i billing.BillingInterval) (app.BillingInterval, error) {
	switch i {
	case billing.BillingIntervalMonth:
		return app.BillingIntervalMonth, nil
	case billing.BillingIntervalYear:
		return app.BillingIntervalYear, nil
	case billing.BillingIntervalLifetime:
		return app.BillingIntervalLifetime, nil
	default:
		return app.BillingIntervalUnspecified, errs.NewInvalid(fmt.Sprintf("invalid billing interval: %q", i))
	}
}

func fromFeatures(in *billing.Features) map[string]bool {
	if in == nil {
		return map[string]bool{}
	}
	return map[string]bool(*in)
}

func fromLimits(in *billing.Limits) map[string]int64 {
	if in == nil {
		return map[string]int64{}
	}
	return map[string]int64(*in)
}
```

Delete the seven admin stubs from `unimplemented.go`.

- [x] **Step 4: Seed binary (mandatory catalog)**

Naming follows the repo convention: **seed** = mandatory data for every environment (idempotent), **fixtures** = dev-only fake data. The seed creates the default prices in Stripe and the DB and makes sure the portal configuration exists.

Rules:

- Own small config (`NOTOPIA_BILLINGSEED_*`, own `.env`): only database, Stripe secret key, catalog values, service account. No Kafka, Redis, authorization or server settings.
- Calls only stable app use cases (`EnsureDefaults`, `ListPrices`, `CreatePrice`, `EnsurePortalConfiguration`) as undecorated handlers; no `app.Server`, no billing helpers.
- Reuses billing infra (`persistence`, `stripeadapter`) through the composition root, which maps the small config onto the types those constructors take.
- Runs migrations first (goose is idempotent), so it works on a fresh database.

Layout:

```
cmd/billingseed/            main.go, wire.go, wire_gen.go, project.json, .env, .gitignore
internal/billingseed/
  config/                   config.go (own Config + viper), wire.go
  seed.go                   default prices + Seeder.Run
  wire.go                   ProviderSet + mappers to billing config / app.Settings
```

```go
// internal/billingseed/config/config.go
package config

import (
	// … same imports as internal/billing/config/config.go
)

type Stripe struct {
	SecretKey string `default:"-" json:"-" mapstructure:"secret_key" validate:"required" yaml:"secret_key"`
}

type Catalog struct {
	Currency   string `default:"vnd" json:"currency"   mapstructure:"currency"     validate:"required"     yaml:"currency"`
	WebBaseURL string `default:"-"   json:"webBaseUrl" mapstructure:"web_base_url" validate:"required,url" yaml:"web_base_url"`
}

type Config struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Catalog  Catalog               `json:"catalog"  mapstructure:"catalog"  validate:"required"  yaml:"catalog"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
}

func New(validate *validator.Validate, viper *viper.Viper) (*Config, error) {
	viper.SetEnvPrefix("notopia_billingseed")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("billingseed.notopia.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	// … ReadInConfig, defaults.Set, Unmarshal, log, validate: same steps as internal/billing/config/config.go
}

var Provide = New

func NewViper() *viper.Viper {
	return viper.NewWithOptions(viper.ExperimentalBindStruct())
}

var ProvideViper = NewViper

func NewValidate() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

var ProvideValidate = NewValidate
```

```go
// internal/billingseed/config/wire.go
package config

var ProviderSet = wire.NewSet(
	ProvideValidate,
	ProvideViper,
	Provide,
	wire.FieldsOf(
		new(*Config),
		"Catalog",
		"Database",
		"General",
		"Identity",
		"Log",
		"Stripe",
	),
)
```

```go
// internal/billingseed/wire.go
package billingseed

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	billingconfig "github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	stripeadapter "github.com/notopia-uit/notopia/internal/billing/infra/payment/stripe"
	"github.com/notopia-uit/notopia/internal/billingseed/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

// Billing infra constructors take billing's config types; map the small seed config onto them here,
// so neither config package imports the other or app.

func NewBillingStripeConfig(stripe *config.Stripe) *billingconfig.Stripe {
	return &billingconfig.Stripe{
		SecretKey:     stripe.SecretKey,
		WebhookSecret: "",
	}
}

var ProvideBillingStripeConfig = NewBillingStripeConfig

func NewBillingConfig(catalog *config.Catalog) *billingconfig.Billing {
	return &billingconfig.Billing{
		Currency:   catalog.Currency,
		WebBaseURL: catalog.WebBaseURL,
	}
}

var ProvideBillingConfig = NewBillingConfig

func NewAppSettings(catalog *config.Catalog, identity *commonconfig.Identity) *app.Settings {
	return &app.Settings{
		WebBaseURL:           catalog.WebBaseURL,
		ServiceAccountUserID: identity.ServiceAccountUserID,
	}
}

var ProvideAppSettings = NewAppSettings

var ProviderSet = wire.NewSet(
	ProvideSeeder,
	ProvideBillingStripeConfig,
	ProvideBillingConfig,
	ProvideAppSettings,
	config.ProviderSet,
	persistence.ProviderSet,
	stripeadapter.Provide,
	wire.Bind(new(app.PaymentProvider), new(*stripeadapter.Provider)),
	app.ProvideEnsureDefaultsHandler,
	app.ProvideListPricesHandler,
	app.ProvideCreatePriceHandler,
	app.ProvideEnsurePortalConfigurationHandler,
	logging.ProviderSet,
	otel.ProviderSet,
)
```

```go
// internal/billingseed/seed.go
package billingseed

type defaultPrice struct {
	plan     app.PlanTier
	interval app.BillingInterval
	amount   int64
}

var defaultPrices = []defaultPrice{
	{plan: app.PlanTierPlus, interval: app.BillingIntervalMonth, amount: 49000},
	{plan: app.PlanTierPlus, interval: app.BillingIntervalYear, amount: 470400},
	{plan: app.PlanTierPro, interval: app.BillingIntervalMonth, amount: 99000},
	{plan: app.PlanTierPro, interval: app.BillingIntervalYear, amount: 950400},
	{plan: app.PlanTierPro, interval: app.BillingIntervalLifetime, amount: 2490000},
}

type Seeder struct {
	pg             *persistence.Pg
	ensureDefaults *app.EnsureDefaultsHandler
	listPrices     *app.ListPricesHandler
	createPrice    *app.CreatePriceHandler
	ensurePortal   *app.EnsurePortalConfigurationHandler
	catalog        *config.Catalog
	identity       *commonconfig.Identity
	logger         *slog.Logger
}

func NewSeeder(/* … every field above */) *Seeder { … }

var ProvideSeeder = NewSeeder

// Run is idempotent: a (plan, interval, currency) that already has an active price is skipped.
func (s *Seeder) Run(ctx context.Context) error {
	if err := s.pg.RunMigrations(ctx); err != nil {
		return err
	}
	if err := s.ensureDefaults.Handle(ctx, &app.EnsureDefaults{}); err != nil {
		return err
	}
	prices, err := s.listPrices.Handle(ctx, &app.ListPrices{})
	if err != nil {
		return err
	}
	actor := s.identity.ServiceAccountUserID
	for _, d := range defaultPrices {
		if hasActive(prices, d, s.catalog.Currency) {
			continue
		}
		if err := s.createPrice.Handle(ctx, &app.CreatePrice{
			ID:       uuid.Must(uuid.NewV7()),
			Plan:     d.plan,
			Interval: d.interval,
			Amount:   d.amount,
			Currency: s.catalog.Currency,
			UserID:   actor,
		}); err != nil {
			return err
		}
		s.logger.InfoContext(
			ctx,
			"price created",
			slog.String("plan", d.plan.String()),
			slog.String("interval", d.interval.String()),
		)
	}
	return s.ensurePortal.Handle(ctx, &app.EnsurePortalConfiguration{UserID: actor})
}

func hasActive(prices []app.Price, d defaultPrice, currency string) bool {
	// … any price with Active && same Plan, Interval, Currency
}
```

```go
// cmd/billingseed/main.go
func main() {
	if err := logging.FirstStart("NOTOPIA_BILLINGSEED_LOG_LEVEL"); err != nil { … os.Exit(1) }
	ctx := context.Background()
	seeder, cleanup, err := InitializeSeeder(ctx)
	if err != nil { … os.Exit(1) }
	defer cleanup()
	if err := seeder.Run(ctx); err != nil { … os.Exit(1) }
	slog.Info("billing seed done")
}
```

```go
// cmd/billingseed/wire.go
//go:build wireinject

func InitializeSeeder(ctx context.Context) (*billingseed.Seeder, func(), error) {
	panic(wire.Build(
		billingseed.ProviderSet,
		wire.Value(metadata.ServiceName("notopia-billingseed")),
		wire.Value(metadata.ServiceVersion("v0.0.0")),
	))
}
```

`cmd/billingseed/.env` (own file, no symlink; the Stripe key goes in `.env.local`):

```dotenv
NOTOPIA_BILLINGSEED_GENERAL_APP_ENV="development"
NOTOPIA_BILLINGSEED_LOG_LEVEL=debug
NOTOPIA_BILLINGSEED_LOG_PRETTY=true
NOTOPIA_BILLINGSEED_DATABASE_URL=postgres://notopiauit:notopiauit@localhost:5435/billing?sslmode=disable
NOTOPIA_BILLINGSEED_CATALOG_WEB_BASE_URL=http://web.notopia.localhost
NOTOPIA_BILLINGSEED_IDENTITY_SERVICE_ACCOUNT_USER_ID=1003
OTEL_SERVICE_NAME="billingseed"
```

`cmd/billingseed/project.json`: `gen:wire`, `run` (`go run ./cmd/billingseed/`, `dependsOn: ["gen:wire"]`), `build`, `lint` (`./cmd/billingseed/... ./internal/billingseed/...`), shaped like `cmd/billing/project.json`. `.gitignore`: `billingseed`.

- [x] **Step 5: Fixtures binary (dev-only fake data)**

Workspace `00000000-0000-4000-8000-000000000110` gets a `pro` lifetime subscription, payer user `110` (`user@notopia.com`, `User`, from `deploy/compose/infrastructure/authentik-blueprints/users.yaml`). Built with domain constructors and saved through the billing repositories. No Stripe calls: provider ids are fake (`cus_fixture_110`, `pi_fixture_110`), so Stripe actions on this workspace (portal, cancel-now refund, payer claim) fail; entitlements, gRPC and the UI work. Needs the seed first (`pro_lifetime` price).

Layout mirrors the seed: `cmd/billingfixtures/{main.go,wire.go,wire_gen.go,project.json,.env,.gitignore}`, `internal/billingfixtures/{config/{config.go,wire.go},fixtures.go,wire.go}`. Config (`NOTOPIA_BILLINGFIXTURES_*`): `General`, `Log`, `Database` only. `wire.go`: `ProvideFixtures`, `config.ProviderSet`, `persistence.ProviderSet`, `logging.ProviderSet`, `otel.ProviderSet`.

```go
// internal/billingfixtures/fixtures.go
package billingfixtures

var (
	workspaceID = uuid.MustParse("00000000-0000-4000-8000-000000000110")
	payer       = domain.Payer{ID: "110", Email: "user@notopia.com", Name: "User"}
)

const actor = "110"

type Fixtures struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

// Run is idempotent: skips when the workspace already has a subscription.
func (f *Fixtures) Run(ctx context.Context) error {
	now := time.Now()
	return f.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		existing, err := r.Subscription().GetCurrentByWorkspace(ctx, workspaceID, true)
		if err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetActive(ctx, domain.PlanTierPro, domain.IntervalLifetime, "vnd", false)
		if err != nil {
			return err
		}
		if price == nil {
			return errors.New("pro lifetime price missing: run just seed-billing first")
		}
		account := domain.OpenBillingAccount(workspaceID, "cus_fixture_110", payer, now, actor)
		if err := r.BillingAccount().Save(ctx, account); err != nil {
			return err
		}
		sub, err := domain.NewLifetimeSubscription(uuid.Must(uuid.NewV7()), workspaceID, price, "pi_fixture_110", now, now, actor)
		if err != nil {
			return err
		}
		return r.Subscription().Save(ctx, sub)
	})
}
```

`cmd/billingfixtures/.env`: `NOTOPIA_BILLINGFIXTURES_*` for log + `DATABASE_URL` (same DB as billing), `OTEL_SERVICE_NAME="billingfixtures"`. `project.json` like the seed's.

`justfile` (append to the existing `seed` / `fixtures` flows):

```just
seed: seed-search seed-billing

seed-billing:
    nx run billingseed:run

fixtures: seed fixtures-note fixtures-document fixtures-search fixtures-billing

fixtures-billing:
    nx run billingfixtures:run

stripe-listen:
    stripe listen --forward-to http://api.notopia.localhost/billing/webhooks/stripe
```

- [x] **Step 6: Verify (Stripe sandbox)**

```bash
nx run billing:gen:wire && nx run billingseed:gen:wire && nx run billingfixtures:gen:wire
nx build billing && nx lint billing && nx lint billingseed && nx lint billingfixtures
just seed-billing
curl -s http://api.notopia.localhost/billing/plans | jq '.[] | {tier, prices: [.prices[].lookupKey]}'
just fixtures-billing
```

Expected: plus has `plus_month`, `plus_year`; pro has `pro_month`, `pro_year`, `pro_lifetime`. Stripe dashboard (test mode) shows products `notopia_plus`/`notopia_pro` and one portal configuration. Running `just seed-billing` again creates nothing new. After fixtures, gRPC `GetWorkspaceEntitlements` for workspace 110 → `PLAN_TIER_PRO`, `SUBSCRIPTION_STATUS_ACTIVE`; running `just fixtures-billing` again changes nothing.

As `admin` (group `notopia-admins`): `PATCH /billing/admin/plans/free {"limits":{"members":4}}` → 200; `{"limits":{"seats":4}}` → 400 (unknown limit key rejected by `toLimitKeys`); as a normal user → 403. `POST /billing/admin/prices {"plan":"plus","interval":"month","amount":59000,"currency":"vnd"}` → old `plus_month` archived, new one active, a new portal configuration id in `billing_policy`.

- [x] **Step 7: Commit**

```bash
git add internal/billing internal/billingseed internal/billingfixtures cmd/billing cmd/billingseed cmd/billingfixtures justfile
git commit -m "feat(billing): admin catalog APIs, seed, fixtures and portal configuration"
```

---

### Task 8: Checkout, portal, webhook pipeline

**Files:**

- Create: `internal/billing/app/{q_startcheckout.go,q_startportal.go,refunds.go,cmd_completecheckout.go,cmd_syncsubscription.go,cmd_syncrefund.go}`
- Create: `internal/billing/app/{memstore_test.go,cmd_completecheckout_test.go}`
- Create: `internal/billing/.mockery.yaml`, `internal/billing/project.json` (nx `internal/billing`, target `gen:mockery`); modify `cmd/billing/project.json` (`gen:mockery` → `^gen:mockery`, `gen` depends on it); generated `internal/billing/app/mock.go` (committed, never hand-edited)
- Create: `internal/billing/controller/http/{checkout.go,webhook.go,webhook_test.go}`; modify `http.go` (`NewStrictHandler` takes `*config.Stripe`), `unimplemented.go`
- Modify: `internal/billing/app/app.go`, `internal/billing/app/wire.go`

**Interfaces:**

- Consumes: `PaymentProvider` (Task 6), `Notifier`, `AuthorizationSvc`, `Settings`.
- Produces: `Queries.StartCheckout` (`StartCheckout{WorkspaceID, PriceID uuid.UUID; Payer app.Payer}` → checkout URL `string`), `Queries.StartPortal` (`StartPortal{WorkspaceID uuid.UUID; UserID string}` → portal URL `string`), both modeled as queries because commands only return `error` and the provider session URL is the answer, `Cmds.CompleteCheckout` (`CompleteCheckout{EventID, SessionID string}`), `Cmds.SyncRefund` (`SyncRefund{EventID, EventType, ProviderRefundID string}`), `Cmds.SyncSubscription` (`SyncSubscription{EventID, EventType, ProviderSubscriptionID string}` — `EventID`/`EventType` empty when called by the reconcile cron), `RefundRunner.Run(ctx, subscriptionID)` (used by Tasks 9, 10, 11).

**Ordering rule used by every webhook use case:** provider _mutations that must not be lost_ (cancel a duplicate subscription, change default card) run **before** the write transaction and are idempotent; the write transaction records the event (`ProviderEventLog.MarkProcessed`) together with the state change; **refunds run after commit** with idempotency key `refund:{id}` and are retried by the reconcile cron while `pending` without a provider id.

- [x] **Step 1: Mocks and failing tests**

Mocks come from mockery (`mise.toml` pins it), same setup as `internal/authorization` (`internal/authorization/.mockery.yaml` + `internal/authorization/project.json` `gen:mockery`, pulled in by `cmd/authorization` `gen`). Mock only the ports whose calls a test asserts:

| Interface | Why a mock |
| --- | --- |
| `app.PaymentProvider` | External system; tests assert cancel/refund calls and stub snapshots/payments |
| `app.Notifier` | Fire-and-forget side effect; tests only allow it (`.Maybe()`) |
| `app.CompleteCheckoutCmd`, `app.SyncSubscriptionCmd`, `app.SyncRefundCmd` | Webhook controller test asserts which command the event routes to |

Not mocked: `domain.UnitOfWork` / repositories. Idempotency and "one current subscription" are state, so tests use the small in-memory `memStore` below; scripting every repo call with expectations would test the call order, not the behaviour. `AuthorizationSvc`, `BillingReadModel` and the other `Cmds`/`Queries` stay unmocked until a test needs them.

```yaml
# internal/billing/.mockery.yaml
template: testify
dir: "{{ .InterfaceDir }}"
filename: mock.go
pkgname: "{{ .SrcPackageName }}"
packages:
  github.com/notopia-uit/notopia/internal/billing/app:
    config:
      include-interface-regex: "^(PaymentProvider|Notifier|CompleteCheckoutCmd|SyncSubscriptionCmd|SyncRefundCmd)$"
```

```json
// internal/billing/project.json
{
  "$schema": "../../node_modules/nx/schemas/project-schema.json",
  "name": "internal/billing",
  "projectType": "library",
  "tags": ["scope:library", "type:go"],
  "namedInputs": {
    "default": ["{projectRoot}/**/*.go", "!{projectRoot}/**/*_test.go", "!{projectRoot}/**/mock.go"]
  },
  "targets": {
    "gen:mockery": {
      "executor": "nx:run-commands",
      "inputs": ["default", "{projectRoot}/.mockery.yaml"],
      "outputs": ["{projectRoot}/app/mock.go"],
      "options": {
        "command": "mockery",
        "cwd": "{projectRoot}"
      }
    }
  },
  "implicitDependencies": ["pkg"]
}
```

`cmd/billing/project.json`: add `"gen:mockery": {"dependsOn": ["^gen:mockery"]}` and make `gen` depend on `["gen:wire", "gen:sqlc", "gen:mockery"]`. Generate after Step 4 adds the `Sync*`/`CompleteCheckout` command types: `nx run billing:gen:mockery`. Generated types: `MockPaymentProvider`, `MockNotifier`, `MockCompleteCheckoutCmd`, … with `NewMockX(t)` (expectations asserted on cleanup) and `.EXPECT()`.

```go
// internal/billing/app/memstore_test.go
package app

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type memStore struct {
	mu       sync.Mutex
	subs     map[uuid.UUID]*domain.Subscription
	accounts map[uuid.UUID]*domain.BillingAccount
	prices   map[uuid.UUID]*domain.Price
	plans    map[domain.PlanTier]*domain.Plan
	policy   *domain.BillingPolicy
	events   map[string]bool
}

func newMemStore() *memStore {
	now := time.Now()
	s := &memStore{
		subs: map[uuid.UUID]*domain.Subscription{}, accounts: map[uuid.UUID]*domain.BillingAccount{},
		prices: map[uuid.UUID]*domain.Price{}, plans: map[domain.PlanTier]*domain.Plan{},
		policy: domain.DefaultBillingPolicy(now), events: map[string]bool{},
	}
	for _, p := range domain.DefaultPlans(now) {
		s.plans[p.Tier()] = p
	}
	return s
}

func (m *memStore) Execute(ctx context.Context, fn func(domain.RepoRegistry) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(m)
}

func (m *memStore) Subscription() domain.SubscriptionRepo     { return memSubs{m} }
func (m *memStore) BillingAccount() domain.BillingAccountRepo { return memAccounts{m} }
func (m *memStore) Price() domain.PriceRepo                   { return memPrices{m} }
func (m *memStore) Plan() domain.PlanRepo                     { return memPlans{m} }
func (m *memStore) BillingPolicy() domain.BillingPolicyRepo   { return memPolicy{m} }
func (m *memStore) ProviderEventLog() domain.ProviderEventLog { return memEvents{m} }

type memSubs struct{ m *memStore }

func (r memSubs) find(pred func(*domain.Subscription) bool) *domain.Subscription {
	for _, s := range r.m.subs {
		if pred(s) {
			return s
		}
	}
	return nil
}
func (r memSubs) GetByID(_ context.Context, id uuid.UUID, _ bool) (*domain.Subscription, error) {
	if s, ok := r.m.subs[id]; ok {
		return s, nil
	}
	return nil, errs.NewNotFound("sub")
}
func (r memSubs) GetCurrentByWorkspace(_ context.Context, ws uuid.UUID, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { return s.WorkspaceID() == ws && s.IsCurrent() }), nil
}
func (r memSubs) GetByProviderSubscriptionID(_ context.Context, id string, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { p := s.ProviderSubscriptionID(); return p != nil && *p == id }), nil
}
func (r memSubs) GetByProviderPaymentID(_ context.Context, id string, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { p := s.ProviderPaymentID(); return p != nil && *p == id }), nil
}
func (r memSubs) GetByRefundID(_ context.Context, id uuid.UUID, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool {
		for _, rf := range s.Refunds() {
			if rf.ID() == id {
				return true
			}
		}
		return false
	}), nil
}
func (r memSubs) ListCurrentByWorkspaces(_ context.Context, ids []uuid.UUID) ([]*domain.Subscription, error) {
	var out []*domain.Subscription
	for _, s := range r.m.subs {
		for _, id := range ids {
			if s.WorkspaceID() == id && s.IsCurrent() {
				out = append(out, s)
			}
		}
	}
	return out, nil
}
func (r memSubs) ListNeedingReconcile(context.Context, time.Time, int) ([]*domain.Subscription, error) {
	return nil, nil
}
func (r memSubs) Save(_ context.Context, s *domain.Subscription) error {
	if s.IsCurrent() {
		if cur := r.find(func(o *domain.Subscription) bool { return o.WorkspaceID() == s.WorkspaceID() && o.IsCurrent() && o.ID() != s.ID() }); cur != nil {
			return errs.NewInternal("unique current subscription violated", nil)
		}
	}
	s.PopEvents()
	r.m.subs[s.ID()] = s
	return nil
}

type memAccounts struct{ m *memStore }

func (r memAccounts) GetByWorkspace(_ context.Context, ws uuid.UUID, _ bool) (*domain.BillingAccount, error) {
	return r.m.accounts[ws], nil
}
func (r memAccounts) ListByPayer(_ context.Context, payerID string) ([]*domain.BillingAccount, error) {
	var out []*domain.BillingAccount
	for _, a := range r.m.accounts {
		if a.Payer().ID == payerID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r memAccounts) Save(_ context.Context, a *domain.BillingAccount) error {
	a.PopEvents()
	r.m.accounts[a.WorkspaceID()] = a
	return nil
}

type memPrices struct{ m *memStore }

func (r memPrices) GetByID(_ context.Context, id uuid.UUID, _ bool) (*domain.Price, error) {
	if p, ok := r.m.prices[id]; ok {
		return p, nil
	}
	return nil, errs.NewNotFound("price")
}
func (r memPrices) GetByProviderPriceID(_ context.Context, id string) (*domain.Price, error) {
	for _, p := range r.m.prices {
		if p.ProviderPriceID() == id {
			return p, nil
		}
	}
	return nil, errs.NewNotFound("price")
}
func (r memPrices) GetActive(_ context.Context, plan domain.PlanTier, iv domain.Interval, cur string, _ bool) (*domain.Price, error) {
	for _, p := range r.m.prices {
		if p.Active() && p.Plan() == plan && p.Interval() == iv && p.Amount().Currency == cur {
			return p, nil
		}
	}
	return nil, nil
}
func (r memPrices) ListActive(context.Context) ([]*domain.Price, error) {
	var out []*domain.Price
	for _, p := range r.m.prices {
		if p.Active() {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r memPrices) List(ctx context.Context) ([]*domain.Price, error) { return r.ListActive(ctx) }
func (r memPrices) Save(_ context.Context, p *domain.Price) error {
	p.PopEvents()
	r.m.prices[p.ID()] = p
	return nil
}

type memPlans struct{ m *memStore }

func (r memPlans) Get(_ context.Context, t domain.PlanTier, _ bool) (*domain.Plan, error) { return r.m.plans[t], nil }
func (r memPlans) ListMap(context.Context) (map[domain.PlanTier]*domain.Plan, error)     { return r.m.plans, nil }
func (r memPlans) Save(_ context.Context, p *domain.Plan) error                          { r.m.plans[p.Tier()] = p; return nil }

type memPolicy struct{ m *memStore }

func (r memPolicy) Get(context.Context, bool) (*domain.BillingPolicy, error)     { return r.m.policy, nil }
func (r memPolicy) Save(_ context.Context, p *domain.BillingPolicy) error        { r.m.policy = p; return nil }

type memEvents struct{ m *memStore }

func (r memEvents) MarkProcessed(_ context.Context, id, _ string) (bool, error) {
	if r.m.events[id] {
		return false, nil
	}
	r.m.events[id] = true
	return true, nil
}
func (r memEvents) PruneBefore(context.Context, time.Time) (int64, error) { return 0, nil }
```

```go
// internal/billing/app/cmd_completecheckout_test.go
package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testPayer = domain.Payer{ID: "112", Email: "a@b.c", Name: "A"}

type fixture struct {
	store    *memStore
	provider *MockPaymentProvider
	handler  *CompleteCheckoutHandler
	ws       uuid.UUID
	plus     *domain.Price
	lifetime *domain.Price
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	store := newMemStore()
	now := time.Now()
	plus, _ := domain.NewPrice(uuid.New(), domain.PlanTierPlus, domain.IntervalMonth, domain.Money{Amount: 30000, Currency: "vnd"}, "price_plus", now, "u")
	lifetime, _ := domain.NewPrice(uuid.New(), domain.PlanTierPro, domain.IntervalLifetime, domain.Money{Amount: 2490000, Currency: "vnd"}, "price_life", now, "u")
	store.prices[plus.ID()], store.prices[lifetime.ID()] = plus, lifetime
	ws := uuid.New()
	store.accounts[ws] = domain.OpenBillingAccount(ws, "cus_1", testPayer, now, "112")

	provider := NewMockPaymentProvider(t)
	provider.EXPECT().UpdateCustomerContact(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	notifier := NewMockNotifier(t)
	notifier.EXPECT().SubscriptionChanged(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	settings := &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"}
	refunds := NewRefundRunner(store, provider, settings)
	return &fixture{
		store:    store,
		provider: provider,
		ws:       ws,
		plus:     plus,
		lifetime: lifetime,
		handler:  NewCompleteCheckoutHandler(store, provider, notifier, settings, refunds),
	}
}

// expectSubscribeSession stubs a paid plus-monthly checkout that created Stripe subscription subID.
func (f *fixture) expectSubscribeSession(subID string) {
	now := time.Now()
	end := now.Add(30 * 24 * time.Hour)
	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_"+subID).Return(CheckoutSessionResult{
		ID: "cs_" + subID, Purpose: CheckoutPurposeSubscribe, Complete: true, WorkspaceID: f.ws, PriceID: f.plus.ID(),
		Payer: testPayer, CustomerID: "cus_1", ProviderSubscriptionID: subID,
	}, nil)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, subID).Return(domain.ProviderSnapshot{
		ProviderSubscriptionID: subID, ProviderPriceID: "price_plus", WorkspaceID: f.ws, Status: domain.StatusActive,
		CurrentPeriodStart: now, CurrentPeriodEnd: end,
	}, nil).Maybe()
	f.provider.EXPECT().PeriodPayments(mock.Anything, subID, mock.Anything, mock.Anything).Return([]domain.PaymentInfo{{
		Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_" + subID, PaidAt: now,
		CoverageStart: now, CoverageEnd: end,
	}}, nil).Maybe()
}

// expectRefund accepts one refund and reports it succeeded; returns the captured amount.
func (f *fixture) expectRefund() *domain.Money {
	var got domain.Money
	f.provider.EXPECT().CreateRefund(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, id uuid.UUID, ref string, amount domain.Money) (RefundResult, error) {
			got = amount
			return RefundResult{
				ProviderRefundID: "re_" + id.String(), RefundID: &id, PaymentRef: ref,
				Amount: amount, Status: domain.RefundStatusSucceeded, FailureMessage: nil,
			}, nil
		}).Once()
	return &got
}

func (f *fixture) current(t *testing.T) *domain.Subscription {
	t.Helper()
	s, _ := f.store.Subscription().GetCurrentByWorkspace(context.Background(), f.ws, false)
	return s
}

func TestCompleteCheckoutIsIdempotent(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession("sub_1")
	cmd := &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}
	for range 2 {
		require.NoError(t, f.handler.Handle(context.Background(), cmd))
	}
	assert.Len(t, f.store.subs, 1)
	assert.Equal(t, domain.PlanTierPlus, f.current(t).Plan())
	// no CancelSubscriptionNow / CreateRefund expectation: the mock fails the test if either is called
}

func TestDuplicateCheckoutIsCanceledAndRefunded(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession("sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.expectSubscribeSession("sub_2")
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_2").Return(nil).Once()
	refunded := f.expectRefund()
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_sub_2"}))

	assert.Equal(t, "sub_1", *f.current(t).ProviderSubscriptionID())
	assert.EqualValues(t, 30000, refunded.Amount)
}

func TestLifetimeReplacesRecurring(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession("sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_life").Return(CheckoutSessionResult{
		ID: "cs_life", Purpose: CheckoutPurposeLifetime, Complete: true, WorkspaceID: f.ws, PriceID: f.lifetime.ID(),
		Payer: testPayer, CustomerID: "cus_1",
		Payment: domain.PaymentInfo{Amount: domain.Money{Amount: 2490000, Currency: "vnd"}, PaymentRef: "pi_life", PaidAt: time.Now()},
	}, nil)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()
	refunded := f.expectRefund()
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_life"}))

	cur := f.current(t)
	assert.Equal(t, domain.IntervalLifetime, cur.Interval())
	assert.Equal(t, domain.PlanTierPro, cur.Plan())
	assert.Positive(t, refunded.Amount)
}
```

Run: `go test ./internal/billing/app/...` → FAIL (handlers missing).

- [x] **Step 2: Refund runner**

```go
// internal/billing/app/refunds.go
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type RefundRunner struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func NewRefundRunner(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *RefundRunner {
	return &RefundRunner{uow: uow, provider: provider, settings: settings}
}

var ProvideRefundRunner = NewRefundRunner

// Run submits every pending refund of the subscription that has no provider id yet.
func (r *RefundRunner) Run(ctx context.Context, subscriptionID uuid.UUID) error {
	var pending []*domain.Refund
	if err := r.uow.Execute(ctx, func(reg domain.RepoRegistry) error {
		sub, err := reg.Subscription().GetByID(ctx, subscriptionID, false)
		if err != nil {
			return err
		}
		for _, rf := range sub.Refunds() {
			if rf.Status() == domain.RefundStatusPending && rf.ProviderRefundID() == nil {
				pending = append(pending, rf)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, rf := range pending {
		res, err := r.provider.CreateRefund(ctx, rf.ID(), rf.PaymentRef(), rf.Amount())
		if err != nil {
			slog.ErrorContext(ctx, "refund submission failed; cron will retry", slog.String("refund_id", rf.ID().String()), slog.Any("error", err))
			continue
		}
		if err := r.uow.Execute(ctx, func(reg domain.RepoRegistry) error {
			sub, err := reg.Subscription().GetByID(ctx, subscriptionID, true)
			if err != nil {
				return err
			}
			if err := sub.UpdateRefund(rf.ID(), res.ProviderRefundID, res.Status, res.FailureMessage, time.Now(), r.settings.ServiceAccountUserID); err != nil {
				return err
			}
			return reg.Subscription().Save(ctx, sub)
		}); err != nil {
			return err
		}
	}
	return nil
}
```

- [x] **Step 3: CompleteCheckout**

```go
// internal/billing/app/cmd_completecheckout.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CompleteCheckout struct {
	EventID   string
	SessionID string
}

type CompleteCheckoutHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	settings *Settings
	refunds  *RefundRunner
}

func NewCompleteCheckoutHandler(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, settings *Settings, refunds *RefundRunner) *CompleteCheckoutHandler {
	return &CompleteCheckoutHandler{uow: uow, provider: provider, notifier: notifier, settings: settings, refunds: refunds}
}

var ProvideCompleteCheckoutHandler = NewCompleteCheckoutHandler

type CompleteCheckoutCmd commonhandler.Cmd[CompleteCheckout]

var _ CompleteCheckoutCmd = (*CompleteCheckoutHandler)(nil)

const eventCheckoutCompleted = "checkout.session.completed"

func (h *CompleteCheckoutHandler) Handle(ctx context.Context, cmd *CompleteCheckout) error {
	session, err := h.provider.GetCheckoutSession(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if !session.Complete {
		return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
			_, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted)
			return err
		})
	}
	switch session.Purpose {
	case CheckoutPurposeSubscribe:
		return h.subscribe(ctx, cmd, session)
	case CheckoutPurposeLifetime:
		return h.lifetime(ctx, cmd, session)
	case CheckoutPurposePayerChange:
		return h.payerChange(ctx, cmd, session)
	}
	return nil
}

func (h *CompleteCheckoutHandler) currentSub(ctx context.Context, ws uuid.UUID) (*domain.Subscription, error) {
	var cur *domain.Subscription
	err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		cur, err = r.Subscription().GetCurrentByWorkspace(ctx, ws, false)
		return err
	})
	return cur, err
}

func (h *CompleteCheckoutHandler) setPayer(ctx context.Context, r domain.RepoRegistry, ws uuid.UUID, payer domain.Payer, now time.Time) error {
	account, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
	if err != nil || account == nil {
		return err
	}
	if account.Payer().ID != payer.ID || account.PayerMissingSince() != nil {
		account.ChangePayer(payer, now, payer.ID)
		return r.BillingAccount().Save(ctx, account)
	}
	return nil
}

func (h *CompleteCheckoutHandler) subscribe(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	snap, err := h.provider.GetSubscriptionSnapshot(ctx, s.ProviderSubscriptionID)
	if err != nil {
		return err
	}
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	duplicate := cur != nil && (cur.ProviderSubscriptionID() == nil || *cur.ProviderSubscriptionID() != s.ProviderSubscriptionID)
	var payments []domain.PaymentInfo
	if duplicate {
		if payments, err = h.provider.PeriodPayments(ctx, s.ProviderSubscriptionID, snap.CurrentPeriodStart, snap.CurrentPeriodEnd); err != nil {
			return err
		}
		if err := h.provider.CancelSubscriptionNow(ctx, s.ProviderSubscriptionID); err != nil {
			return err
		}
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	var saved *domain.Subscription
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if existing, err := r.Subscription().GetByProviderSubscriptionID(ctx, s.ProviderSubscriptionID, true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByProviderPriceID(ctx, snap.ProviderPriceID)
		if err != nil {
			return err
		}
		sub, err := domain.NewSubscription(uuid.Must(uuid.NewV7()), s.WorkspaceID, price, snap, now, actor)
		if err != nil {
			return err
		}
		if duplicate {
			if _, err := sub.CancelNow(now, domain.RefundReasonDuplicateCheckout, domain.FullRefundQuote(payments...), actor); err != nil {
				return err
			}
		} else if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		saved = sub
		return r.Subscription().Save(ctx, sub)
	})
	if err != nil || saved == nil {
		return err
	}
	if duplicate {
		return h.refunds.Run(ctx, saved.ID())
	}
	notifySubscription(ctx, h.notifier, saved)
	return nil
}

func (h *CompleteCheckoutHandler) lifetime(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	var replaced []domain.PaymentInfo
	replacing := cur != nil && cur.Interval() != domain.IntervalLifetime
	duplicate := cur != nil && cur.Interval() == domain.IntervalLifetime &&
		(cur.ProviderPaymentID() == nil || *cur.ProviderPaymentID() != s.Payment.PaymentRef)
	if replacing {
		if replaced, err = h.provider.PeriodPayments(ctx, *cur.ProviderSubscriptionID(), *cur.CurrentPeriodStart(), *cur.CurrentPeriodEnd()); err != nil {
			return err
		}
		if err := h.provider.CancelSubscriptionNow(ctx, *cur.ProviderSubscriptionID()); err != nil {
			return err
		}
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	var toRefund []uuid.UUID
	var saved *domain.Subscription
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if existing, err := r.Subscription().GetByProviderPaymentID(ctx, s.Payment.PaymentRef, true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByID(ctx, s.PriceID, false)
		if err != nil {
			return err
		}
		policy, err := r.BillingPolicy().Get(ctx, false)
		if err != nil {
			return err
		}
		if replacing {
			old, err := r.Subscription().GetByID(ctx, cur.ID(), true)
			if err != nil {
				return err
			}
			quote, err := old.QuoteCancelNow(now, policy, replaced)
			if err != nil {
				return err
			}
			if _, err := old.CancelNow(now, domain.RefundReasonReplacedByLifetime, quote, actor); err != nil {
				return err
			}
			if err := r.Subscription().Save(ctx, old); err != nil {
				return err
			}
			toRefund = append(toRefund, old.ID())
		}
		sub, err := domain.NewLifetimeSubscription(uuid.Must(uuid.NewV7()), s.WorkspaceID, price, s.Payment.PaymentRef, s.Payment.PaidAt, now, actor)
		if err != nil {
			return err
		}
		if duplicate {
			if _, err := sub.CancelNow(now, domain.RefundReasonDuplicateCheckout, domain.FullRefundQuote(s.Payment), actor); err != nil {
				return err
			}
			toRefund = append(toRefund, sub.ID())
		} else if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		saved = sub
		return r.Subscription().Save(ctx, sub)
	})
	if err != nil || saved == nil {
		return err
	}
	for _, id := range toRefund {
		if err := h.refunds.Run(ctx, id); err != nil {
			return err
		}
	}
	if !duplicate {
		notifySubscription(ctx, h.notifier, saved)
	}
	return nil
}

func (h *CompleteCheckoutHandler) payerChange(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	var account *domain.BillingAccount
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, s.WorkspaceID, false)
		return err
	}); err != nil || account == nil {
		return err
	}
	resume := account.PayerMissingSince() != nil && cur != nil && cur.CancelAtPeriodEnd() && cur.ProviderSubscriptionID() != nil
	var providerSubID *string
	if cur != nil {
		providerSubID = cur.ProviderSubscriptionID()
	}
	if err := h.provider.SetDefaultPaymentMethod(ctx, account.ProviderCustomerID(), s.PaymentMethodID, providerSubID); err != nil {
		return err
	}
	if err := h.provider.DetachPaymentMethods(ctx, account.ProviderCustomerID(), s.PaymentMethodID); err != nil {
		return err
	}
	if err := h.provider.UpdateCustomerContact(ctx, account.ProviderCustomerID(), s.Payer); err != nil {
		return err
	}
	if resume {
		if err := h.provider.SetCancelAtPeriodEnd(ctx, *cur.ProviderSubscriptionID(), false); err != nil {
			return err
		}
	}

	now := time.Now()
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		if !resume {
			return nil
		}
		sub, err := r.Subscription().GetByID(ctx, cur.ID(), true)
		if err != nil {
			return err
		}
		sub.ResumeCancel(now, s.Payer.ID)
		return r.Subscription().Save(ctx, sub)
	})
	if err == nil && cur != nil {
		notifySubscription(ctx, h.notifier, cur)
	}
	return err
}
```

`setPayer` uses `payer.ID` as actor (the user who completed checkout).

- [x] **Step 4: SyncSubscription, SyncRefund, StartCheckout (query), StartPortal (query)**

```go
// internal/billing/app/cmd_syncsubscription.go
package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SyncSubscription struct {
	EventID                string // empty when called by the reconcile cron
	EventType              string
	ProviderSubscriptionID string
}

type SyncSubscriptionHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	settings *Settings
}

func NewSyncSubscriptionHandler(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, settings *Settings) *SyncSubscriptionHandler {
	return &SyncSubscriptionHandler{uow: uow, provider: provider, notifier: notifier, settings: settings}
}

var ProvideSyncSubscriptionHandler = NewSyncSubscriptionHandler

type SyncSubscriptionCmd commonhandler.Cmd[SyncSubscription]

var _ SyncSubscriptionCmd = (*SyncSubscriptionHandler)(nil)

func (h *SyncSubscriptionHandler) Handle(ctx context.Context, cmd *SyncSubscription) error {
	snap, err := h.provider.GetSubscriptionSnapshot(ctx, cmd.ProviderSubscriptionID)
	if err != nil {
		return err
	}
	var changed *domain.Subscription
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if cmd.EventID != "" {
			if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
				return err
			}
		}
		sub, err := r.Subscription().GetByProviderSubscriptionID(ctx, cmd.ProviderSubscriptionID, true)
		if err != nil || sub == nil {
			return err // unknown yet: checkout.session.completed creates it
		}
		if !sub.IsCurrent() {
			return nil // canceled locally (cancel-now / duplicate / replaced): never resurrect
		}
		price, err := r.Price().GetByProviderPriceID(ctx, snap.ProviderPriceID)
		if err != nil {
			return err
		}
		beforeStatus, beforePlan, beforeCancel, beforeEnd := sub.Status(), sub.Plan(), sub.CancelAtPeriodEnd(), sub.CurrentPeriodEnd()
		sub.ApplySnapshot(snap, price, time.Now(), h.settings.ServiceAccountUserID)
		end := sub.CurrentPeriodEnd()
		if beforeStatus != sub.Status() || beforePlan != sub.Plan() || beforeCancel != sub.CancelAtPeriodEnd() ||
			beforeEnd == nil || end == nil || !beforeEnd.Equal(*end) {
			changed = sub
		}
		return r.Subscription().Save(ctx, sub)
	})
	if err == nil {
		notifySubscription(ctx, h.notifier, changed)
	}
	return err
}
```

```go
// internal/billing/app/cmd_syncrefund.go
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SyncRefund struct {
	EventID          string
	EventType        string
	ProviderRefundID string
}

type SyncRefundHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func NewSyncRefundHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *SyncRefundHandler {
	return &SyncRefundHandler{uow: uow, provider: provider, settings: settings}
}

var ProvideSyncRefundHandler = NewSyncRefundHandler

type SyncRefundCmd commonhandler.Cmd[SyncRefund]

var _ SyncRefundCmd = (*SyncRefundHandler)(nil)

func (h *SyncRefundHandler) Handle(ctx context.Context, cmd *SyncRefund) error {
	res, err := h.provider.GetRefund(ctx, cmd.ProviderRefundID)
	if err != nil {
		return err
	}
	now, actor := time.Now(), h.settings.ServiceAccountUserID
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
			return err
		}
		if res.RefundID != nil {
			sub, err := r.Subscription().GetByRefundID(ctx, *res.RefundID, true)
			if err != nil || sub == nil {
				return err
			}
			if err := sub.UpdateRefund(*res.RefundID, res.ProviderRefundID, res.Status, res.FailureMessage, now, actor); err != nil {
				return err
			}
			return r.Subscription().Save(ctx, sub)
		}
		sub, err := r.Subscription().GetByProviderPaymentID(ctx, res.PaymentRef, true)
		if err != nil {
			return err
		}
		if sub == nil {
			slog.InfoContext(ctx, "ignoring external refund of an unknown payment", slog.String("refund", res.ProviderRefundID))
			return nil
		}
		sub.RecordExternalRefund(res.ProviderRefundID, res.PaymentRef, res.Amount, res.Status, now, actor)
		return r.Subscription().Save(ctx, sub)
	})
}
```

```go
// internal/billing/app/q_startcheckout.go
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type StartCheckout struct {
	WorkspaceID uuid.UUID
	PriceID     uuid.UUID
	Payer       Payer
}

type StartCheckoutHandler struct {
	uow      domain.UnitOfWork
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartCheckoutHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *StartCheckoutHandler {
	return &StartCheckoutHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideStartCheckoutHandler = NewStartCheckoutHandler

type StartCheckoutQuery commonhandler.Query[StartCheckout, string]

var _ StartCheckoutQuery = (*StartCheckoutHandler)(nil)

func billingPageURL(settings *Settings, ws uuid.UUID, result string) string {
	return fmt.Sprintf("%s/workspace/%s/settings/billing?checkout=%s", settings.WebBaseURL, ws, result)
}

// ensureAccount opens the billing account (and Stripe customer) on first use.
func ensureAccount(ctx context.Context, uow domain.UnitOfWork, provider PaymentProvider, ws uuid.UUID, payer domain.Payer) (*domain.BillingAccount, error) {
	var account *domain.BillingAccount
	if err := uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, ws, false)
		return err
	}); err != nil || account != nil {
		return account, err
	}
	customerID, err := provider.CreateCustomer(ctx, ws, payer)
	if err != nil {
		return nil, err
	}
	err = uow.Execute(ctx, func(r domain.RepoRegistry) error {
		existing, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
		if err != nil || existing != nil {
			account = existing
			return err
		}
		account = domain.OpenBillingAccount(ws, customerID, payer, time.Now(), payer.ID)
		return r.BillingAccount().Save(ctx, account)
	})
	return account, err
}

func (h *StartCheckoutHandler) Handle(ctx context.Context, q *StartCheckout) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var price *domain.Price
	var cur *domain.Subscription
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if price, err = r.Price().GetByID(ctx, q.PriceID, false); err != nil {
			return err
		}
		cur, err = r.Subscription().GetCurrentByWorkspace(ctx, q.WorkspaceID, false)
		return err
	}); err != nil {
		return "", err
	}
	if !price.Active() {
		return "", errs.New(errs.CodePriceNotActive, "price is archived", nil)
	}
	lifetimeUpgrade := price.Interval() == domain.IntervalLifetime && cur != nil && cur.Interval() != domain.IntervalLifetime
	if cur != nil && !lifetimeUpgrade {
		return "", errs.New(errs.CodeSubscriptionAlreadyActive, "workspace already has a subscription; use the portal to change plan", nil)
	}
	payer := toDomainPayer(q.Payer)
	account, err := ensureAccount(ctx, h.uow, h.provider, q.WorkspaceID, payer)
	if err != nil {
		return "", err
	}
	purpose := CheckoutPurposeSubscribe
	if price.Interval() == domain.IntervalLifetime {
		purpose = CheckoutPurposeLifetime
	}
	return h.provider.CreateCheckoutSession(ctx, CheckoutInput{
		Purpose: purpose, WorkspaceID: q.WorkspaceID, CustomerID: account.ProviderCustomerID(), Payer: payer,
		PriceID: price.ID(), ProviderPriceID: price.ProviderPriceID(),
		SuccessURL: billingPageURL(h.settings, q.WorkspaceID, "success"), CancelURL: billingPageURL(h.settings, q.WorkspaceID, "cancel"),
	})
}
```

```go
// internal/billing/app/q_startportal.go
package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type StartPortal struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type StartPortalHandler struct {
	uow      domain.UnitOfWork
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartPortalHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *StartPortalHandler {
	return &StartPortalHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideStartPortalHandler = NewStartPortalHandler

type StartPortalQuery commonhandler.Query[StartPortal, string]

var _ StartPortalQuery = (*StartPortalHandler)(nil)

func (h *StartPortalHandler) Handle(ctx context.Context, q *StartPortal) (string, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var account *domain.BillingAccount
	var policy *domain.BillingPolicy
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if account, err = r.BillingAccount().GetByWorkspace(ctx, q.WorkspaceID, false); err != nil {
			return err
		}
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	}); err != nil {
		return "", err
	}
	if account == nil {
		return "", errs.New(errs.CodeNoBillingAccount, "workspace has no billing account", nil)
	}
	if policy.PortalConfigurationID() == nil {
		return "", errs.NewInternal("portal configuration missing: run just seed-billing", nil)
	}
	return h.provider.CreatePortalSession(ctx, account.ProviderCustomerID(), *policy.PortalConfigurationID(),
		fmt.Sprintf("%s/workspace/%s/settings/billing?portal=return", h.settings.WebBaseURL, q.WorkspaceID))
}
```

Register `CompleteCheckout`, `SyncSubscription`, `SyncRefund` in `app.go` (Cmds) and `StartCheckout`, `StartPortal` (Queries) and the providers + `ProvideRefundRunner` in `wire.go`. The webhook controller calls the decorated `Cmds.*` directly, one per event type.

- [x] **Step 5: HTTP**

```go
// internal/billing/controller/http/checkout.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) CreateCheckoutSession(ctx context.Context, req billing.CreateCheckoutSessionRequestObject) (billing.CreateCheckoutSessionResponseObject, error) {
	payer, err := currentPayer(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.StartCheckout.Handle(ctx, &app.StartCheckout{WorkspaceID: req.WorkspaceId, PriceID: req.Body.PriceId, Payer: payer})
	if err != nil {
		return nil, err
	}
	return billing.CreateCheckoutSession200JSONResponse{Url: url}, nil
}

func (h *StrictHandler) CreatePortalSession(ctx context.Context, req billing.CreatePortalSessionRequestObject) (billing.CreatePortalSessionResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.StartPortal.Handle(ctx, &app.StartPortal{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	return billing.CreatePortalSession200JSONResponse{Url: url}, nil
}
```

The Stripe webhook is an inbound adapter, like `controller/grpc`: it verifies the signature, reads the event type and object id, and routes to one command. It is the only place outside `infra/payment/stripe` that imports stripe-go (`webhook` package only). `NewStrictHandler` gets `stripeCfg *config.Stripe` and stores `webhookSecret`.

```go
// internal/billing/controller/http/webhook.go
package http

import (
	"context"
	"encoding/json"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/stripe/stripe-go/v87/webhook"
)

type stripeObject struct {
	ID     string `json:"id"`
	Parent *struct {
		SubscriptionDetails *struct {
			Subscription json.RawMessage `json:"subscription"`
		} `json:"subscription_details"`
	} `json:"parent"`
}

// stripeID reads an expandable field: either "id" or {"id": ...}.
func stripeID(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(raw, &obj)
	return obj.ID
}

func (h *StrictHandler) StripeWebhook(ctx context.Context, req billing.StripeWebhookRequestObject) (billing.StripeWebhookResponseObject, error) {
	body := rawBody(ctx)
	if body == nil {
		return nil, errs.NewInvalid("missing body")
	}
	ev, err := webhook.ConstructEventWithOptions(
		body,
		req.Params.StripeSignature,
		h.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		return nil, errs.New(errs.CodeWebhookSignature, "invalid webhook signature", err)
	}
	var obj stripeObject
	if err := json.Unmarshal(ev.Data.Raw, &obj); err != nil {
		return nil, errs.NewInvalid("invalid webhook object")
	}

	switch string(ev.Type) {
	case "checkout.session.completed":
		err = h.App.Cmds.CompleteCheckout.Handle(ctx, &app.CompleteCheckout{EventID: ev.ID, SessionID: obj.ID})
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{EventID: ev.ID, EventType: string(ev.Type), ProviderSubscriptionID: obj.ID})
	case "invoice.paid", "invoice.payment_failed":
		if obj.Parent == nil || obj.Parent.SubscriptionDetails == nil {
			break
		}
		if sub := stripeID(obj.Parent.SubscriptionDetails.Subscription); sub != "" {
			err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{EventID: ev.ID, EventType: string(ev.Type), ProviderSubscriptionID: sub})
		}
	case "refund.created", "refund.updated", "refund.failed":
		err = h.App.Cmds.SyncRefund.Handle(ctx, &app.SyncRefund{EventID: ev.ID, EventType: string(ev.Type), ProviderRefundID: obj.ID})
	}
	if err != nil {
		return nil, err
	}
	return billing.StripeWebhook204Response{}, nil
}
```

`webhook_test.go` (controller test; build `StrictHandler` by hand with `&app.Server{Cmds: &app.Cmds{CompleteCheckout: app.NewMockCompleteCheckoutCmd(t), SyncSubscription: …, SyncRefund: …}}`, put the payload in ctx the way `captureRawBody` does). A mock with no expectation fails the test when called, so "no other command ran" needs no extra assert:

```go
func TestStripeWebhookInvoicePaidSyncsSubscription(t *testing.T) {
	payload := []byte(`{"id":"evt_1","object":"event","api_version":"2020-08-27","type":"invoice.paid","data":{"object":{"id":"in_1","object":"invoice","parent":{"subscription_details":{"subscription":"sub_123"}}}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: "whsec_test"})
	// … handler with webhookSecret "whsec_test"
	// syncSub.EXPECT().Handle(mock.Anything, &app.SyncSubscription{EventID: "evt_1", EventType: "invoice.paid", ProviderSubscriptionID: "sub_123"}).Return(nil).Once()
	// call StripeWebhook with signed.Header → 204
}

func TestStripeWebhookBadSignature(t *testing.T) {
	// same payload, header "t=1,v1=bad" → errs.CodeWebhookSignature; mocks have no expectations, so any call fails
}
```

(`Url` may be generated as `Url` or `URL`; `StripeSignature` is the header param field name — check `server.gen.go`.) Delete the three stubs.

- [x] **Step 6: Run tests, build, E2E**

```bash
nx run billing:gen:mockery && go test ./internal/billing/...
nx run billing:gen:wire && nx build billing && nx lint billing
just stripe-listen            # copy the whsec_ into cmd/billing/.env.local, restart billing
```

As owner of workspace W (user `kevinnitro`): `POST /billing/workspaces/W/checkout-session {"priceId":"<plus_month id>"}` → open the URL, pay with `4242 4242 4242 4242`. Then `GET /billing/workspaces/W/entitlements` → `plan: plus`, `status: active`. In `stripe listen` output every event returns 204. Replaying the last event from the Stripe dashboard ("Resend") changes nothing. `POST .../portal-session` opens the Portal; switching to `pro` monthly updates entitlements to `pro` after the `customer.subscription.updated` webhook.

- [x] **Step 7: Commit**

```bash
git add internal/billing cmd/billing/wire_gen.go cmd/billing/project.json
git commit -m "feat(billing): checkout, portal and stripe webhook processing"
```

---

### Task 9: Cancel now with refund

**Files:**

- Create: `internal/billing/app/{canceller.go,cmd_cancelnow.go,q_previewcancelnow.go}`
- Create: `internal/billing/controller/http/cancel.go`; modify `unimplemented.go`
- Modify: `internal/billing/app/app.go`, `wire.go`

**Interfaces:**

- Produces: `Canceller.CancelNow(ctx, workspaceID uuid.UUID, reason domain.RefundReason, actor string, refundOptional bool) error` (reused by Task 10); `Cmds.CancelNow` (`CancelNow{WorkspaceID uuid.UUID; UserID string}`); `Queries.PreviewCancelNow` (`PreviewCancelNow{WorkspaceID; UserID}` → `domain.RefundQuote`).

- [x] **Step 1: Canceller**

```go
// internal/billing/app/canceller.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Canceller struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	refunds  *RefundRunner
}

func NewCanceller(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, refunds *RefundRunner) *Canceller {
	return &Canceller{uow: uow, provider: provider, notifier: notifier, refunds: refunds}
}

var ProvideCanceller = NewCanceller

func (c *Canceller) load(ctx context.Context, ws uuid.UUID) (*domain.Subscription, *domain.BillingPolicy, error) {
	var sub *domain.Subscription
	var policy *domain.BillingPolicy
	err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if sub, err = r.Subscription().GetCurrentByWorkspace(ctx, ws, false); err != nil {
			return err
		}
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	})
	if err == nil && sub == nil {
		err = errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	return sub, policy, err
}

func (c *Canceller) payments(ctx context.Context, sub *domain.Subscription) ([]domain.PaymentInfo, error) {
	if sub.Interval() == domain.IntervalLifetime {
		p, err := c.provider.PaymentForIntent(ctx, *sub.ProviderPaymentID())
		return []domain.PaymentInfo{p}, err
	}
	if sub.CurrentPeriodStart() == nil || sub.CurrentPeriodEnd() == nil {
		return nil, errs.NewInternal("subscription has no period", nil)
	}
	return c.provider.PeriodPayments(ctx, *sub.ProviderSubscriptionID(), *sub.CurrentPeriodStart(), *sub.CurrentPeriodEnd())
}

func (c *Canceller) Quote(ctx context.Context, ws uuid.UUID) (domain.RefundQuote, error) {
	sub, policy, err := c.load(ctx, ws)
	if err != nil {
		return domain.RefundQuote{}, err
	}
	payments, err := c.payments(ctx, sub)
	if err != nil {
		return domain.RefundQuote{}, err
	}
	return sub.QuoteCancelNow(time.Now(), policy, payments)
}

// CancelNow cancels at the provider first, then locally, then submits the refund.
// refundOptional=true (workspace deleted) cancels even when no refund is due (lifetime past its window).
func (c *Canceller) CancelNow(ctx context.Context, ws uuid.UUID, reason domain.RefundReason, actor string, refundOptional bool) error {
	sub, policy, err := c.load(ctx, ws)
	if err != nil {
		return err
	}
	payments, err := c.payments(ctx, sub)
	if err != nil {
		return err
	}
	now := time.Now()
	quote, err := sub.QuoteCancelNow(now, policy, payments)
	if err != nil {
		if !refundOptional || !hasCode(err, errs.CodeRefundWindowExpired) {
			return err
		}
		quote = domain.RefundQuote{}
	}
	if sub.ProviderSubscriptionID() != nil {
		if err := c.provider.CancelSubscriptionNow(ctx, *sub.ProviderSubscriptionID()); err != nil {
			return err
		}
	}
	var saved *domain.Subscription
	if err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		s, err := r.Subscription().GetByID(ctx, sub.ID(), true)
		if err != nil || !s.IsCurrent() {
			return err
		}
		if _, err := s.CancelNow(now, reason, quote, actor); err != nil {
			return err
		}
		saved = s
		return r.Subscription().Save(ctx, s)
	}); err != nil || saved == nil {
		return err
	}
	notifySubscription(ctx, c.notifier, saved)
	return c.refunds.Run(ctx, saved.ID())
}
```

Lifetime `reason`: callers pass `RefundReasonLifetimeWindow` when the subscription is lifetime and the reason is a user cancel (see Step 2).

- [x] **Step 2: Command, query, HTTP**

```go
// internal/billing/app/cmd_cancelnow.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CancelNow struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type CancelNowHandler struct {
	authz     AuthorizationSvc
	canceller *Canceller
}

func NewCancelNowHandler(authz AuthorizationSvc, canceller *Canceller) *CancelNowHandler {
	return &CancelNowHandler{authz: authz, canceller: canceller}
}

var ProvideCancelNowHandler = NewCancelNowHandler

type CancelNowCmd commonhandler.Cmd[CancelNow]

var _ CancelNowCmd = (*CancelNowHandler)(nil)

func (h *CancelNowHandler) Handle(ctx context.Context, cmd *CancelNow) error {
	if err := requirePermission(ctx, h.authz, cmd.UserID, cmd.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return err
	}
	sub, _, err := h.canceller.load(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	reason := domain.RefundReasonCancelNow
	if sub.Interval() == domain.IntervalLifetime {
		reason = domain.RefundReasonLifetimeWindow
	}
	return h.canceller.CancelNow(ctx, cmd.WorkspaceID, reason, cmd.UserID, false)
}
```

```go
// internal/billing/app/q_previewcancelnow.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PreviewCancelNow struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type PreviewCancelNowHandler struct {
	authz     AuthorizationSvc
	canceller *Canceller
}

func NewPreviewCancelNowHandler(authz AuthorizationSvc, canceller *Canceller) *PreviewCancelNowHandler {
	return &PreviewCancelNowHandler{authz: authz, canceller: canceller}
}

var ProvidePreviewCancelNowHandler = NewPreviewCancelNowHandler

type PreviewCancelNowQuery commonhandler.Query[PreviewCancelNow, RefundQuote]

var _ PreviewCancelNowQuery = (*PreviewCancelNowHandler)(nil)

func (h *PreviewCancelNowHandler) Handle(ctx context.Context, q *PreviewCancelNow) (RefundQuote, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return RefundQuote{}, err
	}
	quote, err := h.canceller.Quote(ctx, q.WorkspaceID)
	if err != nil {
		return RefundQuote{}, err
	}
	return newRefundQuote(quote), nil
}
```

```go
// internal/billing/controller/http/cancel.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) PreviewCancelNow(ctx context.Context, req billing.PreviewCancelNowRequestObject) (billing.PreviewCancelNowResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	q, err := h.App.Queries.PreviewCancelNow.Handle(ctx, &app.PreviewCancelNow{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	payments := make([]billing.RefundQuotePayment, 0, len(q.Lines))
	for _, l := range q.Lines {
		payments = append(payments, billing.RefundQuotePayment{
			PaidAmount: l.PaidAmount, Amount: l.Amount, CoverageSeconds: l.CoverageSeconds, UnusedSeconds: l.UnusedSeconds,
		})
	}
	return billing.PreviewCancelNow200JSONResponse{
		Amount: q.Amount.Amount, Currency: q.Amount.Currency, UnusedSeconds: q.UnusedSeconds, PeriodSeconds: q.PeriodSeconds,
		Payments: payments,
	}, nil
}

func (h *StrictHandler) CancelNow(ctx context.Context, req billing.CancelNowRequestObject) (billing.CancelNowResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.App.Cmds.CancelNow.Handle(ctx, &app.CancelNow{WorkspaceID: req.WorkspaceId, UserID: u.ID}); err != nil {
		return nil, err
	}
	v, err := h.App.Queries.GetWorkspaceBilling.Handle(ctx, &app.GetWorkspaceBilling{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	out, err := toWorkspaceBilling(v)
	if err != nil {
		return nil, err
	}
	return billing.CancelNow200JSONResponse(out), nil
}
```

Note `GetWorkspaceBilling` returns only the **current** subscription, which is now none: the response shows `subscription: null` and the payer. Register handlers + `ProvideCanceller` in `wire.go`; delete the two stubs.

- [x] **Step 3: Build, E2E, commit**

```bash
nx run billing:gen:wire && nx build billing && nx lint billing && go test ./internal/billing/...
```

E2E: on the `plus` workspace from Task 8, `GET .../subscription/cancel-now/preview` → amount ≈ unused fraction of 49000. `POST .../subscription/cancel-now` → 200; entitlements → `free`; Stripe dashboard: subscription canceled, refund of the previewed amount succeeded; `refund.updated` webhook processed (refund row `succeeded`).

```bash
git add internal/billing cmd/billing/wire_gen.go
git commit -m "feat(billing): cancel now with prorated refund"
```

---

### Task 10: Payer claim, payer removal, workspace deletion (Kafka)

**Files:**

- Create: `internal/billing/app/{q_claimpayer.go,cmd_handlepayerremoved.go,cmd_handleworkspacedeleted.go}`
- Create: `internal/billing/controller/http/payer.go`; modify `unimplemented.go`
- Create: `internal/billing/controller/event/{event.go,wire.go}`
- Modify: `internal/billing/controller/wire.go`, `internal/billing/server.go`, `internal/billing/app/app.go`, `wire.go`

**Interfaces:**

- Consumes: Kafka topics `events.integration.note.workspace.deleted` (`share.WorkspaceDeletedIntegrationEvent`), `events.integration.authorization.user_workspace_role_updated` (`share.UserWorkspaceRoleUpdatedEvent`), `events.integration.authorization.workspace_member_removed` (`share.WorkspaceMemberRemovedEvent`).
- Produces: `Queries.ClaimPayer` (`ClaimPayer{WorkspaceID uuid.UUID; Payer app.Payer}` → checkout URL `string`), `Cmds.HandlePayerRemoved` (`HandlePayerRemoved{WorkspaceID uuid.UUID; UserID string}`), `Cmds.HandleWorkspaceDeleted` (`HandleWorkspaceDeleted{WorkspaceID uuid.UUID}`).

- [x] **Step 1: Commands**

```go
// internal/billing/app/q_claimpayer.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ClaimPayer struct {
	WorkspaceID uuid.UUID
	Payer       Payer
}

type ClaimPayerHandler struct {
	uow      domain.UnitOfWork
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewClaimPayerHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *ClaimPayerHandler {
	return &ClaimPayerHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideClaimPayerHandler = NewClaimPayerHandler

type ClaimPayerQuery commonhandler.Query[ClaimPayer, string]

var _ ClaimPayerQuery = (*ClaimPayerHandler)(nil)

func (h *ClaimPayerHandler) Handle(ctx context.Context, q *ClaimPayer) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var account *domain.BillingAccount
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, q.WorkspaceID, false)
		return err
	}); err != nil {
		return "", err
	}
	if account == nil {
		return "", errs.New(errs.CodeNoBillingAccount, "workspace has no billing account", nil)
	}
	if account.Payer().ID == q.Payer.ID && account.PayerMissingSince() == nil {
		return "", errs.New(errs.CodeAlreadyPayer, "you are already the payer", nil)
	}
	return h.provider.CreateCheckoutSession(ctx, CheckoutInput{
		Purpose: CheckoutPurposePayerChange, WorkspaceID: q.WorkspaceID, CustomerID: account.ProviderCustomerID(), Payer: toDomainPayer(q.Payer),
		SuccessURL: billingPageURL(h.settings, q.WorkspaceID, "payer"), CancelURL: billingPageURL(h.settings, q.WorkspaceID, "cancel"),
	})
}
```

```go
// internal/billing/app/cmd_handlepayerremoved.go
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type HandlePayerRemoved struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type HandlePayerRemovedHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	settings *Settings
}

func NewHandlePayerRemovedHandler(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, settings *Settings) *HandlePayerRemovedHandler {
	return &HandlePayerRemovedHandler{uow: uow, provider: provider, notifier: notifier, settings: settings}
}

var ProvideHandlePayerRemovedHandler = NewHandlePayerRemovedHandler

type HandlePayerRemovedCmd commonhandler.Cmd[HandlePayerRemoved]

var _ HandlePayerRemovedCmd = (*HandlePayerRemovedHandler)(nil)

func (h *HandlePayerRemovedHandler) Handle(ctx context.Context, cmd *HandlePayerRemoved) error {
	var account *domain.BillingAccount
	var sub *domain.Subscription
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if account, err = r.BillingAccount().GetByWorkspace(ctx, cmd.WorkspaceID, false); err != nil {
			return err
		}
		sub, err = r.Subscription().GetCurrentByWorkspace(ctx, cmd.WorkspaceID, false)
		return err
	}); err != nil {
		return err
	}
	if account == nil || account.Payer().ID != cmd.UserID || account.PayerMissingSince() != nil {
		return nil
	}
	recurring := sub != nil && sub.ProviderSubscriptionID() != nil
	if recurring {
		if err := h.provider.SetCancelAtPeriodEnd(ctx, *sub.ProviderSubscriptionID(), true); err != nil {
			return err
		}
	}
	if err := h.provider.DetachPaymentMethods(ctx, account.ProviderCustomerID(), ""); err != nil {
		return err
	}
	now, actor := time.Now(), h.settings.ServiceAccountUserID
	err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		a, err := r.BillingAccount().GetByWorkspace(ctx, cmd.WorkspaceID, true)
		if err != nil {
			return err
		}
		a.MarkPayerMissing(now, actor)
		if err := r.BillingAccount().Save(ctx, a); err != nil {
			return err
		}
		if !recurring {
			return nil
		}
		s, err := r.Subscription().GetByID(ctx, sub.ID(), true)
		if err != nil {
			return err
		}
		if err := s.ScheduleCancel(now, actor); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, s)
	})
	if err == nil && sub != nil {
		notifySubscription(ctx, h.notifier, sub)
	}
	return err
}
```

```go
// internal/billing/app/cmd_handleworkspacedeleted.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type HandleWorkspaceDeleted struct {
	WorkspaceID uuid.UUID
}

type HandleWorkspaceDeletedHandler struct {
	canceller *Canceller
	settings  *Settings
}

func NewHandleWorkspaceDeletedHandler(canceller *Canceller, settings *Settings) *HandleWorkspaceDeletedHandler {
	return &HandleWorkspaceDeletedHandler{canceller: canceller, settings: settings}
}

var ProvideHandleWorkspaceDeletedHandler = NewHandleWorkspaceDeletedHandler

type HandleWorkspaceDeletedCmd commonhandler.Cmd[HandleWorkspaceDeleted]

var _ HandleWorkspaceDeletedCmd = (*HandleWorkspaceDeletedHandler)(nil)

func (h *HandleWorkspaceDeletedHandler) Handle(ctx context.Context, cmd *HandleWorkspaceDeleted) error {
	err := h.canceller.CancelNow(ctx, cmd.WorkspaceID, domain.RefundReasonWorkspaceDeleted, h.settings.ServiceAccountUserID, true)
	if hasCode(err, errs.CodeNoActiveSubscription) {
		return nil
	}
	return err
}
```

- [x] **Step 2: HTTP claim**

```go
// internal/billing/controller/http/payer.go
package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) ClaimPayer(ctx context.Context, req billing.ClaimPayerRequestObject) (billing.ClaimPayerResponseObject, error) {
	payer, err := currentPayer(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.ClaimPayer.Handle(ctx, &app.ClaimPayer{WorkspaceID: req.WorkspaceId, Payer: payer})
	if err != nil {
		return nil, err
	}
	return billing.ClaimPayer200JSONResponse{Url: url}, nil
}
```

`unimplemented.go` should now be empty: delete the file.

- [x] **Step 3: Kafka consumer controller**

```go
// internal/billing/controller/event/event.go
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
)

type Event struct {
	router *message.Router
	app    *app.Server
}

func New(kafkaCfg *commonconfig.Kafka, a *app.Server, tracer kafka.SaramaTracer, logger watermill.LoggerAdapter, _ otel.Global) (*Event, error) {
	subscriber, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers: kafkaCfg.Brokers, ConsumerGroup: kafkaCfg.ConsumerGroup, Tracer: tracer,
		OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}
	retry := middleware.Retry{MaxRetries: 5, InitialInterval: time.Second, MaxInterval: 30 * time.Second, Multiplier: 2, Logger: logger}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer, retry.Middleware, wotel.Trace())

	e := &Event{router: router, app: a}
	router.AddConsumerHandler("WorkspaceDeleted", component.TopicWorkspaceDeleted, subscriber, e.workspaceDeleted)
	router.AddConsumerHandler("RoleUpdated", component.TopicRoleUpdated, subscriber, e.roleUpdated)
	router.AddConsumerHandler("MemberRemoved", component.TopicMemberRemoved, subscriber, e.memberRemoved)
	return e, nil
}

var Provide = New

func (e *Event) workspaceDeleted(msg *message.Message) error {
	var ev share.WorkspaceDeletedIntegrationEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("decode workspace deleted: %w", err)
	}
	return e.app.Cmds.HandleWorkspaceDeleted.Handle(msg.Context(), &app.HandleWorkspaceDeleted{WorkspaceID: ev.Id})
}

func (e *Event) roleUpdated(msg *message.Message) error {
	var ev share.UserWorkspaceRoleUpdatedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("decode role updated: %w", err)
	}
	if ev.Role == share.Owner || ev.WorkspaceId == nil {
		return nil
	}
	return e.app.Cmds.HandlePayerRemoved.Handle(msg.Context(), &app.HandlePayerRemoved{WorkspaceID: *ev.WorkspaceId, UserID: ev.UserId})
}

func (e *Event) memberRemoved(msg *message.Message) error {
	var ev share.WorkspaceMemberRemovedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("decode member removed: %w", err)
	}
	if ev.WorkspaceId == nil {
		return nil
	}
	return e.app.Cmds.HandlePayerRemoved.Handle(msg.Context(), &app.HandlePayerRemoved{WorkspaceID: *ev.WorkspaceId, UserID: ev.UserId})
}

func (e *Event) Run(ctx context.Context) error { return e.router.Run(ctx) }

func (e *Event) IsRunning() bool { return e.router.IsRunning() }
```

Type conversions: `share.Id`, `share.Id2` are aliases of `openapi_types.UUID` (= `uuid.UUID`) / `string`; convert as needed (`string(ev.UserId)`, `uuid.UUID(*ev.WorkspaceId)`). Each handler gets its own consumer group because Watermill's Kafka subscriber shares the group across topics in one subscriber — if a topic is starved, create one subscriber per handler with `ConsumerGroup + "." + name` as note's event processor does.

`controller/event/wire.go`: `ProviderSet = wire.NewSet(Provide)`; add to `controller.ProviderSet`. `server.go`: add `event *event.Event` and a `g.Go(func() error { return s.event.Run(ctx) })`.

- [x] **Step 4: Build, E2E, commit**

```bash
nx run billing:gen:wire && nx build billing && nx lint billing
```

E2E (workspace on `plus`, payer `kevinnitro`, second owner `ntgnguyen`):

1. Demote `kevinnitro` to editor through note's members API (plan 04 adds the guard; before plan 04 it is allowed) → billing logs `PayerMissing`; Stripe subscription shows "cancels at period end"; customer has no payment methods.
2. As `ntgnguyen`: `POST /billing/workspaces/W/payer/claim` → complete setup checkout with `4242…` → `GET .../subscription` shows payer `ntgnguyen`, `missingSince: null`, `cancelAtPeriodEnd: false`.
3. Delete the workspace (note API) → subscription canceled, prorated refund in Stripe.

```bash
git add internal/billing cmd/billing/wire_gen.go
git commit -m "feat(billing): payer claim and kafka handlers for payer removal and workspace deletion"
```

---

### Task 11: Crons — reconcile, refund retry, event pruning

**Files:**

- Modify: `internal/billing/domain/repo.go` (add `ListIDsWithUnsubmittedRefunds`), `internal/billing/infra/persistence/pgsqlc/subscription.sql`, `internal/billing/infra/persistence/pgrepo/subscription.go`, `internal/billing/app/memstore_test.go`
- Create: `internal/billing/app/{cmd_reconcilesubscriptions.go,cmd_pruneproviderevents.go}`
- Create: `internal/billing/controller/cron/{cron.go,wire.go}`
- Modify: `internal/billing/controller/wire.go`, `internal/billing/server.go`, `app.go`, `wire.go`

**Interfaces:**

- Consumes: `pkg/cron` (`cron.Job`, `cron.NewScheduler(locker, logger, location)`, `cron.NewRedisLocker`) from plan 02. Location is `time.Local` (zone from `TZ` env or system), never hard-coded.
- Produces: jobs `billing:reconcile-subscriptions` (hourly) and `billing:prune-provider-events` (daily 04:00).

- [ ] **Step 1: Repo addition**

`domain/repo.go` `SubscriptionRepo`: `ListIDsWithUnsubmittedRefunds(ctx context.Context, limit int) ([]uuid.UUID, error)`.

`subscription.sql`:

```sql
-- name: ListSubscriptionIDsWithUnsubmittedRefunds :many
SELECT DISTINCT subscription_id FROM refunds
WHERE status = 'pending' AND provider_refund_id IS NULL
LIMIT @row_limit;
```

`pgrepo/subscription.go`:

```go
func (r *subscriptionRepo) ListIDsWithUnsubmittedRefunds(ctx context.Context, limit int) ([]uuid.UUID, error) {
	ids, err := r.q.ListSubscriptionIDsWithUnsubmittedRefunds(ctx, int32(limit))
	return ids, wrap("list subscriptions with unsubmitted refunds", err)
}
```

`memstore_test.go` `memSubs`: `func (r memSubs) ListIDsWithUnsubmittedRefunds(context.Context, int) ([]uuid.UUID, error) { return nil, nil }`. Run `nx run billing:gen:sqlc`.

- [ ] **Step 2: Commands**

```go
// internal/billing/app/cmd_reconcilesubscriptions.go
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ReconcileSubscriptions struct{}

type ReconcileSubscriptionsHandler struct {
	uow     domain.UnitOfWork
	sync    *SyncSubscriptionHandler
	refunds *RefundRunner
}

func NewReconcileSubscriptionsHandler(uow domain.UnitOfWork, sync *SyncSubscriptionHandler, refunds *RefundRunner) *ReconcileSubscriptionsHandler {
	return &ReconcileSubscriptionsHandler{uow: uow, sync: sync, refunds: refunds}
}

var ProvideReconcileSubscriptionsHandler = NewReconcileSubscriptionsHandler

type ReconcileSubscriptionsCmd commonhandler.Cmd[ReconcileSubscriptions]

var _ ReconcileSubscriptionsCmd = (*ReconcileSubscriptionsHandler)(nil)

func (h *ReconcileSubscriptionsHandler) Handle(ctx context.Context, _ *ReconcileSubscriptions) error {
	var stale []*domain.Subscription
	var refundSubs []uuid.UUID
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if stale, err = r.Subscription().ListNeedingReconcile(ctx, time.Now(), 100); err != nil {
			return err
		}
		refundSubs, err = r.Subscription().ListIDsWithUnsubmittedRefunds(ctx, 100)
		return err
	}); err != nil {
		return err
	}
	for _, s := range stale {
		if err := h.sync.Handle(ctx, &SyncSubscription{ProviderSubscriptionID: *s.ProviderSubscriptionID()}); err != nil {
			slog.WarnContext(ctx, "reconcile failed", slog.String("subscription_id", s.ID().String()), slog.Any("error", err))
		}
	}
	for _, id := range refundSubs {
		if err := h.refunds.Run(ctx, id); err != nil {
			slog.WarnContext(ctx, "refund retry failed", slog.String("subscription_id", id.String()), slog.Any("error", err))
		}
	}
	return nil
}
```

```go
// internal/billing/app/cmd_pruneproviderevents.go
package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PruneProviderEvents struct{}

type PruneProviderEventsHandler struct{ uow domain.UnitOfWork }

func NewPruneProviderEventsHandler(uow domain.UnitOfWork) *PruneProviderEventsHandler {
	return &PruneProviderEventsHandler{uow: uow}
}

var ProvidePruneProviderEventsHandler = NewPruneProviderEventsHandler

type PruneProviderEventsCmd commonhandler.Cmd[PruneProviderEvents]

var _ PruneProviderEventsCmd = (*PruneProviderEventsHandler)(nil)

func (h *PruneProviderEventsHandler) Handle(ctx context.Context, _ *PruneProviderEvents) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		_, err := r.ProviderEventLog().PruneBefore(ctx, time.Now().Add(-30*24*time.Hour))
		return err
	})
}
```

- [ ] **Step 3: Cron controller**

```go
// internal/billing/controller/cron/cron.go
package cron

import (
	"context"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/notopia-uit/notopia/internal/billing/app"
	pkgcron "github.com/notopia-uit/notopia/pkg/cron"
)

// NOTE: move this controller to cmd/billing-worker when crons are split out of the API process.
type Cron struct {
	scheduler *pkgcron.Scheduler
}

func New(a *app.Server, redis *goredis.Client, logger *slog.Logger) (*Cron, error) {
	locker, err := pkgcron.NewRedisLocker(redis)
	if err != nil {
		return nil, err
	}
	s, err := pkgcron.NewScheduler(locker, logger, time.Local)
	if err != nil {
		return nil, err
	}
	if err := s.Register(
		pkgcron.Job{
			Name: "billing:reconcile-subscriptions", Schedule: "0 * * * *", Timeout: 15 * time.Minute,
			Run: func(ctx context.Context) error {
				return a.Cmds.ReconcileSubscriptions.Handle(ctx, &app.ReconcileSubscriptions{})
			},
		},
		pkgcron.Job{
			Name: "billing:prune-provider-events", Schedule: "0 4 * * *", Timeout: 5 * time.Minute,
			Run: func(ctx context.Context) error {
				return a.Cmds.PruneProviderEvents.Handle(ctx, &app.PruneProviderEvents{})
			},
		},
	); err != nil {
		return nil, err
	}
	return &Cron{scheduler: s}, nil
}

var Provide = New

func (c *Cron) Run(ctx context.Context) error {
	c.scheduler.Start()
	<-ctx.Done()
	return c.scheduler.Shutdown()
}
```

`controller/cron/wire.go`: `ProviderSet = wire.NewSet(Provide)`; add to `controller.ProviderSet`. `server.go`: add `cron *cron.Cron`, `g.Go(func() error { return s.cron.Run(ctx) })`. Register the two commands in `app.go`/`wire.go`.

- [ ] **Step 4: Verify, commit**

```bash
nx run billing:gen:sqlc && nx run billing:gen:wire && nx build billing && nx lint billing && go test ./internal/billing/...
```

Manual: temporarily change the reconcile schedule to `* * * * *`, run two billing instances (second with `NOTOPIA_BILLING_SERVER_HTTP_PORT=8093 NOTOPIA_BILLING_SERVER_GRPC_PORT=18093 NOTOPIA_BILLING_SERVER_HEALTH_PORT=28093`), and confirm in logs that each tick's `cron job done billing:reconcile-subscriptions` appears on only one instance. Revert the schedule.

```bash
git add internal/billing cmd/billing/wire_gen.go
git commit -m "feat(billing): reconcile, refund retry and event pruning crons"
```

---

### Task 12: Docs and full sandbox E2E

**Files:**

- Modify: `internal/billing/AGENTS.md`, root `CLAUDE.md`
- Create: `docs/src/<billing page>.md` (follow the existing VitePress layout in `docs/src`): context diagram, webhook sequence, subscription state diagram (mermaid)

- [ ] **Step 1: Docs**

`internal/billing/AGENTS.md` final content: layers; ordering rule for webhook use cases; notifications are best-effort; `just seed-billing`, `just fixtures-billing`, `just stripe-listen`; test cards; TODO (free trial, live payments via VN gateway, more currencies).

Docs page diagrams:

```mermaid
stateDiagram-v2
  [*] --> incomplete: checkout created subscription
  incomplete --> active: first invoice paid
  incomplete --> incomplete_expired: not paid in 23h
  active --> past_due: renewal payment failed
  past_due --> active: retry succeeded
  past_due --> unpaid: retries exhausted
  active --> canceled: period end after cancel_at_period_end / cancel-now
  past_due --> canceled
  unpaid --> canceled
```

```mermaid
sequenceDiagram
  participant Stripe
  participant Billing
  participant DB
  participant Kafka
  Stripe->>Billing: POST /billing/webhooks/stripe (signed)
  Billing->>Billing: verify signature (raw body)
  Billing->>Stripe: GET subscription / checkout session (fetch-and-sync)
  Billing->>DB: tx: MarkProcessed(event.id) + apply state
  Billing->>Stripe: refunds (idempotency key refund:{id})
  Billing-->>Kafka: subscription.changed (best-effort)
  Billing-->>Stripe: 204
```

- [ ] **Step 2: Full E2E checklist (sandbox, `just stripe-listen` running)**

1. Free workspace → `GET entitlements` = free.
2. Checkout `plus_month` (`4242 4242 4242 4242`) → plus.
3. Portal: upgrade to `pro_month` → pro immediately; downgrade to `plus_year`… scheduled at period end (Portal shows "Changes on …").
4. `stripe trigger invoice.payment_failed` against the subscription (or use test clock) → `past_due`, still `pro` until period end + 3 days.
5. Portal: cancel at period end → `cancelAtPeriodEnd: true`, still entitled.
6. Cancel-now → refund succeeded, free.
7. Buy `pro_lifetime` → pro forever; cancel-now within 14 days → full refund.
8. Two browser tabs, two checkouts → second auto-cancelled and refunded.
9. Payer demoted → payer missing; claim by another owner restores.
10. Workspace deleted → canceled + refunded.
11. Resend any webhook from the dashboard → no state change, 204.

- [ ] **Step 3: Commit**

```bash
git add internal/billing/AGENTS.md CLAUDE.md docs/src
git commit -m "docs(billing): service guide and diagrams"
```
