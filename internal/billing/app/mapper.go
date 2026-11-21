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

func newSubscriptionStatus(s domain.SubscriptionStatus) SubscriptionStatus {
	switch s {
	case domain.SubscriptionStatusIncomplete:
		return SubscriptionStatusIncomplete
	case domain.SubscriptionStatusActive:
		return SubscriptionStatusActive
	case domain.SubscriptionStatusPastDue:
		return SubscriptionStatusPastDue
	case domain.SubscriptionStatusUnpaid:
		return SubscriptionStatusUnpaid
	case domain.SubscriptionStatusCanceled:
		return SubscriptionStatusCanceled
	case domain.SubscriptionStatusIncompleteExpired:
		return SubscriptionStatusIncompleteExpired
	case domain.SubscriptionStatusNone:
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

func newFeatures(m map[domain.FeatureKey]bool) Features {
	out := Features{}
	if v, ok := m[domain.FeatureKeyAutoCommit]; ok {
		out.AutoCommit = &v
	}
	if v, ok := m[domain.FeatureKeyWorkspaceGraph]; ok {
		out.WorkspaceGraph = &v
	}
	return out
}

func newLimits(m map[domain.LimitKey]int64) Limits {
	out := Limits{}
	if v, ok := m[domain.LimitKeyMembers]; ok {
		out.Members = &v
	}
	if v, ok := m[domain.LimitKeyPublishedNotes]; ok {
		out.PublishedNotes = &v
	}
	if v, ok := m[domain.LimitKeyNoteGraphMaxDepth]; ok {
		out.NoteGraphMaxDepth = &v
	}
	if v, ok := m[domain.LimitKeyAutoCommitMinInterval]; ok {
		out.AutoCommitMinIntervalSeconds = &v
	}
	if v, ok := m[domain.LimitKeyRevisionHistoryDays]; ok {
		out.RevisionHistoryDays = &v
	}
	if v, ok := m[domain.LimitKeyTrashRetentionDays]; ok {
		out.TrashRetentionDays = &v
	}
	if v, ok := m[domain.LimitKeyAICreditsMonthly]; ok {
		out.AICreditsMonthly = &v
	}
	return out
}

func newPlan(p *domain.Plan) Plan {
	return Plan{Tier: newPlanTier(p.Tier()), Features: newFeatures(p.Features()), Limits: newLimits(p.Limits())}
}

func newPrice(p *domain.Price) Price {
	return Price{
		ID: p.ID(), Plan: newPlanTier(p.Plan()), Interval: newBillingInterval(p.Interval()),
		Amount: p.Amount().Amount(), Currency: p.Amount().Currency().String(), LookupKey: p.ProviderLookupKey(), Active: p.Active(),
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
			ID: r.ID(), Amount: r.Amount().Amount(), Currency: r.Amount().Currency().String(),
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
	return Payer{ID: p.ID(), Email: p.Email(), Name: p.Name(), MissingSince: a.PayerMissingSince()}
}

func newRefundQuote(q domain.RefundQuote) RefundQuote {
	quoteLines := q.Lines()
	lines := make([]RefundQuoteLine, 0, len(quoteLines))
	for _, l := range quoteLines {
		lines = append(lines, RefundQuoteLine{PaidAmount: l.PaidAmount().Amount(), Amount: l.Amount().Amount(), CoverageSeconds: l.CoverageSeconds(), UnusedSeconds: l.UnusedSeconds()})
	}
	currency := ""
	if c := q.Amount().Currency(); c.IsValid() {
		currency = c.String()
	}
	return RefundQuote{Amount: q.Amount().Amount(), Currency: currency, UnusedSeconds: q.UnusedSeconds(), PeriodSeconds: q.PeriodSeconds(), Lines: lines}
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

func toDomainPayer(p Payer) (domain.Payer, error) {
	return domain.NewPayer(p.ID, p.Email, p.Name)
}

func toDomainFeatureUpdates(in Features) map[domain.FeatureKey]bool {
	out := map[domain.FeatureKey]bool{}
	if in.AutoCommit != nil {
		out[domain.FeatureKeyAutoCommit] = *in.AutoCommit
	}
	if in.WorkspaceGraph != nil {
		out[domain.FeatureKeyWorkspaceGraph] = *in.WorkspaceGraph
	}
	return out
}

func toDomainLimitUpdates(in Limits) map[domain.LimitKey]int64 {
	out := map[domain.LimitKey]int64{}
	if in.Members != nil {
		out[domain.LimitKeyMembers] = *in.Members
	}
	if in.PublishedNotes != nil {
		out[domain.LimitKeyPublishedNotes] = *in.PublishedNotes
	}
	if in.NoteGraphMaxDepth != nil {
		out[domain.LimitKeyNoteGraphMaxDepth] = *in.NoteGraphMaxDepth
	}
	if in.AutoCommitMinIntervalSeconds != nil {
		out[domain.LimitKeyAutoCommitMinInterval] = *in.AutoCommitMinIntervalSeconds
	}
	if in.RevisionHistoryDays != nil {
		out[domain.LimitKeyRevisionHistoryDays] = *in.RevisionHistoryDays
	}
	if in.TrashRetentionDays != nil {
		out[domain.LimitKeyTrashRetentionDays] = *in.TrashRetentionDays
	}
	if in.AICreditsMonthly != nil {
		out[domain.LimitKeyAICreditsMonthly] = *in.AICreditsMonthly
	}
	return out
}
