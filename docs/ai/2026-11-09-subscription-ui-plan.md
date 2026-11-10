# Workspace subscription UI plan

## Scope

Build the workspace upgrade and billing experience on branch `ui/subscription`, using the
existing billing service and generated API client. This is a frontend-only change in
`apps/web` and `packages/ui`; payment collection remains in hosted Stripe Checkout and Portal.

## Backend findings

- Billing is workspace-scoped. Authorization policy grants `billing_manage` to owners only.
  Members can read effective entitlements, but subscription and payer details require an owner.
- `listPlans` returns tier, feature flags, limits, and active prices. Render values from the
  response rather than duplicating the default catalog in the frontend.
- Catalog tiers are Free, Plus, and Pro; intervals are month, year, and lifetime. The seeded
  lifetime offer is Pro only. Missing prices mean that purchase option is unavailable.
- Seeded prices are Plus 49,000 VND/month or 470,400 VND/year; Pro 99,000 VND/month,
  950,400 VND/year, or 2,490,000 VND lifetime. These are reference values, not UI constants.
- Features include automatic commits and workspace graph. Limits include members, published
  notes, graph depth, automatic commit interval, revision history, trash retention, and monthly
  AI credits. A limit of `-1` means unlimited.
- New subscriptions use `createCheckoutSession({priceId})`. An existing current subscription
  normally requires Portal; the checkout handler makes an exception for recurring-to-lifetime.
- `createPortalSession` manages billing through Stripe. `claimPayer` opens a setup Checkout
  for another owner to take over payment responsibility.
- `previewCancelNow` provides a refund estimate. `cancelNow` cancels local access and requests
  asynchronous provider cancellation/refunds. Lifetime refunds can be refused after the window.
- Return URLs already point to `/workspace/{id}/settings/billing`, with `checkout=success`,
  `checkout=cancel`, `checkout=payer`, or `portal=return`.
- The generated client already exports billing queries and mutations. No API regeneration is
  needed for the existing contract.

## Design proposal

Use one Billing settings page as the permanent destination, with an upgrade dialog reusing
the same plan comparison. This keeps subscription details in settings while letting users
compare plans without leaving their current note.

### Billing page

1. **Workspace heading:** name and an explicit “Plan applies to this workspace” description.
2. **Current plan:** tier, effective access, subscription status, monthly/yearly/lifetime,
   renewal or access-end date, and scheduled cancellation where available.
3. **Plan comparison:** Free / Plus / Pro cards, monthly/yearly selector, current-plan badge,
   benefits, and separate Pro lifetime offer. Show the full yearly charge even when also
   showing a monthly equivalent. Calculate savings only from comparable returned prices.
4. **Payment management:** current payer, Manage billing, and eligible Become payer action.
5. **Cancellation:** separate Cancel now action, with a freshly loaded refund estimate and
   explicit confirmation of immediate loss of paid access.

Use existing shadcn components and semantic Tailwind tokens. Cards stack on small screens;
the upgrade dialog scrolls without hiding its close control or primary action.

### Actions by workspace state

| State                          | Experience                                                                                     |
| ------------------------------ | ---------------------------------------------------------------------------------------------- |
| Free / no current subscription | Compare plans and continue to Checkout                                                         |
| Recurring paid subscription    | Current plan badge; Manage/change plan through Portal; separate eligible lifetime purchase     |
| Lifetime subscription          | Lifetime badge, no renewal date, no duplicate purchase or recurring-plan CTA                   |
| Payment problem                | Payment-status banner and Manage billing action; effective access comes from entitlements      |
| Scheduled cancellation         | Show access-end date and Portal action to manage the subscription                              |
| Missing payer                  | Explain payment responsibility and offer eligible owner a payer-claim flow                     |
| Non-owner member               | Effective plan and comparison only; “Ask a workspace owner to upgrade”; no private payer fetch |

## API wiring

| UI capability                                                        | Existing operation                   |
| -------------------------------------------------------------------- | ------------------------------------ |
| Plan comparison                                                      | `listPlans`                          |
| Effective plan/features/allowances                                   | `getWorkspaceEntitlements`           |
| Owner subscription/payer details                                     | `getWorkspaceBilling`                |
| New purchase / eligible lifetime upgrade                             | `createCheckoutSession`              |
| Change recurring plan/payment method and provider billing management | `createPortalSession`                |
| Become payer                                                         | `claimPayer`                         |
| Refund confirmation                                                  | `previewCancelNow`, then `cancelNow` |

