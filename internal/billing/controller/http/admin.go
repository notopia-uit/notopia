package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) AdminListPlans(
	ctx context.Context,
	_ billing.AdminListPlansRequestObject,
) (billing.AdminListPlansResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	views, err := h.App.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out, err := toPlans(views)
	if err != nil {
		return nil, err
	}
	return billing.AdminListPlans200JSONResponse(out), nil
}

func (h *StrictHandler) AdminUpdatePlan(
	ctx context.Context,
	req billing.AdminUpdatePlanRequestObject,
) (billing.AdminUpdatePlanResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	tier, err := toPlanTier(req.Tier)
	if err != nil {
		return nil, err
	}
	cmd := &app.UpdatePlan{
		Tier:     tier,
		Features: fromFeatures(req.Body.Features),
		Limits:   fromLimits(req.Body.Limits),
		UserID:   u.ID,
	}
	if err := h.App.Cmds.UpdatePlan.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	v, err := h.App.Queries.GetPlan.Handle(ctx, &app.GetPlan{Tier: tier})
	if err != nil {
		return nil, err
	}
	out, err := toPlan(v)
	if err != nil {
		return nil, err
	}
	return billing.AdminUpdatePlan200JSONResponse(out), nil
}

func (h *StrictHandler) AdminGetPolicy(
	ctx context.Context,
	_ billing.AdminGetPolicyRequestObject,
) (billing.AdminGetPolicyResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	p, err := h.App.Queries.GetPolicy.Handle(ctx, &app.GetPolicy{})
	if err != nil {
		return nil, err
	}
	return billing.AdminGetPolicy200JSONResponse(toPolicy(p)), nil
}

func (h *StrictHandler) AdminUpdatePolicy(
	ctx context.Context,
	req billing.AdminUpdatePolicyRequestObject,
) (billing.AdminUpdatePolicyResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	cmd := &app.UpdatePolicy{
		PastDueGraceDays:         req.Body.PastDueGraceDays,
		LifetimeRefundWindowDays: req.Body.LifetimeRefundWindowDays,
		UserID:                   u.ID,
	}
	if err := h.App.Cmds.UpdatePolicy.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	p, err := h.App.Queries.GetPolicy.Handle(ctx, &app.GetPolicy{})
	if err != nil {
		return nil, err
	}
	return billing.AdminUpdatePolicy200JSONResponse(toPolicy(p)), nil
}

func (h *StrictHandler) AdminListPrices(
	ctx context.Context,
	_ billing.AdminListPricesRequestObject,
) (billing.AdminListPricesResponseObject, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	prices, err := h.App.Queries.ListPrices.Handle(ctx, &app.ListPrices{})
	if err != nil {
		return nil, err
	}
	out, err := toPrices(prices)
	if err != nil {
		return nil, err
	}
	return billing.AdminListPrices200JSONResponse(out), nil
}

func (h *StrictHandler) AdminCreatePrice(
	ctx context.Context,
	req billing.AdminCreatePriceRequestObject,
) (billing.AdminCreatePriceResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := toCreatePricePlan(req.Body.Plan)
	if err != nil {
		return nil, err
	}
	interval, err := toInterval(req.Body.Interval)
	if err != nil {
		return nil, err
	}
	cmd := &app.CreatePrice{
		ID:       uuid.Must(uuid.NewV7()),
		Plan:     plan,
		Interval: interval,
		Amount:   req.Body.Amount,
		Currency: req.Body.Currency,
		UserID:   u.ID,
	}
	if err := h.App.Cmds.CreatePrice.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return h.adminPriceResponse(ctx, cmd.ID, func(p billing.Price) billing.AdminCreatePriceResponseObject {
		return billing.AdminCreatePrice201JSONResponse(p)
	})
}

func (h *StrictHandler) AdminArchivePrice(
	ctx context.Context,
	req billing.AdminArchivePriceRequestObject,
) (billing.AdminArchivePriceResponseObject, error) {
	u, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	cmd := &app.ArchivePrice{
		PriceID: req.PriceId,
		UserID:  u.ID,
	}
	if err := h.App.Cmds.ArchivePrice.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return h.adminPriceResponse(ctx, req.PriceId, func(p billing.Price) billing.AdminArchivePriceResponseObject {
		return billing.AdminArchivePrice200JSONResponse(p)
	})
}

func (h *StrictHandler) adminPriceResponse[R any](ctx context.Context, id uuid.UUID, respond func(billing.Price) R) (R, error) {
	var zero R
	price, err := h.App.Queries.GetPrice.Handle(ctx, &app.GetPrice{ID: id})
	if err != nil {
		return zero, err
	}
	out, err := toPrice(price)
	if err != nil {
		return zero, err
	}
	return respond(out), nil
}
