# Billing price rollout and Stripe reconciliation

Status: Draft — pending human approval. No implementation is authorized by this document.

## Plain-language summary

An administrator changes a monthly or annual price once through Notopia. New customers see the new price immediately after publication. Notopia updates eligible existing subscriptions in the background. Stripe collects the new amount at each customer's next renewal, without an extra charge for the current paid period.

Setup also becomes safe to repeat: Notopia checks the Stripe resources it owns, repairs supported differences, and reuses existing resources instead of creating duplicates. Running setup does not itself authorize increasing existing customers' subscription prices.

This extends the current billing service. It does not replace checkout, refunds, entitlements, or the payment-provider abstraction.

## Intent and confirmed decisions

- The future web admin portal must be able to use the billing API to change prices.
- Both monthly and annual recurring prices are in scope. Each subscription keeps its interval, currency, quantity, and renewal date.
- Existing subscribers eventually pay the replacement price; they are not permanently grandfathered.
- Price changes must not charge a mid-period difference or shorten time already purchased.
- A background operation replaces manual edits to each customer's Stripe subscription.
- Include improvements to catalog and portal-configuration reconciliation where needed for this flow.
- Do not add a custom customer-notification feature. This does not waive applicable notice, contractual, or payment-network requirements.
- The earlier phrase “change directly” was withdrawn and supplies no requirement.

The current Stripe integration is sandbox-only, as documented in `internal/billing/AGENTS.md`. This spec therefore verifies and delivers the flow in sandbox without a notice period. It does not authorize live price increases. Before enabling a future live payment provider, review applicable notice/consent requirements and customer terms; add any required eligibility delay or communication separately. No arbitrary 30-day notice rule is assumed here.

## Engineering proposals, not additional human decisions

- Notopia remains the owner of its Stripe products, prices, and customer portal configuration. Terraform adoption is deferred.
- Use a direct same-interval subscription item update with `proration_behavior=none` and unchanged billing dates for ordinary eligible subscriptions.
- Persist operations and per-subscription progress, trigger processing through the existing transactional outbox, and use reconciliation as a retry safety net.
- Reuse the current billing process and infrastructure; do not introduce a separate worker service or new queue system.
- Keep the implementation in two dependent parts: reliable catalog/configuration reconciliation, followed by automatic subscriber rollout.

## Current implementation

- `api/billing/paths/admin_prices.yaml` already defines an admin-only `POST /billing/admin/prices` that replaces the current price for a plan and interval.
- `internal/billing/controller/http/admin.go` implements that endpoint, generates a fresh price UUID, and delegates to `CreatePrice`.
- `internal/billing/app/catalog.go` creates the provider price before saving the local replacement. It then archives the previous Stripe price and creates a fresh portal configuration.
- `internal/billing/seed/seed.go` creates default prices only if no active local price exists for the same plan, interval, and currency. It deliberately does not overwrite an administrator's price with a hard-coded default amount.
- `internal/billing/app/cmd_ensureportalconfiguration.go` only checks whether a portal configuration ID exists. It does not check the provider's current settings.
- `internal/billing/infra/payment/stripe/catalog.go` creates prices using the supplied local price UUID as an idempotency key. Fresh UUIDs on later HTTP or seed attempts do not identify the same logical operation. Portal creation has no equivalent recovery path.
- `internal/billing/app/svc_paymentprovider.go` has no subscription-price replacement operation or portal retrieve/update reconciliation operation.
- `internal/billing/app/cmd_syncsubscription.go` fetches Stripe state, resolves its provider price to a local price, and updates the local subscription. It must continue to work with archived historical price records.
- `internal/billing/app/cmd_reconcilesubscriptions.go` already retries existing cancellation/refund work and re-syncs stale subscriptions. It does not discover or execute price rollouts today.

## Approaches considered

### Application-owned catalog and automatic rollout — proposed

Retain the existing admin price API and database catalog. Add durable provider operations, portal reconciliation, and a background subscription rollout. This fits the existing admin workflows, database-linked checkout, outbox, and future non-Stripe provider direction.

Cost: Notopia must implement failure recovery and subscriber rollout tracking.

### Terraform-owned catalog and settings — deferred

Terraform can manage products, prices, and portal configuration with reviewed plans. It would require importing existing resources, disabling conflicting application writes, and feeding provider identifiers back into Notopia.

It does not automatically migrate existing subscribers when a catalog price changes. The subscriber-update operation would still be needed. This is additional ownership migration, not a substitute for this feature.

### Stripe Dashboard or external bulk tool — not the requested flow

An operator could change each subscription, or delegate bulk changes to a reviewed external tool. This reduces application code but does not provide the agreed one-action Notopia workflow.

