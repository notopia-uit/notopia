package app

import "github.com/notopia-uit/notopia/internal/billing/domain"

func newPlanTier(t domain.PlanTier) PlanTier {
	switch t {
	case domain.PlanTierFree:
		return PlanTierFree
	case domain.PlanTierPlus:
		return PlanTierPlus
	case domain.PlanTierPro:
		return PlanTierPro
	case domain.PlanTierUnspecified:
	}
	return PlanTierUnspecified
}

func newBillingInterval(i domain.Interval) BillingInterval {
	switch i {
	case domain.IntervalMonth:
		return BillingIntervalMonth
	case domain.IntervalYear:
		return BillingIntervalYear
	case domain.IntervalLifetime:
		return BillingIntervalLifetime
	case domain.IntervalUnspecified:
	}
	return BillingIntervalUnspecified
}

func newSubscriptionStatus(s domain.Status) SubscriptionStatus {
	switch s {
	case domain.StatusIncomplete:
		return SubscriptionStatusIncomplete
	case domain.StatusActive:
		return SubscriptionStatusActive
	case domain.StatusPastDue:
		return SubscriptionStatusPastDue
	case domain.StatusUnpaid:
		return SubscriptionStatusUnpaid
	case domain.StatusCanceled:
		return SubscriptionStatusCanceled
	case domain.StatusIncompleteExpired:
		return SubscriptionStatusIncompleteExpired
	case domain.StatusNone:
	}
	return SubscriptionStatusNone
}

func newRefundReason(r domain.RefundReason) RefundReason {
	switch r {
	case domain.RefundReasonCancelNow:
		return RefundReasonCancelNow
	case domain.RefundReasonLifetimeWindow:
		return RefundReasonLifetimeWindow
	case domain.RefundReasonReplacedByLifetime:
		return RefundReasonReplacedByLifetime
	case domain.RefundReasonWorkspaceDeleted:
		return RefundReasonWorkspaceDeleted
	case domain.RefundReasonDuplicateCheckout:
		return RefundReasonDuplicateCheckout
	case domain.RefundReasonExternal:
		return RefundReasonExternal
	case domain.RefundReasonUnspecified:
	}
	return RefundReasonUnspecified
}

func newRefundStatus(s domain.RefundStatus) RefundStatus {
	switch s {
	case domain.RefundStatusPending:
		return RefundStatusPending
	case domain.RefundStatusSucceeded:
		return RefundStatusSucceeded
	case domain.RefundStatusFailed:
		return RefundStatusFailed
	case domain.RefundStatusCanceled:
		return RefundStatusCanceled
	case domain.RefundStatusUnspecified:
	}
	return RefundStatusUnspecified
}

func newFeatures(m map[domain.FeatureKey]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return out
}

func newLimits(m map[domain.LimitKey]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return out
}

func newPlan(p *domain.Plan) Plan {
	return Plan{Tier: newPlanTier(p.Tier()), Features: newFeatures(p.Features()), Limits: newLimits(p.Limits())}
}

func newPrice(p *domain.Price) Price {
	return Price{
		ID: p.ID(), Plan: newPlanTier(p.Plan()), Interval: newBillingInterval(p.Interval()),
		Amount: p.Amount().Amount, Currency: p.Amount().Currency, LookupKey: p.LookupKey(), Active: p.Active(),
		CreatedAt: p.CreatedAt(), ArchivedAt: p.ArchivedAt(),
	}
}

func newPrices(ps []*domain.Price) []Price {
	out := make([]Price, 0, len(ps))
	for _, p := range ps {
		out = append(out, newPrice(p))
	}
	return out
}

func newBillingPolicy(p *domain.BillingPolicy) BillingPolicy {
	return BillingPolicy{PastDueGraceDays: p.PastDueGraceDays(), LifetimeRefundWindowDays: p.LifetimeRefundWindowDays()}
}

func newEntitlements(e domain.Entitlements) Entitlements {
	return Entitlements{
		WorkspaceID: e.WorkspaceID, Plan: newPlanTier(e.Plan), Status: newSubscriptionStatus(e.Status),
		AccessUntil: e.AccessUntil, Features: newFeatures(e.Features), Limits: newLimits(e.Limits),
	}
}

func newSubscription(s *domain.Subscription, policy *domain.BillingPolicy) Subscription {
	refunds := make([]Refund, 0, len(s.Refunds()))
	for _, r := range s.Refunds() {
		refunds = append(refunds, Refund{
			ID: r.ID(), Amount: r.Amount().Amount, Currency: r.Amount().Currency,
			Reason: newRefundReason(r.Reason()), Status: newRefundStatus(r.Status()), CreatedAt: r.CreatedAt(),
		})
	}
	return Subscription{
		ID: s.ID(), Plan: newPlanTier(s.Plan()), Interval: newBillingInterval(s.Interval()), Status: newSubscriptionStatus(s.Status()),
		CurrentPeriodStart: s.CurrentPeriodStart(), CurrentPeriodEnd: s.CurrentPeriodEnd(), CancelAtPeriodEnd: s.CancelAtPeriodEnd(),
		CanceledAt: s.CanceledAt(), AccessUntil: s.AccessUntil(policy), Refunds: refunds,
	}
}

func newPayer(a *domain.BillingAccount) Payer {
	p := a.Payer()
	return Payer{ID: p.ID, Email: p.Email, Name: p.Name, MissingSince: a.PayerMissingSince()}
}

func newRefundQuote(q domain.RefundQuote) RefundQuote {
	lines := make([]RefundQuoteLine, 0, len(q.Lines))
	for _, l := range q.Lines {
		lines = append(lines, RefundQuoteLine{PaidAmount: l.PaidAmount.Amount, Amount: l.Amount.Amount, CoverageSeconds: l.CoverageSeconds, UnusedSeconds: l.UnusedSeconds})
	}
	return RefundQuote{Amount: q.Amount.Amount, Currency: q.Amount.Currency, UnusedSeconds: q.UnusedSeconds, PeriodSeconds: q.PeriodSeconds, Lines: lines}
}

func toDomainPlanTier(t PlanTier) domain.PlanTier {
	switch t {
	case PlanTierFree:
		return domain.PlanTierFree
	case PlanTierPlus:
		return domain.PlanTierPlus
	case PlanTierPro:
		return domain.PlanTierPro
	case PlanTierUnspecified:
	}
	return domain.PlanTierUnspecified
}

func toDomainInterval(i BillingInterval) domain.Interval {
	switch i {
	case BillingIntervalMonth:
		return domain.IntervalMonth
	case BillingIntervalYear:
		return domain.IntervalYear
	case BillingIntervalLifetime:
		return domain.IntervalLifetime
	case BillingIntervalUnspecified:
	}
	return domain.IntervalUnspecified
}

func toDomainPayer(p Payer) domain.Payer {
	return domain.Payer{ID: p.ID, Email: p.Email, Name: p.Name}
}

// parseFeatures/parseLimits reject keys the domain does not know (errs.NewInvalid → 400).
func parseFeatures(in map[string]bool) (map[domain.FeatureKey]bool, error) {
	out := make(map[domain.FeatureKey]bool, len(in))
	for s, v := range in {
		k, err := domain.ParseFeatureKey(s)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

func parseLimits(in map[string]int64) (map[domain.LimitKey]int64, error) {
	out := make(map[domain.LimitKey]int64, len(in))
	for s, v := range in {
		k, err := domain.ParseLimitKey(s)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}
