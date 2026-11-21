package http

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func toPlanTier(t billing.PlanTier) (app.PlanTier, error) {
	switch t {
	case billing.PlanTierFree:
		return app.PlanTierFree, nil
	case billing.PlanTierPlus:
		return app.PlanTierPlus, nil
	case billing.PlanTierPro:
		return app.PlanTierPro, nil
	}
	return app.PlanTierUnspecified, errs.NewInvalid(fmt.Sprintf("invalid plan tier: %q", t))
}

func toCreatePricePlan(p billing.CreatePriceRequestPlan) (app.PlanTier, error) {
	switch p {
	case billing.CreatePriceRequestPlanPlus:
		return app.PlanTierPlus, nil
	case billing.CreatePriceRequestPlanPro:
		return app.PlanTierPro, nil
	}
	return app.PlanTierUnspecified, errs.NewInvalid(fmt.Sprintf("invalid price plan: %q", p))
}

func toInterval(i billing.BillingInterval) (app.BillingInterval, error) {
	switch i {
	case billing.BillingIntervalMonth:
		return app.BillingIntervalMonth, nil
	case billing.BillingIntervalYear:
		return app.BillingIntervalYear, nil
	case billing.BillingIntervalLifetime:
		return app.BillingIntervalLifetime, nil
	}
	return app.BillingIntervalUnspecified, errs.NewInvalid(fmt.Sprintf("invalid billing interval: %q", i))
}

func fromFeatures(in *billing.Features) app.Features {
	if in == nil {
		return app.Features{}
	}
	return app.Features{AutoCommit: in.AutoCommit, WorkspaceGraph: in.WorkspaceGraph}
}

func fromLimits(in *billing.Limits) app.Limits {
	if in == nil {
		return app.Limits{}
	}
	return app.Limits{
		Members: in.Members, PublishedNotes: in.PublishedNotes,
		NoteGraphMaxDepth: in.NoteGraphMaxDepth, AutoCommitMinIntervalSeconds: in.AutoCommitMinIntervalSeconds,
		RevisionHistoryDays: in.RevisionHistoryDays, TrashRetentionDays: in.TrashRetentionDays,
		AICreditsMonthly: in.AiCreditsMonthly,
	}
}