### Subscription schedules versus direct updates

Schedules preserve the old subscription price until the transition date. Direct same-interval updates without proration change the subscription's price reference now, while the next regular renewal uses the new amount.

Use direct updates for this design because deferred charging, not deferred display of the price reference, is the requirement. Schedules remain relevant for subscriptions already carrying a pending plan change; this version must not silently replace those schedules.

## Scope and non-goals

Included: price publication, historical price retention, owned-resource reconciliation, durable rollout progress, retries, and an admin API for observing outcomes and retrying blocked work after correction.

Excluded: admin web UI, Terraform migration, new payment providers, interval/currency conversion, changes to lifetime purchases, forced collection of unpaid invoices, notifications, and a general subscription-schedule editor.

Do not rewrite prior invoices, calculate refunds from the newly published catalog amount, create duplicate subscription items, revive canceled subscriptions, or silently override customers' pending plan changes.

## Part A: reliable catalog and portal reconciliation

### Ownership and desired state

- Existing local active price records are authoritative for amounts and availability. Seed constants initialize missing catalog entries, not overwrite administrator changes.
- Track explicit administrative retirement of a price slot so a later seed run cannot resurrect it merely because no active row remains. Legacy archived-only slots require review rather than automatic resurrection when intent is ambiguous.
- The desired portal options come from active recurring local prices and the app-owned settings currently built in `stripe/catalog.go`. Lifetime prices remain excluded from subscription switching.
- Billing-policy grace/refund windows remain local rules, not automatically Stripe portal fields.
- Tag newly managed Stripe objects with application identity, environment/catalog identity, and stable local operation/resource identifiers. Adopt legacy objects only through verified local references or an explicit migration procedure.
- Do not mutate unrelated products, prices, or portal configurations. An API error is not evidence that a resource is missing.

### Durable price publication

1. Persist a logical publication operation with a stable UUID, requested plan/interval/currency/amount, actor, request identity, and progress before its first Stripe creation attempt. This pending intent is separate from the active catalog change.
2. An admin retry must reuse that operation identity. Same identity with different input is rejected; repeating an unchanged current amount is a no-op rather than a new rollout.
3. Find an already-created provider price by its saved provider ID or verified operation metadata before creating another. Verify product, currency, interval, amount, and ownership. Lookup keys are convenience pointers, not sufficient proof of ownership or operation identity.
4. Create a price when needed using the persisted operation UUID as the Stripe idempotency key. Stripe can expire idempotency records after at least 24 hours; durable identity plus provider discovery is necessary for later retries.
5. In one local transaction, activate the new price, archive the previous local active row, record the publication, and durably request portal/archive work and any authorized subscriber rollout.
6. Retain previous price rows and their provider IDs for historical subscriptions, invoices, and webhook synchronization. Archive old Stripe prices instead of deleting them.
7. Record and retry provider archive/portal failures after publication. Do not report a committed catalog change as if nothing happened or invite a fresh duplicate publication.

Provider calls must not run inside a long database transaction. Synchronize conflicting mutations per plan/interval/currency using the service's existing locking patterns and persist enough identity to recover after lock loss or process restart. A stale operation must not overwrite a newer active price.

### Reconciliation behavior

- `just seed-billing` initializes missing defaults, reconciles managed active prices, and reconciles the portal once after all catalog entries settle. It must not call a create-new-portal flow after each missing price.
- Retrieve the saved portal configuration. Compare only app-owned fields, normalizing unordered product/price lists. Do nothing when equal; update that configuration when different.
- Preserve fields outside application ownership. For owned settings, Dashboard edits are drift and the next reconciliation restores Notopia's desired state.
- If no configuration exists, create it with durable operation identity and recover it after a failed local save. Replace a verified deleted configuration when necessary. Authentication, connectivity, and rate-limit failures must return/retry errors, not trigger replacement.
- Inspect a referenced inactive owned configuration and reactivate it when appropriate; do not create another solely because it is inactive.
- Verify referenced prices against their local immutable definition. Repair mutable differences such as active state when authorized. If the provider reference is missing or immutable values conflict, block publication/checkout for the affected slot as appropriate and expose an actionable reconciliation failure. Do not silently charge a different amount or rebind historical subscriptions.
- Repairs to an existing price reference do not authorize a subscriber price increase. A new subscriber rollout requires an explicit successful admin publication. Initial seed creates no rollout. Existing pre-feature subscriptions are not migrated merely by enabling this feature.
- Explicitly archiving a price removes it from new purchases/portal options and triggers reconciliation; it does not invent a replacement or migrate subscribers.
- Reuse the existing cron infrastructure for retrying durable catalog operations. One failing catalog operation must not prevent other existing cancellation/refund reconciliation work.

