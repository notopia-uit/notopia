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
  past_due_since TIMESTAMPTZ,
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
