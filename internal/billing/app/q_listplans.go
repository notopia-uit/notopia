package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ListPlans struct{}

type ListPlansHandler struct{ read BillingReadModel }

func NewListPlansHandler(read BillingReadModel) *ListPlansHandler {
	return &ListPlansHandler{read: read}
}

var ProvideListPlansHandler = NewListPlansHandler

type ListPlansQuery commonhandler.Query[ListPlans, []PlanView]

var _ ListPlansQuery = (*ListPlansHandler)(nil)

func (h *ListPlansHandler) Handle(ctx context.Context, _ *ListPlans) ([]PlanView, error) {
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return nil, err
	}
	prices, err := h.read.ActivePrices(ctx)
	if err != nil {
		return nil, err
	}
	var views []PlanView
	for _, tier := range domain.PlanTiers {
		p, ok := plans[tier]
		if !ok {
			continue
		}
		v := PlanView{Plan: newPlan(p), Prices: nil}
		for _, pr := range prices {
			if pr.Plan() == tier {
				v.Prices = append(v.Prices, newPrice(pr))
			}
		}
		views = append(views, v)
	}
	return views, nil
}
