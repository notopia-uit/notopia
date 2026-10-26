# Billing web handoff outline

Frontend work for billing enforcement (no code). Backend is done; use `@notopia-uit/api-gen`.

## Billing settings tab (workspace)

- Plan, status, renewal date or access-until date.
- Payer and a "Claim payer" button for another owner.
- Upgrade: opens the checkout URL. Manage: opens the portal URL.
- Cancel now: show the amount from `GET …/cancel-now/preview` before confirming.
- Returning from Stripe with `?checkout=success|payer` or `?portal=return`: poll entitlements
  every 2 s (max 20 s) until the plan changes.

## Pricing page

- From `GET /billing/plans`.
- Format VND with `Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' })`.

## Enforcement UX

- Global 402 handler on the `@notopia-uit/api-gen` client: upgrade dialog using
  `details.requiredPlan` (`code`: `featureNotInPlan` | `planLimitExceeded`; `current`/`max`
  for limits).
- 409 `billingPayerMustTransfer`: tell the user the payer must transfer payer status
  (another owner claims it) before being demoted or removed.
- 503 `billingServiceUnavailable`: generic retry message.
- `<FeatureGate>` reads workspace entitlements (`-1` = unlimited, missing feature = false,
  missing limit = 0). Refetch on SSE `WorkspaceBillingChangedEvent` (also poll as above).
- Downgrade warning: shorter trash and revision retention will purge old data.

## Editor and revisions

- Editor shows "Saved automatically" when the plan has `auto_commit`.
- Manual commit (`POST /document/documents/{id}/commit`) status semantics:
  - `201` revision created
  - `200` already up to date (returns the latest revision)
  - `204` nothing to commit yet (no revision exists)
  - `403` for non-editors (viewers): hide or disable the button
- Revision list collapses `trigger=auto` entries (filter via `trigger`).

## Admin

- Pages for plans, prices and policy; only for group `notopia-admins`.
