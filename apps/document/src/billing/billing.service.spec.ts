import { PlanTier as PlanTierPb } from '@notopia-uit/pb/billing';

import { toEntitlements, toLimits } from './billing.service';
import { PlanFeature, PlanLimit, PlanTier } from './plan';

describe('billing mappers', () => {
  it('maps known keys and drops unknown ones', () => {
    const result = toEntitlements({
      workspaceId: 'w',
      plan: PlanTierPb.PLAN_TIER_PLUS,
      status: 0,
      features: { [PlanFeature.AutoCommit]: true, mystery: true },
      limits: { [PlanLimit.RevisionHistoryDays]: 30, mystery: 1 },
    });
    expect(result).toEqual({
      plan: PlanTier.Plus,
      features: { auto_commit: true },
      limits: { revision_history_days: 30 },
    });
  });

  it('converts Long-like limits to number', () => {
    expect(toLimits({ [PlanLimit.RevisionHistoryDays]: { toNumber: () => 7 } })).toEqual({
      revision_history_days: 7,
    });
  });

  it.each([PlanTierPb.PLAN_TIER_UNSPECIFIED, PlanTierPb.UNRECOGNIZED])(
    'throws on tier %s',
    (plan) => {
      expect(() =>
        toEntitlements({ workspaceId: 'w', plan, status: 0, features: {}, limits: {} })
      ).toThrow(/billing/i);
    }
  );

  it('throws on missing entitlements', () => {
    expect(() => toEntitlements(undefined)).toThrow(/billing/i);
  });
});
