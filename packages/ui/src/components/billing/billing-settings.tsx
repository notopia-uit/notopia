'use client';

import { CreditCard, ShieldCheck } from 'lucide-react';
import { useEffect, useState } from 'react';

import { useWorkspaceBilling } from '../../hooks/use-workspace-billing';
import { getAuthClient } from '../../lib/auth-client';
import {
  billingError,
  formatBillingDate,
  formatMoney,
  isCheckoutConfirmed,
  type CheckoutIntent,
} from '../../lib/billing';
import { Badge } from '../shadcn/badge';
import { Button } from '../shadcn/button';
import { CancelSubscriptionDialog } from './cancel-subscription-dialog';
import { PlanComparison } from './plan-comparison';

export function BillingSettings({
  workspaceId,
  comparisonOnly = false,
  returnType = null,
}: {
  workspaceId: string;
  comparisonOnly?: boolean;
  returnType?: string | null;
}) {
  const state = useWorkspaceBilling(workspaceId);
  const { data: session } = getAuthClient().useSession();
  const [cancelOpen, setCancelOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [polling, setPolling] = useState(!!returnType && returnType !== 'cancel');
  const [checkoutIntent] = useState<CheckoutIntent | null>(() => {
    if (typeof window === 'undefined') return null;
    try {
      return JSON.parse(sessionStorage.getItem(`billing-checkout:${workspaceId}`) ?? 'null');
    } catch (error) {
      console.warn('Unable to read the local checkout reference', error);
      return null;
    }
  });
  const refresh = state.refresh;
  const hasWorkspace = !!state.workspace;
  // Poll both views for a bounded time; a redirect parameter is never proof of payment.
  useEffect(() => {
    if (!hasWorkspace || !returnType || returnType === 'cancel') return;
    let stopped = false;
    let timer: ReturnType<typeof setTimeout>;
    let attempts = 0;
    async function poll() {
      await refresh();
      if (stopped) return;
      attempts++;
      if (attempts < 10) timer = setTimeout(() => void poll(), 3000);
      else setPolling(false);
    }
    void poll();
    return () => {
      stopped = true;
      clearTimeout(timer);
    };
  }, [workspaceId, returnType, hasWorkspace, refresh]);

  if (state.loading)
    return (
      <p className="text-muted-foreground py-12 text-center text-sm">Loading workspace billing…</p>
    );
  if (state.loadError || state.missingWorkspace)
    return (
      <div className="space-y-3 rounded-xl border p-6">
        <p role="alert">
          {state.loadError ? billingError(state.loadError) : 'Workspace not found.'}
        </p>
        <Button variant="outline" onClick={() => window.location.reload()}>
          Try again
        </Button>
      </div>
    );
  const sub = state.billing?.subscription;
  const payer = state.billing?.payer;
  const accessEnd = formatBillingDate(sub?.accessUntil ?? state.entitlements?.accessUntil);
  const returnConfirmed =
    returnType === 'success'
      ? isCheckoutConfirmed(checkoutIntent, sub, state.entitlements?.plan)
      : returnType === 'payer'
        ? payer?.id === session?.user.id && !payer?.missingSince
        : false;

  return (
    <div className="space-y-8">
      {!comparisonOnly && (
        <header className="space-y-2">
          <p className="text-muted-foreground text-sm">Workspace settings</p>
          <h1 className="text-3xl font-semibold tracking-tight">Plans & billing</h1>
          <p className="text-muted-foreground">
            Manage the plan for{' '}
            <span className="text-foreground font-medium">{state.workspace?.workspace.name}</span>.
            Your subscription applies to this workspace.
          </p>
        </header>
      )}
      <div className="text-muted-foreground flex items-center gap-2 rounded-lg border border-dashed px-4 py-3 text-sm">
        <ShieldCheck className="size-4 shrink-0" />
        Sandbox payments · No live charges
      </div>
      {returnType && !comparisonOnly && (
        <div className="bg-muted/40 rounded-lg border p-4 text-sm" aria-live="polite">
          {returnType === 'cancel'
            ? 'Checkout was canceled. Your plan has not been changed by this checkout.'
            : returnConfirmed
              ? 'The backend confirms your current billing state below.'
              : returnType === 'portal'
                ? 'Your billing details have been refreshed from the server.'
                : polling
                  ? 'Confirming your payment with the billing service…'
                  : 'Payment confirmation may still be pending. Refresh billing to check again.'}
          <Button className="ml-2" size="sm" variant="ghost" onClick={() => void state.refresh()}>
            Refresh billing
          </Button>
        </div>
      )}
      {notice && <output className="bg-muted block rounded-lg p-4 text-sm">{notice}</output>}
      {state.error && (
        <p
          role="alert"
          className="border-destructive text-destructive rounded-lg border p-4 text-sm"
        >
          {state.error}
        </p>
      )}
      {!comparisonOnly && (
        <section className="bg-card flex flex-wrap items-center justify-between gap-6 rounded-xl border p-6">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-3">
              <CreditCard className="text-primary size-5" />
              <h2 className="text-xl font-semibold capitalize">{state.entitlements?.plan} plan</h2>
              <Badge variant="secondary">{state.entitlements?.status.replaceAll('_', ' ')}</Badge>
              {sub?.interval === 'lifetime' && <Badge>Lifetime</Badge>}
            </div>
            <p className="text-muted-foreground text-sm">
              {sub?.interval === 'lifetime'
                ? 'One-time purchase. No recurring renewal.'
                : sub?.cancelAtPeriodEnd
                  ? `Scheduled to end${accessEnd ? ` on ${accessEnd}` : ''}.`
                  : sub?.currentPeriodEnd
                    ? `Current billing period ends ${formatBillingDate(sub.currentPeriodEnd)}.`
                    : 'A flexible place to start building your knowledge.'}
            </p>
            {accessEnd && <p className="text-sm">Paid access until {accessEnd}</p>}
            {['past_due', 'unpaid', 'incomplete'].includes(state.entitlements?.status ?? '') && (
              <p className="text-destructive text-sm">
                Your payment needs attention. Open billing management to review it.
              </p>
            )}
          </div>
          {state.isOwner && payer && (
            <Button
              variant="outline"
              disabled={state.pending}
              onClick={() => void state.redirect('portal')}
            >
              Manage billing
            </Button>
          )}
        </section>
      )}
      {!state.isOwner && (
        <p className="bg-muted rounded-lg p-4 text-sm">
          You can view this workspace’s plan. Ask a workspace owner to manage subscriptions and
          payments.
        </p>
      )}
      <PlanComparison state={state} />
      {!comparisonOnly && state.isOwner && payer && (
        <section className="space-y-4 rounded-xl border p-6">
          <h2 className="text-lg font-semibold">Payment responsibility</h2>
          <p className="text-sm">
            {payer.name} <span className="text-muted-foreground">· {payer.email}</span>
          </p>
          {payer.missingSince && (
            <p className="text-destructive text-sm">
              The previous payer is no longer an owner. Another owner needs to take over payment
              responsibility.
            </p>
          )}
          {(payer.id !== session?.user.id || payer.missingSince) && (
            <Button
              variant="outline"
              disabled={state.pending || !session}
              onClick={() => void state.redirect('claim')}
            >
              Become payer
            </Button>
          )}
          <p className="text-muted-foreground text-sm">
            Payment methods and available invoices are managed in the billing portal.
          </p>
        </section>
      )}
      {!comparisonOnly && sub?.refunds.length ? (
        <section className="space-y-3 rounded-xl border p-6">
          <h2 className="font-semibold">Subscription refunds</h2>
          {sub.refunds.map((refund) => (
            <p key={refund.id} className="text-sm">
              {formatMoney(refund.amount, refund.currency)} · {refund.status} ·{' '}
              {formatBillingDate(refund.createdAt)}
            </p>
          ))}
        </section>
      ) : null}
      {!comparisonOnly && state.isOwner && sub && (
        <section className="flex flex-wrap items-center justify-between gap-4 border-t pt-6">
          <div>
            <h2 className="font-medium">End your subscription</h2>
            <p className="text-muted-foreground text-sm">
              Preview any eligible refund before canceling immediately.
            </p>
          </div>
          <Button variant="outline" disabled={state.pending} onClick={() => setCancelOpen(true)}>
            Cancel now…
          </Button>
        </section>
      )}
      {cancelOpen && (
        <CancelSubscriptionDialog
          workspaceId={workspaceId}
          onClose={() => setCancelOpen(false)}
          onCancelled={async () => {
            await state.refresh();
            setNotice(
              'Subscription canceled. Any eligible refund has been requested and may take time to process.'
            );
          }}
        />
      )}
    </div>
  );
}
