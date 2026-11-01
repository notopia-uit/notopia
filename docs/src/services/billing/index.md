---
title: Billing
order: 1
---

# Billing Service

Plans, prices, subscriptions and entitlements. Stripe (sandbox) through hosted Checkout and
Customer Portal. DDD + CQRS like the Note service, with a Watermill transactional outbox:
domain writes and event publishes commit together. Other services read entitlements
over gRPC and never cache them.

- [Domain](./domain): class diagrams
- [Data](./data): database diagram
- [Flows](./flows): checkout, webhooks, cancellation, crons

## Architecture

```d2
direction: right

vars: {
  d2-config: {
    layout-engine: elk
    theme-overrides: {
      N1: "#4c4f69"
      N2: "#5c5f77"
      N4: "#acb0be"
      N5: "#ccd0da"
      N7: "#eff1f5"
      B1: "#4c4f69"
      B2: "#6c6f85"
      B3: "#bcc0cc"
      B4: "#ccd0da"
      B5: "#dce0e8"
      B6: "#eff1f5"
      AA4: "#1e66f5"
      AA5: "#7287fd"
      AB4: "#8839ef"
      AB5: "#dc8a78"
    }
  }
}

controllers: Controllers {
  http: HTTP (gin strict) {
    api: Workspace + admin APIs
    webhook: Stripe webhook
  }
  grpc: gRPC (entitlements)
  event: Kafka consumers
  cron: Crons (Redis-locked)
}

application: Application {
  cmds: Commands
  queries: Queries
}

domain: Domain {
  subscription: Subscription + Refund
  account: BillingAccount
  catalog: Price / Plan / BillingPolicy
}

infrastructure: Infrastructure {
  persistence: Postgres (sqlc)
  stripe: Stripe adapter
  outbox: Outbox (Watermill, outbox_messages)
  authorization: Authorization gRPC client
  redis: Redis (cron locks)
}

stripe_cloud: Stripe {shape: cloud}
kafka: Redpanda {shape: queue}
authz: Authorization Service
note: Note Service

controllers -> application
application -> domain
application -> infrastructure
infrastructure.stripe -> stripe_cloud
stripe_cloud -> controllers.http.webhook: webhook events
infrastructure.persistence -> infrastructure.outbox: events in the same tx
infrastructure.outbox -> kafka: forwarder
kafka -> controllers.event: workspace.deleted\nrole_updated\nmember_removed\nsubscription.canceled_now\nsubscription.refund_requested
infrastructure.authorization -> authz
note -> controllers.grpc: entitlement checks
```

<!-- diagram id="architecture-billing-diagram" -->

## Events

Repo `Save` writes the aggregate's popped domain events into `outbox_messages` in the same
transaction. `outbox.Forwarder` relays them to Kafka, keyed by the aggregate id (workspace id
for the integration event). Topics are declared once in `internal/billing/component/eventtopic.go`.

### Published

| Topic                                             | Payload                       | When                                                                                                                                                          |
| ------------------------------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `events.integration.billing.subscription.changed` | `{workspaceId, plan, status}` | `SubscriptionStateChangedEvent`: recorded whenever a subscription's plan or status changes (not for duplicate checkouts)                                      |
| `events.internal.billing.{aggregate}.{event}`     | the domain event              | every other domain event, e.g. `subscription.canceled_now`, `subscription.refund_requested`, `billing_account.payer_missing`, `price.created`, `plan.updated` |

The `events.internal.*` topics are billing's own: only billing consumes them (see below).

### Consumed

| Topic                                                          | Use case                                              |
| -------------------------------------------------------------- | ----------------------------------------------------- |
| `events.integration.note.workspace.deleted`                    | `HandleWorkspaceDeleted`                              |
| `events.integration.authorization.user_workspace_role_updated` | `HandlePayerRemoved` (payer no longer owner)          |
| `events.integration.authorization.workspace_member_removed`    | `HandlePayerRemoved`                                  |
| `events.internal.billing.subscription.canceled_now`            | `CancelProviderSubscription` (cancel at the provider) |
| `events.internal.billing.subscription.refund_requested`        | `SubmitRefund` (create the provider refund)           |

One consumer group per topic. An undecodable payload or a permanent error (`notFound`, `invalid`) is logged and acked; other errors are retried.
