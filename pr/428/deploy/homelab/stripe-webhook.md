---
url: /notopia/pr/428/deploy/homelab/stripe-webhook.md
---
# Stripe webhook

For the homelab Stripe **sandbox**, create a webhook endpoint with:

* **Scope:** Your account. This is a routing scope, not a request to use Accounts v2 events; Notopia does not use Stripe Connect.
* **Payload:** Snapshot.
* **API version:** `2026-10-28.endive`, matching `stripe-go/v87`. If only `2026-11-18.endive` is available, test it; it is in the same Endive release family.
* **URL:** `https://notopia-api.<PUBLIC_DOMAIN>/billing/webhooks/stripe`.

Select these events:

* `checkout.session.completed`
* `customer.subscription.created`, `customer.subscription.updated`, `customer.subscription.deleted`
* `invoice.paid`, `invoice.payment_failed`
* `refund.created`, `refund.updated`, `refund.failed`

The authoritative event list and handling are in [`internal/billing/controller/http/webhook.go`](https://github.com/notopia-uit/notopia/blob/main/internal/billing/controller/http/webhook.go). Check its event switch when updating the destination.

Store the destination's `whsec_...` signing secret as `NOTOPIA_BILLING_STRIPE_WEBHOOK_SECRET` and the sandbox API key as `NOTOPIA_BILLING_STRIPE_SECRET_KEY`, following the [homelab provisioning instructions](https://github.com/notopia-uit/notopia/blob/main/deploy/k8s/clusters/homelab/README.md).
