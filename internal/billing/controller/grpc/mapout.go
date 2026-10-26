package grpc

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toPBPlanTier(t app.PlanTier) (pb.PlanTier, error) {
	switch t {
	case app.PlanTierFree:
		return pb.PlanTier_PLAN_TIER_FREE, nil
	case app.PlanTierPlus:
		return pb.PlanTier_PLAN_TIER_PLUS, nil
	case app.PlanTierPro:
		return pb.PlanTier_PLAN_TIER_PRO, nil
	case app.PlanTierUnspecified:
	}
	return pb.PlanTier_PLAN_TIER_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid plan tier: %v", t), nil)
}

func toPBStatus(s app.SubscriptionStatus) (pb.SubscriptionStatus, error) {
	switch s {
	case app.SubscriptionStatusNone:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_NONE, nil
	case app.SubscriptionStatusIncomplete:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE, nil
	case app.SubscriptionStatusActive:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE, nil
	case app.SubscriptionStatusPastDue:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE, nil
	case app.SubscriptionStatusUnpaid:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_UNPAID, nil
	case app.SubscriptionStatusCanceled:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED, nil
	case app.SubscriptionStatusIncompleteExpired:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE_EXPIRED, nil
	default:
		return pb.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid subscription status: %v", s), nil)
	}
}

func toPBEntitlements(e app.Entitlements) (*pb.Entitlements, error) {
	plan, err := toPBPlanTier(e.Plan)
	if err != nil {
		return nil, err
	}
	status, err := toPBStatus(e.Status)
	if err != nil {
		return nil, err
	}
	out := &pb.Entitlements{
		WorkspaceId: e.WorkspaceID.String(), Plan: plan, Status: status,
		Features: e.Features, Limits: e.Limits, AccessUntil: nil,
	}
	if e.AccessUntil != nil {
		out.AccessUntil = timestamppb.New(*e.AccessUntil)
	}
	return out, nil
}
