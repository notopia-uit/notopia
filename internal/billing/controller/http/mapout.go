package http

import (
	"fmt"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/oapi-codegen/nullable"
)

func nullableTime(t *time.Time) nullable.Nullable[time.Time] {
	if t == nil {
		return nullable.NewNullNullable[time.Time]()
	}
	return nullable.NewNullableWithValue(*t)
}

func toPlanTierDTO(t app.PlanTier) (billing.PlanTier, error) {
	switch t {
	case app.PlanTierFree:
		return billing.PlanTierFree, nil
	case app.PlanTierPlus:
		return billing.PlanTierPlus, nil
	case app.PlanTierPro:
		return billing.PlanTierPro, nil
	case app.PlanTierUnspecified:
	}
	return "", errs.NewInternal(fmt.Sprintf("invalid plan tier: %v", t), nil)
}

func toBillingIntervalDTO(i app.BillingInterval) (billing.BillingInterval, error) {
	switch i {
	case app.BillingIntervalMonth:
		return billing.BillingIntervalMonth, nil
	case app.BillingIntervalYear:
		return billing.BillingIntervalYear, nil
	case app.BillingIntervalLifetime:
		return billing.BillingIntervalLifetime, nil
	case app.BillingIntervalUnspecified:
	}
	return "", errs.NewInternal(fmt.Sprintf("invalid billing interval: %v", i), nil)
}

func toSubscriptionStatusDTO(s app.SubscriptionStatus) (billing.SubscriptionStatus, error) {
	switch s {
	case app.SubscriptionStatusNone:
		return billing.SubscriptionStatusNone, nil
	case app.SubscriptionStatusIncomplete:
		return billing.SubscriptionStatusIncomplete, nil
	case app.SubscriptionStatusActive:
		return billing.SubscriptionStatusActive, nil
	case app.SubscriptionStatusPastDue:
		return billing.SubscriptionStatusPastDue, nil
	case app.SubscriptionStatusUnpaid:
		return billing.SubscriptionStatusUnpaid, nil
	case app.SubscriptionStatusCanceled:
		return billing.SubscriptionStatusCanceled, nil
	case app.SubscriptionStatusIncompleteExpired:
		return billing.SubscriptionStatusIncompleteExpired, nil
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid subscription status: %v", s), nil)
	}
}

func toRefundReasonDTO(r app.RefundReason) (billing.RefundReason, error) {
	switch r {
	case app.RefundReasonCancelNow:
		return billing.RefundReasonCancelNow, nil
	case app.RefundReasonLifetimeWindow:
		return billing.RefundReasonLifetimeWindow, nil
	case app.RefundReasonReplacedByLifetime:
		return billing.RefundReasonReplacedByLifetime, nil
	case app.RefundReasonWorkspaceDeleted:
		return billing.RefundReasonWorkspaceDeleted, nil
	case app.RefundReasonDuplicateCheckout:
		return billing.RefundReasonDuplicateCheckout, nil
	case app.RefundReasonExternal:
		return billing.RefundReasonExternal, nil
	case app.RefundReasonUnspecified:
	}
	return "", errs.NewInternal(fmt.Sprintf("invalid refund reason: %v", r), nil)
}

func toRefundStatusDTO(s app.RefundStatus) (billing.RefundStatus, error) {
	switch s {
	case app.RefundStatusPending:
		return billing.RefundStatusPending, nil
	case app.RefundStatusSucceeded:
		return billing.RefundStatusSucceeded, nil
	case app.RefundStatusFailed:
		return billing.RefundStatusFailed, nil
	case app.RefundStatusCanceled:
		return billing.RefundStatusCanceled, nil
	case app.RefundStatusUnspecified:
	}
	return "", errs.NewInternal(fmt.Sprintf("invalid refund status: %v", s), nil)
}

