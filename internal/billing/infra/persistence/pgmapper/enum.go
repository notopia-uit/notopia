package pgmapper

import (
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

func FromPlanTier(t domain.PlanTier) pgsqlc.PlanTier {
	switch t {
	case domain.PlanTierFree:
		return pgsqlc.PlanTierFree
	case domain.PlanTierPlus:
		return pgsqlc.PlanTierPlus
	case domain.PlanTierPro:
		return pgsqlc.PlanTierPro
	case domain.PlanTierUnspecified:
		return ""
	}
	return ""
}

func ToPlanTier(t pgsqlc.PlanTier) domain.PlanTier {
	switch t {
	case pgsqlc.PlanTierFree:
		return domain.PlanTierFree
	case pgsqlc.PlanTierPlus:
		return domain.PlanTierPlus
	case pgsqlc.PlanTierPro:
		return domain.PlanTierPro
	default:
		return domain.PlanTierUnspecified
	}
}

func FromInterval(i domain.Interval) pgsqlc.BillingInterval {
	switch i {
	case domain.IntervalMonth:
		return pgsqlc.BillingIntervalMonth
	case domain.IntervalYear:
		return pgsqlc.BillingIntervalYear
	case domain.IntervalLifetime:
		return pgsqlc.BillingIntervalLifetime
	case domain.IntervalUnspecified:
		return ""
	}
	return ""
}

func ToInterval(i pgsqlc.BillingInterval) domain.Interval {
	switch i {
	case pgsqlc.BillingIntervalMonth:
		return domain.IntervalMonth
	case pgsqlc.BillingIntervalYear:
		return domain.IntervalYear
	case pgsqlc.BillingIntervalLifetime:
		return domain.IntervalLifetime
	default:
		return domain.IntervalUnspecified
	}
}

func FromStatus(s domain.Status) pgsqlc.SubscriptionStatus {
	switch s {
	case domain.StatusIncomplete:
		return pgsqlc.SubscriptionStatusIncomplete
	case domain.StatusActive:
		return pgsqlc.SubscriptionStatusActive
	case domain.StatusPastDue:
		return pgsqlc.SubscriptionStatusPastDue
	case domain.StatusUnpaid:
		return pgsqlc.SubscriptionStatusUnpaid
	case domain.StatusCanceled:
		return pgsqlc.SubscriptionStatusCanceled
	case domain.StatusIncompleteExpired:
		return pgsqlc.SubscriptionStatusIncompleteExpired
	case domain.StatusNone:
		return ""
	}
	return ""
}

func ToStatus(s pgsqlc.SubscriptionStatus) domain.Status {
	switch s {
	case pgsqlc.SubscriptionStatusIncomplete:
		return domain.StatusIncomplete
	case pgsqlc.SubscriptionStatusActive:
		return domain.StatusActive
	case pgsqlc.SubscriptionStatusPastDue:
		return domain.StatusPastDue
	case pgsqlc.SubscriptionStatusUnpaid:
		return domain.StatusUnpaid
	case pgsqlc.SubscriptionStatusCanceled:
		return domain.StatusCanceled
	case pgsqlc.SubscriptionStatusIncompleteExpired:
		return domain.StatusIncompleteExpired
	default:
		return domain.StatusNone
	}
}

func FromRefundReason(r domain.RefundReason) pgsqlc.RefundReason {
	switch r {
	case domain.RefundReasonCancelNow:
		return pgsqlc.RefundReasonCancelNow
	case domain.RefundReasonLifetimeWindow:
		return pgsqlc.RefundReasonLifetimeWindow
	case domain.RefundReasonReplacedByLifetime:
		return pgsqlc.RefundReasonReplacedByLifetime
	case domain.RefundReasonWorkspaceDeleted:
		return pgsqlc.RefundReasonWorkspaceDeleted
	case domain.RefundReasonDuplicateCheckout:
		return pgsqlc.RefundReasonDuplicateCheckout
	case domain.RefundReasonExternal:
		return pgsqlc.RefundReasonExternal
	case domain.RefundReasonUnspecified:
		return ""
	}
	return ""
}

func ToRefundReason(r pgsqlc.RefundReason) domain.RefundReason {
	switch r {
	case pgsqlc.RefundReasonCancelNow:
		return domain.RefundReasonCancelNow
	case pgsqlc.RefundReasonLifetimeWindow:
		return domain.RefundReasonLifetimeWindow
	case pgsqlc.RefundReasonReplacedByLifetime:
		return domain.RefundReasonReplacedByLifetime
	case pgsqlc.RefundReasonWorkspaceDeleted:
		return domain.RefundReasonWorkspaceDeleted
	case pgsqlc.RefundReasonDuplicateCheckout:
		return domain.RefundReasonDuplicateCheckout
	case pgsqlc.RefundReasonExternal:
		return domain.RefundReasonExternal
	default:
		return domain.RefundReasonUnspecified
	}
}

func FromRefundStatus(s domain.RefundStatus) pgsqlc.RefundStatus {
	switch s {
	case domain.RefundStatusPending:
		return pgsqlc.RefundStatusPending
	case domain.RefundStatusSucceeded:
		return pgsqlc.RefundStatusSucceeded
	case domain.RefundStatusFailed:
		return pgsqlc.RefundStatusFailed
	case domain.RefundStatusCanceled:
		return pgsqlc.RefundStatusCanceled
	case domain.RefundStatusUnspecified:
		return ""
	}
	return ""
}

func ToRefundStatus(s pgsqlc.RefundStatus) domain.RefundStatus {
	switch s {
	case pgsqlc.RefundStatusPending:
		return domain.RefundStatusPending
	case pgsqlc.RefundStatusSucceeded:
		return domain.RefundStatusSucceeded
	case pgsqlc.RefundStatusFailed:
		return domain.RefundStatusFailed
	case pgsqlc.RefundStatusCanceled:
		return domain.RefundStatusCanceled
	default:
		return domain.RefundStatusUnspecified
	}
}