Scope query keys and local purchase state to the workspace. Disable duplicate actions while
a mutation is pending. Recover from stale prices and subscription conflicts by refreshing
the relevant queries and explaining the available next action.

On Checkout/Portal return, refetch billing and entitlements with bounded polling. Do not
infer payment success from the URL alone. If confirmation is delayed, show “Payment is being
confirmed” with a manual refresh action rather than claiming an upgrade or retrying a charge.

## Implementation sequence

1. **Billing view model and API hooks:** derive permissions from existing workspace membership;
   build explicit action rules and currency/interval/limit formatting around generated types.
2. **Billing settings page:** add route, client boundary, loading/error states, current-plan
   summary, and responsive comparison. Consume existing APIs for real data.
3. **Upgrade entry points:** wire the sidebar Upgrade and Billing items; reuse plan comparison
   in a dialog opened from Upgrade. Make labels reflect the current workspace plan.
4. **Hosted payment flows:** Checkout, Portal, payer claim, and return-state refresh. Keep card
   inputs and payment-method handling in the provider's hosted pages.
5. **Cancellation and recovery:** refund preview/confirmation, asynchronous-refund messaging,
   payment issues, missing payer, and readable permission/conflict errors.
6. **Verification and polish:** verify action selection, price formatting, workspace switching,
   pending states, keyboard navigation, and narrow layouts; run affected lint/typecheck/tests
   and the web build. Smoke-test real sandbox Checkout and return synchronization separately
   from entitlement-only fixtures, whose provider IDs are fake.

## Proposed file boundaries

- `apps/web/app/(workspace)/workspace/[workspaceId]/settings/billing/page.tsx`
- Adjacent `billing-boundary.tsx` for client error handling.
- `packages/ui/src/components/billing/billing-settings.tsx`
- `packages/ui/src/components/billing/plan-comparison.tsx`
- `packages/ui/src/components/billing/upgrade-dialog.tsx`
- `packages/ui/src/components/billing/cancel-subscription-dialog.tsx`
- `packages/ui/src/hooks/use-workspace-billing.ts`
- `packages/ui/src/lib/billing.ts` for formatting and action derivation, with meaningful tests.
- Update `packages/ui/src/components/workspace-sidebar.tsx`; reuse settings navigation.

## Contract limits to respect

- Entitlements expose allowances, not consumed usage. Show “200 AI credits/month”, for
  example, rather than inventing a remaining-credit counter or a usage progress bar.
- Subscription responses do not expose the purchased price/amount. Current public catalog
  prices are offers, not evidence of an existing customer's billed amount. Invoice/payment
  details belong in Portal.
- There is no general subscription/refund history endpoint. Refunds attached to a returned
  subscription may be shown, but do not promise a complete history after cancellation.
- Refund preview is an estimate; cancellation recalculates it. A successful cancel request
  does not mean money has already arrived.
- The current provider integration is sandbox-only; use clear test-payment labeling.

## Implementation status

Implemented the billing route, responsive shared plan comparison and upgrade dialog,
owner-only hosted payment actions, bounded return-state refresh, payer claims, and immediate
cancellation with a refund preview. The sidebar now links to billing and opens Explore plans.
Settings layout adapts to small screens and gives the plan comparison sufficient width.

Checkout confirmation checks the selected plan/interval, a changed subscription ID, active
status, and matching effective entitlements. Merely returning to a success URL or having an
older active subscription is not treated as evidence of a completed upgrade.

Verification:

- UI unit suite: 23 tests passed, including action routing, lifetime/archived-price guards,
  currency formatting, unlimited limits, and delayed checkout confirmation.
- UI typecheck and web production build passed (including the web TypeScript check).
- Scoped billing lint: no errors or warnings.
- Full UI lint remains blocked by 17 existing `require-await` errors in editor/assistant
  files. Web lint has existing warnings but no errors.
- Web test command remains blocked by a stale local Vitest binary path, even after a frozen
  dependency install. No web test files currently exist.
- Browser visual review and authenticated sandbox payment smoke tests were not performed:
  the desktop browser is disconnected and local identity-provider discovery failed during
  the build. Successful compilation does not establish real Checkout/webhook behavior.
