---
order: 4
---

# Billing Database Diagram

:::info

- Only column marked with `N` mean nullable
- The Watermill outbox table (`outbox_messages`) is omitted
- `workspace_id` is a logical reference to the Note service workspace (no cross-service FK)

:::

```d2
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

plan_tier: {
  shape: rectangle
  free
  plus
  pro
}

billing_interval: {
  shape: rectangle
  month
  year
  lifetime
}

subscription_status: {
  shape: rectangle
  incomplete
  active
  past_due
  unpaid
  canceled
  incomplete_expired
}

refund_reason: {
  shape: rectangle
  cancel_now
  lifetime_window
  replaced_by_lifetime
  workspace_deleted
  duplicate_checkout
  external
}

refund_status: {
  shape: rectangle
  pending
  succeeded
  failed
  canceled
}

plans: {
  tier: plan_tier {constraint: PK}
  features: jsonb
  limits: jsonb
  updated_at: timestamptz

  shape: sql_table
}

billing_policy: {
  id: smallint {constraint: PK}
  past_due_grace_days: integer
  lifetime_refund_window_days: integer
  portal_configuration_id: text {constraint: N}
  updated_at: timestamptz

  shape: sql_table
}

prices: {
  id: uuid {constraint: PK}
  plan: plan_tier
  billing_interval: billing_interval
  amount: bigint
  currency: text
  provider_price_id: text {constraint: U}
  lookup_key: text
  active: boolean
  created_at: timestamptz
  archived_at: timestamptz {constraint: N}

  shape: sql_table
}

billing_accounts: {
  workspace_id: uuid {constraint: PK}
  provider_customer_id: text {constraint: U}
  payer_id: text
  payer_email: text
  payer_name: text
  payer_missing_since: timestamptz {constraint: N}
  created_at: timestamptz
  updated_at: timestamptz

  shape: sql_table
}

subscriptions: {
  id: uuid {constraint: PK}
  workspace_id: uuid {constraint: FK}
  plan: plan_tier
  billing_interval: billing_interval
  price_id: uuid {constraint: FK}
  status: subscription_status
  provider_subscription_id: text {constraint: U, N}
  provider_payment_id: text {constraint: U, N}
  current_period_start: timestamptz {constraint: N}
  current_period_end: timestamptz {constraint: N}
  cancel_at_period_end: boolean
  canceled_at: timestamptz {constraint: N}
  paid_at: timestamptz {constraint: N}
  provider_synced_at: timestamptz
  created_at: timestamptz
  updated_at: timestamptz
  past_due_since: timestamptz {constraint: N}

  shape: sql_table
}

refunds: {
  id: uuid {constraint: PK}
  subscription_id: uuid {constraint: FK}
  provider_refund_id: text {constraint: U, N}
  payment_ref: text
  amount: bigint
  currency: text
  reason: refund_reason
  status: refund_status
  failure_message: text {constraint: N}
  created_at: timestamptz
  updated_at: timestamptz

  shape: sql_table
}

processed_provider_events: {
  event_id: text {constraint: PK}
  event_type: text
  received_at: timestamptz

  shape: sql_table
}

subscriptions.workspace_id -> billing_accounts.workspace_id
subscriptions.price_id -> prices.id
refunds.subscription_id -> subscriptions.id
```

<!-- diagram id="database-diagram-billing" -->
