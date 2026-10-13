package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) ListPlans(ctx context.Context, _ billing.ListPlansRequestObject) (billing.ListPlansResponseObject, error) {
	views, err := h.App.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out, err := toPlans(views)
	if err != nil {
		return nil, err
	}
	return billing.ListPlans200JSONResponse(out), nil
}