## Part B: automatic subscriber price rollout

### Publication and target selection

- A successful admin replacement of a recurring price records a rollout target in the same local transaction as the catalog change. A lifetime price publication has no subscriber rollout.
- The rollout applies within the same plan, interval, and currency. It targets eligible current subscriptions on the replaced price and any older prices covered by earlier admin rollouts for that slot, so repeated price changes converge to the latest authorized target.
- Track rollout lineage explicitly. Do not sweep unrelated legacy/custom price rows solely because they share a plan and interval.
- Before mutation, resolve the latest authorized target again. Older pending work becomes superseded rather than moving subscribers back to an obsolete price.
- Discover candidates in bounded, paginated batches. Do not hold the admin HTTP request open while processing every subscription.
- Also discover late arrivals through subscription sync/checkout completion and periodic reconciliation. An old Checkout session completed after publication must not permanently escape the applicable rollout. Discovery must distinguish initial seeding from an authorized admin replacement.

### Per-subscription processing

1. Claim durable work and verify the subscription is still current locally.
2. Retrieve fresh provider state, including subscription items, billing interval, cancellation flags, pending updates, and any attached schedule. The existing snapshot needs an additional provider operation/result model to expose the necessary eligibility data.
3. Verify that the provider subscription belongs to the expected workspace/customer, contains the supported single recurring item, and still uses a source price covered by this rollout.
4. If already at the target, synchronize local state and mark the item complete. If the customer switched to another plan or an unrelated price, skip it without overriding that choice.
5. Replace the existing subscription item's price. Never append another item. Preserve quantity and billing dates; set `proration_behavior=none`; do not reset the billing anchor or create an immediate invoice.
6. Re-fetch or use a verified resulting provider snapshot, apply the existing local synchronization rules, and persist completion. Provider success followed by local failure must be recoverable by observing that the target price is already applied.

Use stable per-item operation identifiers for retries. Recheck the latest target and eligibility after acquiring the relevant mutation lock. Multiple app replicas and outbox redelivery must not cause duplicate or backward changes. App locks cannot prevent Stripe Dashboard/portal races; never claim cross-system atomicity. If a conflicting provider change is detected, stop and re-evaluate rather than blindly repeat stale mutations.

### Eligibility and exceptions

- Automatically update ordinary active monthly/annual subscriptions only.
- Skip canceled/expired subscriptions, local cancellation in progress, cancel-at-period-end subscriptions, and lifetime purchases.
- Defer `incomplete`, `past_due`, and `unpaid` subscriptions until they become eligible; do not rewrite old unpaid invoices or trigger collection. Reconciliation must revisit deferred items.
- A subscription with an attached schedule, pending update, unsupported item structure, or an unresolved concurrent change requires attention. Preserve its current setup. Expose the reason and support retry after it has been resolved; do not mark it successfully migrated.
- After ordinary plan changes, cancellation changes, payment recovery, or schedule release, re-evaluate deferred or blocked work using current state and the latest target.
- A publication replacing an unfinished rollout supersedes stale pending targets. Already completed customers and eligible older-price customers converge to the latest target; paid invoices remain unchanged.

### Billing and display semantics

- A same-interval replacement without proration updates the subscription price reference immediately. API consumers may therefore display the new recurring price before its first charge.
- Paid amount/history must come from actual invoices/payments, not the current catalog amount. Review billing views and refund paths for this distinction.
- Stripe handles future recurring invoices. The worker does not sleep until renewal and does not charge the customer itself.
- If an invoice is already finalized, or an outage prevents rollout before renewal, do not promise or perform a retroactive extra charge. Report the delay and apply the new price for the next invoice that can still use it. Validate the exact invoice-boundary behavior in Stripe sandbox.

## Durable state and administration API

Persist publication/rollout identity, source lineage, target local/provider IDs, actor, timestamps, and aggregate progress. Persist per-subscription target revision, status, attempt count, next retry time, last error category/message, and lease/claim information as needed.

Item outcomes must distinguish pending, retryable failure, deferred, completed, skipped, superseded, and needs attention. A rollout with unresolved exceptions is not fully completed. Counts and queries must be based on durable state rather than logs alone; late candidates can reopen discovery/progress.

Extend the API contract in `api/billing/` before generating servers/clients:

- Preserve the existing price-publication operation and admin authorization. Add an explicit repeatable request identity and expose the publication/rollout reference without requiring subscribers to finish before the price response.
- Provide admin-only publication/rollout status, paginated item outcomes, and explicit retry after correction. Keep committed publication status distinct from follow-up Stripe synchronization failures.
- Initial new price creation and unchanged-price requests must not produce misleading migration counts.
- Follow CQRS: commands return errors only; the controller supplies identifiers and uses queries for responses. Do not expose Stripe SDK types outside the adapter.

