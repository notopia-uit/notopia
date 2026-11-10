import type { BillingPrice, BillingSubscription } from '@notopia-uit/api-gen';

export function formatMoney(amount: bigint | number, currency: string) {
  return new Intl.NumberFormat('en', {
    style: 'currency',
    currency: currency.toUpperCase(),
    maximumFractionDigits: currency.toLowerCase() === 'vnd' ? 0 : 2,
  }).format(Number(amount) / (currency.toLowerCase() === 'vnd' ? 1 : 100));
}

export function purchaseAction(
  subscription: BillingSubscription | null | undefined,
  price: BillingPrice
) {
  if (!price.active) return 'unavailable';
  if (!subscription) return 'checkout';
  if (subscription.interval === 'lifetime') return 'unavailable';
  if (price.interval === 'lifetime') return 'checkout';
  if (subscription.plan === price.plan && subscription.interval === price.interval)
    return 'current';
  return 'portal';
}

export type CheckoutIntent = { plan: string; interval: string; previousId: string | null };

export function isCheckoutConfirmed(
  intent: CheckoutIntent | null,
  subscription: BillingSubscription | null | undefined,
  effectivePlan: string | undefined
) {
  return (
    !!intent &&
    !!subscription &&
    subscription.status === 'active' &&
    subscription.plan === intent.plan &&
    subscription.interval === intent.interval &&
    subscription.id !== intent.previousId &&
    effectivePlan === intent.plan
  );
}

export function formatBillingDate(value: Date | null | undefined) {
  return value ? new Date(value).toLocaleDateString(undefined, { dateStyle: 'medium' }) : null;
}

export function billingError(error: unknown) {
  if (
    error &&
    typeof error === 'object' &&
    'message' in error &&
    typeof error.message === 'string'
  ) {
    return error.message;
  }
  return 'Unable to update billing. Please try again.';
}

export function formatAllowance(value: bigint | number | undefined, suffix: string) {
  if (value === undefined) return null;
  return Number(value) === -1
    ? `Unlimited ${suffix}`
    : `${Number(value).toLocaleString()} ${suffix}`;
}