func toFeatures(m map[string]bool) billing.Features {
	out := billing.Features{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func toLimits(m map[string]int64) billing.Limits {
	out := billing.Limits{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func toEntitlements(e app.Entitlements) (billing.Entitlements, error) {
	plan, err := toPlanTierDTO(e.Plan)
	if err != nil {
		return billing.Entitlements{}, err
	}
	status, err := toSubscriptionStatusDTO(e.Status)
	if err != nil {
		return billing.Entitlements{}, err
	}
	return billing.Entitlements{
		WorkspaceId: e.WorkspaceID, Plan: plan, Status: status,
		AccessUntil: nullableTime(e.AccessUntil), Features: toFeatures(e.Features), Limits: toLimits(e.Limits),
	}, nil
}

func toPrice(p app.Price) (billing.Price, error) {
	plan, err := toPlanTierDTO(p.Plan)
	if err != nil {
		return billing.Price{}, err
	}
	interval, err := toBillingIntervalDTO(p.Interval)
	if err != nil {
		return billing.Price{}, err
	}
	return billing.Price{
		Id: p.ID, Plan: plan, Interval: interval,
		Amount: p.Amount, Currency: p.Currency, LookupKey: p.LookupKey, Active: p.Active,
		CreatedAt: p.CreatedAt, ArchivedAt: nullableTime(p.ArchivedAt),
	}, nil
}

func toPrices(ps []app.Price) ([]billing.Price, error) {
	out := make([]billing.Price, 0, len(ps))
	for _, p := range ps {
		dto, err := toPrice(p)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func toPlan(v app.PlanView) (billing.Plan, error) {
	tier, err := toPlanTierDTO(v.Plan.Tier)
	if err != nil {
		return billing.Plan{}, err
	}
	prices, err := toPrices(v.Prices)
	if err != nil {
		return billing.Plan{}, err
	}
	return billing.Plan{Tier: tier, Features: toFeatures(v.Plan.Features), Limits: toLimits(v.Plan.Limits), Prices: prices}, nil
}

func toPlans(views []app.PlanView) ([]billing.Plan, error) {
	out := make([]billing.Plan, 0, len(views))
	for _, v := range views {
		dto, err := toPlan(v)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func toRefund(r app.Refund) (billing.Refund, error) {
	reason, err := toRefundReasonDTO(r.Reason)
	if err != nil {
		return billing.Refund{}, err
	}
	status, err := toRefundStatusDTO(r.Status)
	if err != nil {
		return billing.Refund{}, err
	}
	return billing.Refund{
		Id: r.ID, Amount: r.Amount, Currency: r.Currency,
		Reason: reason, Status: status, CreatedAt: r.CreatedAt,
	}, nil
}

func toSubscription(s app.Subscription) (billing.Subscription, error) {
	plan, err := toPlanTierDTO(s.Plan)
	if err != nil {
		return billing.Subscription{}, err
	}
	interval, err := toBillingIntervalDTO(s.Interval)
	if err != nil {
		return billing.Subscription{}, err
	}
	status, err := toSubscriptionStatusDTO(s.Status)
	if err != nil {
		return billing.Subscription{}, err
	}
	refunds := make([]billing.Refund, 0, len(s.Refunds))
	for _, r := range s.Refunds {
		dto, err := toRefund(r)
		if err != nil {
			return billing.Subscription{}, err
		}
		refunds = append(refunds, dto)
	}
	return billing.Subscription{
		Id: s.ID, Plan: plan, Interval: interval, Status: status,
		CurrentPeriodStart: nullableTime(s.CurrentPeriodStart), CurrentPeriodEnd: nullableTime(s.CurrentPeriodEnd),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd, CanceledAt: nullableTime(s.CanceledAt),
		AccessUntil: nullableTime(s.AccessUntil), Refunds: refunds,
	}, nil
}

func toWorkspaceBilling(v app.WorkspaceBillingView) (billing.WorkspaceBilling, error) {
	out := billing.WorkspaceBilling{}
	if v.Payer != nil {
		p := v.Payer
		out.Payer = nullable.NewNullableWithValue(billing.Payer{Id: p.ID, Email: p.Email, Name: p.Name, MissingSince: nullableTime(p.MissingSince)})
	}
	if v.Subscription != nil {
		sub, err := toSubscription(*v.Subscription)
		if err != nil {
			return billing.WorkspaceBilling{}, err
		}
		out.Subscription = nullable.NewNullableWithValue(sub)
	}
	return out, nil
}