Exact endpoint names, status enums, schema decomposition, retry delays, and claim implementation are builder decisions. They must preserve these observable behaviors and be documented in the implementation plan. No manual SQL or Stripe Dashboard work is required for the ordinary successful flow; exceptional conflicts remain visible for an operator.

## Architecture and delivery boundaries

- Follow `docs/src/conventions/go-clean-architecture.md` and `internal/billing/AGENTS.md`.
- Add domain operation/progress models, repository interfaces, and transactional writes as needed. Generate SQLC and Wire code through existing Nx targets.
- Keep Stripe reads/writes behind `app.PaymentProvider`. Verify exact generated APIs against pinned `stripe-go/v87`.
- Publish durable work through Watermill's transactional outbox and existing event-topic constants. Consumers delegate to application commands; cron provides bounded retry/discovery.
- Use typed internal event payloads; if a new cross-service contract is genuinely needed, define it once in `api/share/`. Do not invent cross-service events for internal-only progress.
- Reuse existing service logging and rate-limit handling. Use bounded concurrency, backoff, and persistent progress; retryable failures must not starve newer candidates.
- Split delivery into catalog/configuration recovery first and subscriber rollout second. Neither part requires a standalone worker deployment.
- Update billing service documentation when model/schema behavior changes.

## Acceptance evidence

1. Monthly and annual replacements retain quantity and renewal dates, create no immediate adjustment invoice, and bill the new price on the next eligible renewal in Stripe sandbox. Use test clocks where supported.
2. New checkout uses the new active price. Existing subscribers remain correctly entitled while rollout is pending.
3. Ordinary active subscribers update automatically. Lifetime, cancellation, arrears, existing schedules, pending updates, and customer plan switches follow the exception rules above.
4. Refund quotes and paid-history responses still use actual payments after an uncharged price-reference change.
5. Outbox redelivery, concurrent workers, process restart, Stripe-success/local-failure, and expired Stripe idempotency records do not create duplicate owned resources or duplicate subscription items.
6. Two successive price publications converge on the latest authorized target. Old Checkout completions and recovered past-due subscriptions are discovered without migrating unrelated legacy prices.
7. Repeated seed/reconcile runs make no provider changes when state matches. Hard-coded seed defaults do not overwrite admin prices or resurrect intentionally retired slots.
8. Portal drift is repaired in place, unordered lists compare equal, and unrelated Dashboard settings/resources remain untouched. A failed retrieve does not cause creation.
9. Provider archive/portal failures remain retriable after catalog commit. Admin status distinguishes publication success from incomplete follow-up work.
10. API authorization, request identity conflicts, operation pagination, error visibility, retry, and rollout counts have focused tests.
11. The implementation plan selects project-specific Go tests, persistence integration tests, API generation/lint, `nx lint billing`, and a sandbox walkthrough. Do not contact live Stripe or use live payments for verification.

## Workflow checklist

- [x] Research existing billing code and Stripe approaches.
- [x] Discuss automatic rollout and renewal-time charging with the user.
- [x] Include monthly and annual subscriptions and relevant reconciliation improvements.
- [x] Draft the design and distinguish human decisions from engineering proposals.
- [x] Run a fresh builder review and resolve material gaps. The reviewer raised notice requirements; clarified sandbox-only delivery and the future live-provider review gate without reopening the user's notification decision.
- [ ] Obtain human approval of the reviewed spec.
- [ ] Invoke writing-plans after approval. Do not implement before approval.

## Sources

- [Investigation handoff](../../../tmp/billing/stripe-catalog-reconciliation-and-terraform-evaluation.md) — local research context; may be untracked.
- [Stripe: manage products and prices](https://docs.stripe.com/products-prices/manage-prices).
- [Stripe: change existing subscription prices](https://docs.stripe.com/billing/subscriptions/change-price).
- [Stripe: subscription schedules and existing-schedule considerations](https://docs.stripe.com/billing/subscriptions/subscription-schedules).
- [Stripe: update portal configuration](https://docs.stripe.com/api/customer_portal/configurations/update).
- [Stripe: idempotent requests and retention limits](https://docs.stripe.com/api/idempotent_requests).
- [Stripe: official Terraform provider](https://docs.stripe.com/terraform/install).
- [Stripe engineering: Terraform and reviewed configuration](https://stripe.dev/blog/ai-agents-terraform-stripe-infrastructure).
- [Vercel subscription starter: externally managed catalog synchronization](https://github.com/vercel/nextjs-subscription-payments).
