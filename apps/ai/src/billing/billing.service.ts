import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  type BillingServiceClient,
  PlanTier,
} from '@notopia-uit/pb/billing';
import { firstValueFrom, timeout } from 'rxjs';

import { AiError } from '#/common/ai-error';

@Injectable()
export class BillingService implements OnModuleInit {
  private client!: BillingServiceClient;
  private readonly logger = new Logger(BillingService.name);

  constructor(@Inject(BILLING_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit() {
    this.client = this.grpc.getService<BillingServiceClient>(BILLING_SERVICE_NAME);
  }

  async requiredPlan(currentLimit: number): Promise<string> {
    try {
      const response = await firstValueFrom(this.client.listPlans({}).pipe(timeout(5000)));
      for (const [tier, name] of [
        [PlanTier.PLAN_TIER_PLUS, 'plus'],
        [PlanTier.PLAN_TIER_PRO, 'pro'],
      ] as const) {
        const plan = response.plans.find((plan) => plan.tier === tier);
        const quota = Number(plan?.limits?.ai_credits_monthly ?? 0);
        if (Number.isSafeInteger(quota) && (quota === -1 || quota > currentLimit)) return name;
      }
      return 'pro';
    } catch (error) {
      this.logger.error({ err: error }, 'Failed to fetch AI upgrade plans');
      throw new AiError(
        503,
        'billingServiceUnavailable',
        'Billing is unavailable. Please retry later.'
      );
    }
  }

  async monthlyCredits(workspaceId: string): Promise<number> {
    try {
      const response = await firstValueFrom(
        this.client.getWorkspaceEntitlements({ workspaceId }).pipe(timeout(5000))
      );
      if (!response.entitlements || response.entitlements.workspaceId !== workspaceId) {
        throw new Error('Missing or mismatched billing entitlements');
      }
      const raw = response.entitlements.limits?.ai_credits_monthly;
      const credits = raw === undefined ? 0 : Number(raw);
      if (!Number.isSafeInteger(credits) || credits < -1) throw new Error('Invalid AI quota');
      return credits;
    } catch (error) {
      this.logger.error({ err: error }, 'Failed to fetch AI entitlements');
      throw new AiError(
        503,
        'billingServiceUnavailable',
        'Billing is unavailable. Please retry later.'
      );
    }
  }
}
