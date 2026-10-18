export const PlanTier = { Free: 'free', Plus: 'plus', Pro: 'pro' } as const;
export type PlanTier = (typeof PlanTier)[keyof typeof PlanTier];

export const PlanFeature = { AutoCommit: 'auto_commit' } as const;
export type PlanFeature = (typeof PlanFeature)[keyof typeof PlanFeature];

export const PlanLimit = {
  AutoCommitMinIntervalSeconds: 'auto_commit_min_interval_seconds',
  RevisionHistoryDays: 'revision_history_days',
} as const;
export type PlanLimit = (typeof PlanLimit)[keyof typeof PlanLimit];

export const PLAN_QUOTA_UNLIMITED = -1;

export interface Entitlements {
  plan: PlanTier;
  features: Partial<Record<PlanFeature, boolean>>;
  limits: Partial<Record<PlanLimit, number>>;
}

export interface PlanEntitlements {
  tier: PlanTier;
  features: Partial<Record<PlanFeature, boolean>>;
  limits: Partial<Record<PlanLimit, number>>;
}

export const feature = (e: Pick<Entitlements, 'features'>, f: PlanFeature) =>
  e.features[f] ?? false;

export const limit = (e: Pick<Entitlements, 'limits'>, l: PlanLimit) => e.limits[l] ?? 0;

export const isUnlimited = (quota: number) => quota === PLAN_QUOTA_UNLIMITED;
