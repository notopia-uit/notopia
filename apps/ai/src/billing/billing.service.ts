import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  type BillingPlan,
  type BillingServiceClient,
  PlanTier,
} from '@notopia-uit/pb/billing';
import { firstValueFrom, timeout } from 'rxjs';

import { AiError } from '#/common/ai-error';

const UPGRADE_PLANS = [
  [PlanTier.PLAN_TIER_PLUS, 'plus'],
  [PlanTier.PLAN_TIER_PRO, 'pro'],
] as const;

@Injectable()
export class BillingService implements OnModuleInit {
  private client!: BillingServiceClient;
  private readonly logger = new Logger(BillingService.name);

  constructor(@Inject(BILLING_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit() {
    this.client = this.grpc.getService<BillingServiceClient>(BILLING_SERVICE_NAME);
  }

  async requiredPlan(currentLimit: number): Promise<string> {
    let plans: BillingPlan[];
    try {
      const response = await firstValueFrom(this.client.listPlans({}).pipe(timeout(5000)));
      plans = response.plans;
    } catch (error) {
      this.logger.error({ err: error }, 'Failed to fetch AI upgrade plans');
      throw new AiError(
        503,
        'billingServiceUnavailable',
        'Billing is unavailable. Please retry later.'
      );
    }

    try {
      for (const [tier, name] of UPGRADE_PLANS) {
        const plan = plans.find((candidate) => candidate.tier === tier);
        // A tier the billing service does not offer is not an upgrade target.
        if (!plan) continue;
        const quota = plan.limits?.ai_credits_monthly;
        if (quota === undefined || !Number.isSafeInteger(quota) || quota < -1) {
          throw new Error(`Plan "${name}" has an invalid ai_credits_monthly limit`);
        }
        if (quota === -1 || quota > currentLimit) return name;
      }
      throw new Error('No plan grants more AI credits than the current limit');
    } catch (error) {
      this.logger.error(
        { err: error, currentLimit, plans },
        'Cannot determine a required AI upgrade plan'
      );
      throw new AiError(
        503,
        'billingServiceUnavailable',
        'No AI upgrade plan is available for this workspace.'
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
