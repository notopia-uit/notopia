'use client';

import { Check, Sparkles } from 'lucide-react';
import { useState } from 'react';

import type { WorkspaceBillingState } from '../../hooks/use-workspace-billing';
import { formatAllowance, formatMoney, purchaseAction } from '../../lib/billing';
import { Badge } from '../shadcn/badge';
import { Button } from '../shadcn/button';

const allowances = [
  ['members', 'members'],
  ['published_notes', 'published notes'],
  ['ai_credits_monthly', 'AI credits / month'],
  ['revision_history_days', 'days of revision history'],
  ['trash_retention_days', 'days of trash retention'],
  ['note_graph_max_depth', 'graph depth'],
] as const;

export function PlanComparison({ state }: { state: WorkspaceBillingState }) {
  const [interval, setInterval] = useState<'month' | 'year'>('month');
  const current = state.entitlements?.plan;
  return (
    <section className="@container space-y-6" aria-label="Compare workspace plans">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-semibold">More room for your ideas</h2>
          <p className="text-muted-foreground text-sm">Choose the right plan for this workspace.</p>
        </div>
        <div className="inline-flex rounded-lg border p-1" aria-label="Billing interval">
          {(['month', 'year'] as const).map((value) => (
            <Button
              key={value}
              size="sm"
              variant={interval === value ? 'secondary' : 'ghost'}
              aria-pressed={interval === value}
              onClick={() => setInterval(value)}
            >
              {value === 'month' ? 'Monthly' : 'Yearly'}
            </Button>
          ))}
        </div>
      </div>
      <div className="grid gap-4 @min-[44rem]:grid-cols-3">
        {[...state.plans]
          .sort(
            (a, b) =>
              ['free', 'plus', 'pro'].indexOf(a.tier) - ['free', 'plus', 'pro'].indexOf(b.tier)
          )
          .map((plan) => {
            const price = plan.prices.find((item) => item.active && item.interval === interval);
            const action = price
              ? purchaseAction(state.billing?.subscription, price)
              : 'unavailable';
            return (
              <article
                key={plan.tier}
                className={`flex flex-col gap-5 rounded-xl border p-6 ${plan.tier === 'pro' ? 'border-primary bg-primary/5' : 'bg-card'}`}
              >
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold capitalize">{plan.tier}</h3>
                  {current === plan.tier && <Badge variant="secondary">Current plan</Badge>}
                </div>
                <div>
                  <p className="text-3xl font-semibold tracking-tight">
                    {plan.tier === 'free'
                      ? 'Free'
                      : price
                        ? formatMoney(price.amount, price.currency)
                        : 'Unavailable'}
                  </p>
                  <p className="text-muted-foreground mt-1 text-sm">
                    {plan.tier === 'free'
                      ? 'For getting started'
                      : interval === 'year'
                        ? 'Billed yearly, per workspace'
                        : 'Per month, per workspace'}
                  </p>
                </div>
                <ul className="flex-1 space-y-3 text-sm">
                  {allowances.map(([key, label]) => {
                    const text = formatAllowance(plan.limits[key], label);
                    return text ? (
                      <li key={key} className="flex gap-2">
                        <Check className="text-primary mt-0.5 size-4 shrink-0" />
                        {text}
                      </li>
                    ) : null;
                  })}
                  {plan.features.auto_commit && (
                    <li className="flex gap-2">
                      <Check className="text-primary size-4 shrink-0" />
                      Automatic commits
                      {plan.limits.auto_commit_min_interval_seconds
                        ? ` (every ${Number(plan.limits.auto_commit_min_interval_seconds)} seconds or longer)`
                        : ''}
                    </li>
                  )}
                  {plan.features.workspace_graph && (
                    <li className="flex gap-2">
                      <Check className="text-primary size-4 shrink-0" />
                      Workspace graph
                    </li>
                  )}
                </ul>
                <Button
                  variant={plan.tier === 'pro' ? 'default' : 'outline'}
                  disabled={
                    !state.isOwner ||
                    state.pending ||
                    !price ||
                    action === 'current' ||
                    action === 'unavailable'
                  }
                  onClick={() =>
                    price &&
                    void state.redirect(action === 'portal' ? 'portal' : 'checkout', price.id)
                  }
                >
                  {!state.isOwner
                    ? 'Ask an owner to upgrade'
                    : plan.tier === 'free'
                      ? current === 'free'
                        ? 'Current plan'
                        : 'Free after cancellation'
                      : action === 'current'
                        ? 'Current subscription'
                        : action === 'portal'
                          ? 'Change in billing portal'
                          : action === 'unavailable'
                            ? 'Unavailable'
                            : `Choose ${plan.tier}`}
                </Button>
              </article>
            );
          })}
      </div>
      {state.plans
        .flatMap((plan) =>
          plan.prices.filter((price) => price.active && price.interval === 'lifetime')
        )
        .map((price) => (
          <article
            key={price.id}
            className="bg-muted/40 flex flex-wrap items-center justify-between gap-4 rounded-xl border p-6"
          >
            <div className="space-y-1">
              <h3 className="flex items-center gap-2 font-semibold">
                <Sparkles className="size-4" />
                {price.plan === 'pro' ? 'Pro' : price.plan} lifetime
              </h3>
              <p className="text-muted-foreground text-sm">
                {formatMoney(price.amount, price.currency)} once. No recurring subscription.
              </p>
            </div>
            <Button
              variant="outline"
              disabled={
                !state.isOwner ||
                state.pending ||
                purchaseAction(state.billing?.subscription, price) !== 'checkout'
              }
              onClick={() => void state.redirect('checkout', price.id)}
            >
              {state.billing?.subscription?.interval === 'lifetime'
                ? 'Lifetime plan active'
                : 'Get lifetime access'}
            </Button>
          </article>
        ))}
    </section>
  );
}
