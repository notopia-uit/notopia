# Billing

:::info

- Stripe runs in **sandbox** only: Stripe doesn't onboard Vietnamese businesses for live payouts. See `story/BILLING_PAYMENT_PROVIDERS.md`.
- Webhooks are **fetch-and-sync**: the payload only gives an id, and billing reads the current state back from Stripe, so event order doesn't matter.
- Entitlements are always read from billing's own database. Other services call gRPC on every check and never cache.

:::

## Subscription States

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

`past_due` keeps paid features until period end + `pastDueGraceDays` (billing policy, default 3). Lifetime purchases have no Stripe subscription: they're created `active` and only leave it through cancel-now (refund window) or workspace deletion.

## Checkout

```mermaid
sequenceDiagram
    autonumber

    actor U as Owner
    participant B as Billing
    participant AS as Authorization Service
    participant S as Stripe

    U->>+B: POST /workspaces/{id}/checkout-session {priceId}
    B->>+AS: HasWorkspacePermission(billing_manage)
    AS-->>-B: Ok
    opt No billing account yet
        B->>S: Create customer (idempotency key customer:{workspaceId})
        B->>B: Save BillingAccount
    end
    B->>S: Create Checkout Session (metadata: workspace, price, purpose)
    B-->>-U: { url }
    U->>S: Pay on checkout.stripe.com
    S-->>U: Redirect to /settings/billing?checkout=success
```

## Webhook Pipeline

```mermaid
sequenceDiagram
    autonumber

    participant S as Stripe
    participant B as Billing
    participant DB as Billing DB
    participant K as Kafka

    S->>+B: POST /billing/webhooks/stripe (signed)
    B->>B: Verify signature on the raw body
    B->>S: GET checkout session / subscription / refund (fetch-and-sync)
    opt Duplicate or replaced subscription
        B->>S: Cancel provider subscription (before the transaction, idempotent)
    end
    B->>DB: tx: MarkProcessed(event.id) + apply state
    Note over B,DB: An event id already processed → no-op
    opt Refund due
        B->>S: Create refund (idempotency key refund:{id})
        B->>DB: Save provider refund id + status
    end
    B--)K: subscription.changed (best-effort, after commit)
    B-->>-S: 204
```

Ordering rule for every webhook use case:

1. Provider mutations that must not be lost (cancel a duplicate subscription, change the default card) run **before** the write transaction and are idempotent.
2. The write transaction records the event (`processed_provider_events`) together with the state change.
3. Refunds run **after** commit. A refund that fails to submit stays `pending` without a provider id, and the hourly reconcile cron retries it.

## Cancel Now

```mermaid
sequenceDiagram
    autonumber

    actor U as Owner
    participant B as Billing
    participant S as Stripe
    participant DB as Billing DB

    U->>+B: GET .../subscription/cancel-now/preview
    B->>S: Period payments (recurring) / payment intent (lifetime)
    B-->>-U: Refund quote (unused share of the period)
    U->>+B: POST .../subscription/cancel-now
    B->>S: Cancel subscription now
    B->>DB: tx: Subscription canceled + pending Refund
    B->>S: Create refund (refund:{id})
    B->>DB: Refund submitted
    B-->>-U: Workspace billing (subscription: null)
    S->>B: refund.updated webhook → refund succeeded
```

## Payer Removed and Claimed

```mermaid
sequenceDiagram
    autonumber

    participant AS as Authorization Service
    participant K as Kafka
    participant B as Billing
    participant S as Stripe
    actor O as Another Owner

    AS--)K: user_workspace_role_updated / workspace_member_removed
    K--)B: Payer lost the owner role
    B->>S: Cancel at period end + detach payment methods
    B->>B: BillingAccount payer missing
    O->>+B: POST .../payer/claim
    B-->>-O: Setup-mode checkout URL
    O->>S: Save a card
    S->>B: checkout.session.completed (payer_change)
    B->>S: Set default payment method, resume subscription
    B->>B: Payer = new owner
```

## Crons

| Job                               | Schedule                   | Effect                                                                                           |
| --------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------ |
| `billing:reconcile-subscriptions` | hourly                     | Re-syncs subscriptions past their period end or not synced for 24 h; retries unsubmitted refunds |
| `billing:prune-provider-events`   | daily 04:00 (process `TZ`) | Deletes processed webhook event ids older than 30 days                                           |

Both use a Redis lock (`billing_redis`), so with several billing instances each tick runs once.
