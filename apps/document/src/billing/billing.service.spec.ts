import { PlanTier as PlanTierPb } from '@notopia-uit/pb/billing';

import { toEntitlements, toLimits } from './billing.service';
import { PlanTier } from './plan';

describe('billing mappers', () => {
  it('maps typed feature and limit fields', () => {
    const result = toEntitlements({
      workspaceId: 'w',
      plan: PlanTierPb.PLAN_TIER_PLUS,
      status: 0,
      features: { autoCommit: true, workspaceGraph: true },
      limits: { revisionHistoryDays: 30, members: 10 },
    });
    expect(result).toEqual({
      plan: PlanTier.Plus,
      features: { auto_commit: true },
      limits: { revision_history_days: 30 },
    });
  });

  it('converts Long-like limits to number', () => {
    expect(toLimits({ revisionHistoryDays: { toNumber: () => 7 } as unknown as number })).toEqual({
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
