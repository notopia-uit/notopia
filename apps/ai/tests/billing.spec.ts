import { of, throwError } from 'rxjs';
import { describe, expect, it, vi } from 'vitest';

import { BillingService } from '../src/billing/billing.service';

const workspaceId = '00000000-0000-4000-8000-000000000001';
const setup = (entitlements: unknown) => {
  const client = {
    getWorkspaceEntitlements: vi.fn().mockReturnValue(of({ entitlements })),
    listPlans: vi.fn().mockReturnValue(
      of({
        plans: [
          { tier: 2, limits: { ai_credits_monthly: 200 } },
          { tier: 3, limits: { ai_credits_monthly: 2000 } },
        ],
      })
    ),
  };
  const service = new BillingService({ getService: () => client } as any);
  service.onModuleInit();
  return { client, service };
};
describe('AI billing integration', () => {
  it('fetches current entitlements on every check', async () => {
    const { client, service } = setup({ workspaceId, limits: { ai_credits_monthly: 200 } });
    expect(await service.monthlyCredits(workspaceId)).toBe(200);
    client.getWorkspaceEntitlements.mockReturnValue(
      of({ entitlements: { workspaceId, limits: { ai_credits_monthly: 0 } } })
    );
    expect(await service.monthlyCredits(workspaceId)).toBe(0);
    expect(client.getWorkspaceEntitlements).toHaveBeenCalledTimes(2);
  });
  it.each([
    undefined,
    { workspaceId: 'another' },
    { workspaceId, limits: { ai_credits_monthly: -2 } },
    { workspaceId, limits: { ai_credits_monthly: '9007199254740993' } },
  ])('rejects invalid entitlements %#', async (value) => {
    await expect(setup(value).service.monthlyCredits(workspaceId)).rejects.toMatchObject({
      status: 503,
    });
  });
  it('treats a missing quota as zero and -1 as unlimited', async () => {
    expect(await setup({ workspaceId }).service.monthlyCredits(workspaceId)).toBe(0);
    expect(
      await setup({ workspaceId, limits: { ai_credits_monthly: -1 } }).service.monthlyCredits(
        workspaceId
      )
    ).toBe(-1);
  });
  it('fails closed when billing is unavailable', async () => {
    const { client, service } = setup({});
    client.getWorkspaceEntitlements.mockReturnValue(throwError(() => new Error('private details')));
    await expect(service.monthlyCredits(workspaceId)).rejects.toMatchObject({
      status: 503,
      message: 'Billing is unavailable. Please retry later.',
    });
  });
  it('uses live plan values to determine upgrade options', async () => {
    const { service } = setup({});
    expect(await service.requiredPlan(0)).toBe('plus');
    expect(await service.requiredPlan(200)).toBe('pro');
  });
});
