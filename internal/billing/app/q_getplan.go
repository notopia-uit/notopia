package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetPlan struct {
	Tier PlanTier
}

type GetPlanHandler struct{ read BillingReadModel }

func NewGetPlanHandler(read BillingReadModel) *GetPlanHandler {
	return &GetPlanHandler{read: read}
}

var ProvideGetPlanHandler = NewGetPlanHandler

type GetPlanQuery commonhandler.Query[GetPlan, PlanView]

var _ GetPlanQuery = (*GetPlanHandler)(nil)

func (h *GetPlanHandler) Handle(ctx context.Context, q *GetPlan) (PlanView, error) {
	tier := toDomainPlanTier(q.Tier)
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return PlanView{}, err
	}
	p, ok := plans[tier]
	if !ok {
		return PlanView{}, errs.NewNotFound("plan not found")
	}
	prices, err := h.read.ActivePrices(ctx)
	if err != nil {
		return PlanView{}, err
	}
	v := PlanView{Plan: newPlan(p), Prices: nil}
	for _, pr := range prices {
		if pr.Plan() == tier {
			v.Prices = append(v.Prices, newPrice(pr))
		}
	}
	return v, nil
}
