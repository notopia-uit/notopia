import { status } from '@grpc/grpc-js';
import { Inject, Injectable, InternalServerErrorException, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  type BillingServiceClient,
  type Entitlements as EntitlementsPb,
  type Features as FeaturesPb,
  type Limits as LimitsPb,
  PlanTier as PlanTierPb,
} from '@notopia-uit/pb/billing';
import { Traceable } from 'nestjs-otel';
import { Observable, firstValueFrom } from 'rxjs';

import { isGrpcError } from '#/common';

import { BillingUnavailableException } from './billing-unavailable.exception';
import { type Entitlements, type PlanEntitlements, PlanFeature, PlanLimit, PlanTier } from './plan';

const unavailableCodes: number[] = [status.UNAVAILABLE, status.DEADLINE_EXCEEDED];

export const toPlanTier = (tier: PlanTierPb): PlanTier => {
  switch (tier) {
    case PlanTierPb.PLAN_TIER_FREE:
      return PlanTier.Free;
    case PlanTierPb.PLAN_TIER_PLUS:
      return PlanTier.Plus;
    case PlanTierPb.PLAN_TIER_PRO:
      return PlanTier.Pro;
    default:
      throw new InternalServerErrorException(`Invalid plan tier from billing: ${tier}`);
  }
};

const toNumber = (value: unknown): number => {
  if (typeof value === 'object' && value !== null && 'toNumber' in value) {
    return (value as { toNumber(): number }).toNumber();
  }
  return Number(value);
};

export const toFeatures = (raw: FeaturesPb | undefined): Partial<Record<PlanFeature, boolean>> => {
  const out: Partial<Record<PlanFeature, boolean>> = {};
  if (raw?.autoCommit !== undefined) out[PlanFeature.AutoCommit] = raw.autoCommit;
  return out;
};

export const toLimits = (raw: LimitsPb | undefined): Partial<Record<PlanLimit, number>> => {
  const out: Partial<Record<PlanLimit, number>> = {};
  if (raw?.autoCommitMinIntervalSeconds !== undefined) {
    out[PlanLimit.AutoCommitMinIntervalSeconds] = toNumber(raw.autoCommitMinIntervalSeconds);
  }
  if (raw?.revisionHistoryDays !== undefined) {
    out[PlanLimit.RevisionHistoryDays] = toNumber(raw.revisionHistoryDays);
  }
  return out;
};

export const toEntitlements = (e: EntitlementsPb | undefined): Entitlements => {
  if (!e) {
    throw new InternalServerErrorException('Missing entitlements from billing');
  }
  return { plan: toPlanTier(e.plan), features: toFeatures(e.features), limits: toLimits(e.limits) };
};

@Injectable()
@Traceable()
export class BillingService implements OnModuleInit {
  private static readonly batchGetEntitlementsLimit = 500;
  private billingServiceClient!: BillingServiceClient;

  constructor(@Inject(BILLING_PACKAGE_NAME) private readonly client: ClientGrpc) {}

  onModuleInit(): void {
    this.billingServiceClient = this.client.getService<BillingServiceClient>(BILLING_SERVICE_NAME);
  }

  private async call<T>(request: Observable<T>): Promise<T> {
    try {
      return await firstValueFrom(request);
    } catch (error) {
      if (isGrpcError(error) && error.code !== undefined && unavailableCodes.includes(error.code)) {
        throw new BillingUnavailableException(error);
      }
      throw error;
    }
  }

  async getEntitlements(workspaceId: string): Promise<Entitlements> {
    const response = await this.call(
      this.billingServiceClient.getWorkspaceEntitlements({ workspaceId })
    );
    return toEntitlements(response.entitlements);
  }

  async batchGetEntitlements(workspaceIds: string[]): Promise<Map<string, Entitlements>> {
    const out = new Map<string, Entitlements>();
    const chunkSize = BillingService.batchGetEntitlementsLimit;
    for (let i = 0; i < workspaceIds.length; i += chunkSize) {
      const response = await this.call(
        this.billingServiceClient.batchGetWorkspaceEntitlements({
          workspaceIds: workspaceIds.slice(i, i + chunkSize),
        })
      );
      for (const [id, entitlements] of Object.entries(response.entitlements)) {
        out.set(id, toEntitlements(entitlements));
      }
    }
    return out;
  }

  async listPlans(): Promise<PlanEntitlements[]> {
    const response = await this.call(this.billingServiceClient.listPlans({}));
    return response.plans.map((plan) => ({
      tier: toPlanTier(plan.tier),
      features: toFeatures(plan.features),
      limits: toLimits(plan.limits),
    }));
  }
}
