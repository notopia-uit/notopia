import type { BillingPrice, BillingSubscription } from '@notopia-uit/api-gen';
import { describe, expect, it } from 'vitest';

import { formatAllowance, formatMoney, purchaseAction, isCheckoutConfirmed } from './billing';

const price: BillingPrice = {
  id: 'price',
  plan: 'pro',
  interval: 'month',
  active: true,
  amount: 99000n,
  currency: 'vnd',
  lookupKey: 'pro_month',
  createdAt: new Date(),
};
const subscription: BillingSubscription = {
  id: 'sub',
  plan: 'plus',
  interval: 'month',
  status: 'active',
  cancelAtPeriodEnd: false,
  refunds: [],
};

describe('workspace billing actions', () => {
  it('does not mistake an old active subscription or delayed entitlements for checkout confirmation', () => {
    const intent = { plan: 'pro', interval: 'lifetime', previousId: 'sub' };
    expect(isCheckoutConfirmed(intent, subscription, 'plus')).toBe(false);
    const upgraded = {
      ...subscription,
      id: 'new-sub',
      plan: 'pro' as const,
      interval: 'lifetime' as const,
    };
    expect(isCheckoutConfirmed(intent, upgraded, 'plus')).toBe(false);
    expect(isCheckoutConfirmed(null, upgraded, 'pro')).toBe(false);
    expect(isCheckoutConfirmed(intent, upgraded, 'pro')).toBe(true);
  });
  it('routes recurring subscriptions through the portal except lifetime upgrades', () => {
    expect(purchaseAction(null, price)).toBe('checkout');
    expect(purchaseAction(subscription, price)).toBe('portal');
    expect(purchaseAction(subscription, { ...price, interval: 'lifetime' })).toBe('checkout');
    expect(purchaseAction({ ...subscription, plan: 'pro' }, price)).toBe('current');
  });
  it('prevents duplicate lifetime purchases and archived-price checkout', () => {
    expect(purchaseAction({ ...subscription, interval: 'lifetime' }, price)).toBe('unavailable');
    expect(purchaseAction(null, { ...price, active: false })).toBe('unavailable');
  });
  it('formats VND without dividing by one hundred and preserves unlimited allowances', () => {
    expect(formatMoney(99000n, 'vnd')).toContain('99,000');
    expect(formatMoney(1999n, 'usd')).toContain('19.99');
    expect(formatAllowance(-1n, 'members')).toBe('Unlimited members');
    expect(formatAllowance(0n, 'AI credits / month')).toBe('0 AI credits / month');
  });
});
