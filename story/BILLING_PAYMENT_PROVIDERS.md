# Taking Money in Vietnam: Stripe, PayPal, SePay, and a Provider-Agnostic Billing Core

_Why Notopia bills through a Stripe sandbox it can never take live, why PayPal "works in Vietnam" but not in VND, and how the billing service is shaped so the payment provider can be swapped later._

Written with AI during billing-03 (October 2026). Provider facts were checked against their docs at that time. Re-check them before acting on this.

---

## The Situation

Notopia is a Vietnamese team selling a SaaS subscription priced in **VND** (Plus 49,000 ₫/month, Pro 99,000 ₫/month, Pro lifetime 2,490,000 ₫). We want what every SaaS wants:

- hosted checkout, so we never touch card numbers;
- **automatic renewals** (monthly / yearly);
- plan changes and card updates without building our own UI;
- refunds through an API (cancel now, duplicate checkout, lifetime upgrade);
- webhooks so our database follows what happened.

Stripe does all of this out of the box. Here's the catch.

---

## Stripe: Perfect API, Wrong Country

Stripe doesn't onboard businesses registered in Vietnam. An account can be created and the **sandbox (test mode) works fully**, but activating live payments requires a business entity and bank account in a supported country (Singapore and the US are the usual workarounds, with real legal and tax overhead).

VND itself isn't the problem: Stripe supports VND as a presentment currency, as a **zero-decimal currency** (`49000` means 49,000 ₫, not 490.00). Only the merchant's country blocks us.

**Decision: build on Stripe sandbox now.** It's the reference implementation of the model we want, it lets us demo the whole flow end to end, and the billing core is written against our own port, not against Stripe.

---

## PayPal: "Supports Vietnam" ≠ "Supports VND"

Both of these are true at once, which is the confusing part:

1. **PayPal supports Vietnam as a country.** A Vietnamese person or business can open a PayPal account, send money, and receive payments.
2. **PayPal does not support VND as a currency.** The [PayPal currency code list](https://developer.paypal.com/reference/currency-codes/) has 24 currencies (USD, EUR, JPY, SGD, THB, PHP, MYR, …) and VND isn't one of them.

"Country supported" means _who_ can hold an account. "Currency supported" means _what an order, subscription plan, or balance can be denominated in_. A Vietnamese merchant on PayPal therefore:

- prices and charges in USD (or another listed currency);
- holds a USD balance in PayPal;
- withdraws to a Vietnamese bank account, converted to VND, paying a conversion spread (commonly reported around 3–4%) plus a withdrawal fee. Users report Vietnamese banks only accepting the payout into a VND account.

Why no VND? PayPal hasn't published a reason. The usual factors for a currency being left out are foreign-exchange regulation (the dong is a managed, not freely convertible, currency, and cross-border flows are controlled by the State Bank of Vietnam), low transaction volume for PayPal in that currency, and settlement partnerships. Treat that as context, not as PayPal's stated reason.

**What this means for us:** PayPal is a real option for _international_ customers paying in USD, not for Vietnamese customers expecting ₫ prices. Its API also lacks a few things Stripe gives us:

| Need                                     | PayPal                                                                                                     |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| Recurring billing                        | ✅ Products + Billing Plans + Subscriptions                                                                |
| Refund API                               | ✅ Refund a capture, with an idempotency request id                                                        |
| Webhooks                                 | ✅ Verified with the verify-webhook-signature API                                                          |
| Customer object                          | ❌ The subscriber is a PayPal account, not a customer we own                                               |
| Hosted portal (change plan, update card) | ❌ We'd build the plan change ourselves with "revise subscription"                                         |
| Change the paying card/person            | ❌ The card belongs to the subscriber's PayPal account. Changing payer means cancelling and re-subscribing |
| VND                                      | ❌                                                                                                         |

Go SDK: [`github.com/plutov/paypal/v4`](https://github.com/plutov/paypal) (community, maintained, covers subscriptions, plans, refunds and webhook verification).

---

## SePay: Built for VND, Built Differently

[SePay](https://docs.sepay.vn/) is Vietnamese and VND-native, but it's a different _kind_ of product. It has two layers:

### 1. Bank-transfer reconciliation (the original SePay)

SePay watches **your own bank account** and calls your webhook whenever money arrives. Customers pay by scanning a **VietQR** code in their banking app. The QR encodes the amount and a payment code (e.g. `SEVN63DC8E5C`) in the transfer description, and you match that code to an order.

- Money lands directly in our bank account. There's no intermediate wallet and no card gateway fee.
- Webhook payload: `id`, `gateway`, `transactionDate`, `accountNumber`, `code`, `content`, `transferType` (`in`/`out`), `transferAmount`, `referenceCode`, …
- Webhook auth: API key (`Authorization: Apikey …`), HMAC-SHA256, or OAuth2. The receiver must answer `200/201` with `{"success": true}` within 30 s. Failed deliveries retry 7 times on a Fibonacci backoff, for up to 5 hours.
- Deduplicate on `id`, the same idea as our `processed_provider_events`.

### 2. SePay Payment Gateway (newer)

[developer.sepay.vn](https://developer.sepay.vn/en/cong-thanh-toan/gioi-thieu) adds an order-based gateway: VietQR, NAPAS QR, and **international cards (Visa / Mastercard / JCB with 3-D Secure, through VPBank + Mastercard MPGS)**. You create an order, get a payment link or QR code, the customer pays, SePay sends an **IPN**, and the browser comes back to your `return_url`. There's a sandbox (`pgapi-sandbox.sepay.vn`), Basic Auth with `merchant_id:secret_key`, and Node.js / PHP SDKs (no official Go SDK, but the REST API is small).

### What SePay doesn't have (not found in the docs, October 2026)

- **No recurring billing / saved-card auto-charge.** Every payment is a customer action.
- **No refund API.** A bank transfer refund is another bank transfer, made by a person.
- **No hosted customer portal.**

So SePay is **pull** (the customer pays each time we ask), whereas Stripe and PayPal are **push** (the provider charges a saved method on schedule). That's the important difference for our design, more than the API shapes.

---

## Side by Side

|                                  | Stripe             | PayPal              | SePay                                  |
| -------------------------------- | ------------------ | ------------------- | -------------------------------------- |
| Live merchant in Vietnam         | ❌                 | ✅                  | ✅                                     |
| Charge in VND                    | ✅ (zero-decimal)  | ❌                  | ✅                                     |
| Automatic renewals               | ✅                 | ✅                  | ❌ (we send a new QR/link each period) |
| Refund API                       | ✅                 | ✅                  | ❌ (manual bank transfer)              |
| Hosted plan change / card update | ✅ Customer Portal | ❌                  | ❌                                     |
| Webhooks + dedup id              | ✅ signed          | ✅ verified via API | ✅ API key / HMAC, `id`                |
| Sandbox                          | ✅                 | ✅                  | ✅ (gateway) / transaction simulator   |
| Go SDK                           | ✅ official        | community           | none (REST)                            |

---

## How the Billing Service Is Shaped Today

The domain and use cases already don't care who the provider is:

```
controller/http/webhook.go  ──► Cmds.CompleteCheckout{SessionID}       (Stripe-specific inbound adapter)
                            ──► Cmds.SyncSubscription{ProviderSubscriptionID}
                            ──► Cmds.SyncRefund{ProviderRefundID}

app/  ── PaymentProvider (port) ──► infra/payment/stripe (the only stripe-go import besides webhook.go)
domain/ Subscription, Refund, BillingAccount: provider ids are plain strings,
        state arrives as ProviderSnapshot / PaymentInfo
```

What already makes a swap cheaper:

- **Fetch-and-sync webhooks.** A webhook is only a signal carrying an id; the handler reads the current state back from the provider. That works for any provider, including SePay's "money arrived" event.
- **Idempotency is ours.** `processed_provider_events` dedups by event id, and refunds carry our own `refund:{id}` key. Every provider in the table gives us an id to dedup on.
- **Entitlements never ask the provider.** `ResolveEntitlements` reads only our database. Other services call gRPC, which never reaches Stripe.
- **Money is `{Amount int64, Currency string}`** in the smallest unit. VND has zero decimals, USD has two, so a USD provider means changing the seed prices, not the code.

Where it's still Stripe-shaped:

- `PaymentProvider` has 17 methods and assumes provider-managed renewals, a customer object, a hosted portal, saved payment methods, and a refund API. PayPal fails 3 of those and SePay fails all 5.
- The webhook controller is Stripe-only (signature check, event names). That's fine as long as each provider gets its **own inbound adapter**, all calling the same commands.

---

## The Provider-Agnostic Design (When We Need It)

**Not now.** We stay on Stripe sandbox and don't add abstractions for a provider we haven't picked. When a second provider becomes real, this is the plan.

### 1. Split the port by capability

```go
// Every provider
type PaymentProvider interface {
    CreateCheckout(ctx, CheckoutInput) (CheckoutLink, error)     // hosted page / payment link / VietQR
    GetCheckout(ctx, id string) (CheckoutSessionResult, error)
    PaymentForRef(ctx, ref string) (domain.PaymentInfo, error)
}

// Optional: provider charges a saved method on schedule (Stripe, PayPal)
type RecurringBilling interface {
    GetSubscriptionSnapshot(ctx, id string) (domain.ProviderSnapshot, error)
    SetCancelAtPeriodEnd(ctx, id string, cancel bool) error
    CancelSubscriptionNow(ctx, id string) error
    PeriodPayments(ctx, id string, start, end time.Time) ([]domain.PaymentInfo, error)
}

type Catalog interface { CreatePrice(...); ArchivePrice(...) }                // Stripe products/prices, PayPal plans; SePay: none (amount per order)
type Refunds interface { CreateRefund(...); GetRefund(...) }               // Stripe, PayPal
type CustomerPortal interface { CreatePortalSession(...); CreatePortalConfiguration(...) } // Stripe
type PaymentMethods interface { SetDefaultPaymentMethod(...); DetachPaymentMethods(...) }  // Stripe
```

Use cases check the capability (`if r, ok := provider.(Refunds); ok`) or take it as a dependency. When it's missing, they fall back to the merchant-managed path below.

### 2. Two renewal modes in the domain

|                          | Provider-managed (Stripe, PayPal)             | Merchant-managed (SePay)                                                                                                                       |
| ------------------------ | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Who knows the period end | the provider, applied with `ApplySnapshot`    | **us**: `Subscription` advances its own period                                                                                                 |
| Renewal                  | provider charges, webhook says `invoice.paid` | our cron creates a **renewal invoice** N days before the period end, sends the VietQR/link by email or in-app, and the SePay IPN marks it paid |
| Unpaid                   | provider retries, `past_due`                  | our grace period (`BillingPolicy.PastDueGraceDays` already exists), then `canceled`                                                            |
| Cancel at period end     | provider flag                                 | just don't issue the next invoice                                                                                                              |

The `Subscription` aggregate gains a `RenewalMode` and an `Invoice` child (amount, due date, payment code, status). Everything downstream (entitlements, gRPC, notifications) stays the same, because it already reads only `Status` / `Plan` / `CurrentPeriodEnd`.

### 3. Manual refunds as a first-class status

Without a refund API, a `Refund` stays `pending` and shows up in an **admin queue** with the payer's bank details. An admin makes the transfer and marks it `succeeded` with the bank reference. The reconcile cron already retries `pending` refunds with no provider id; for manual providers it simply skips them.

### 4. One inbound adapter per provider

```
POST /billing/webhooks/stripe  → verify Stripe signature  → Cmds.*
POST /billing/webhooks/paypal  → verify via PayPal API    → Cmds.*
POST /billing/webhooks/sepay   → API key / HMAC           → Cmds.MarkInvoicePaid / CompleteCheckout
```

Each is a separate public Traefik route, just like today's Stripe one.

### 5. Record the provider on the account

Add `provider` to `billing_accounts` (and to `subscriptions` and `prices`, since provider price ids differ). This lets two providers coexist during a migration: existing Stripe subscriptions run until their period ends, and new checkouts go to the new provider. Routing comes from the account's provider, not from a global switch.

---

## Recommendation

1. **Now:** Stripe sandbox, as built. It's the most complete model and proves the domain.
2. **First live provider for Vietnamese customers:** SePay (VND, money straight to our bank, cards through the gateway), with merchant-managed renewals and manual refunds. That's the design work above, mostly in the domain plus a renewal cron, not in the adapter.
3. **International customers in USD:** PayPal, or Stripe through a foreign entity if the team ever has one. It reuses the provider-managed path that already exists.

Until step 2 is scheduled, the only rule is the one we already follow: **nothing outside `infra/payment/<provider>` and that provider's webhook adapter imports a payment SDK.**

---

## Sources

- PayPal currency codes: <https://developer.paypal.com/reference/currency-codes/>
- PayPal in Vietnam (user reports, unofficial): <https://www.expat.com/en/forum/asia/vietnam/543698-how-do-paypal-work-in-vietnam.html>
- VND withdrawal costs (third-party guide): <https://vaultleap.com/blog/receiving-usd-payments-in-vietnam-the-complete-guide-2026-2>
- plutov/paypal Go SDK: <https://github.com/plutov/paypal>
- SePay docs: <https://docs.sepay.vn/>, webhooks: <https://docs.sepay.vn/tich-hop-webhooks.html>
- SePay Payment Gateway: <https://developer.sepay.vn/en/cong-thanh-toan/gioi-thieu>
