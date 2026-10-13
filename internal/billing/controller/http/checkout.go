package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) CreateCheckoutSession(ctx context.Context, req billing.CreateCheckoutSessionRequestObject) (billing.CreateCheckoutSessionResponseObject, error) {
	payer, err := currentPayer(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.StartCheckout.Handle(ctx, &app.StartCheckout{WorkspaceID: req.WorkspaceId, PriceID: req.Body.PriceId, Payer: payer})
	if err != nil {
		return nil, err
	}
	return billing.CreateCheckoutSession200JSONResponse{Url: url}, nil
}

func (h *StrictHandler) CreatePortalSession(ctx context.Context, req billing.CreatePortalSessionRequestObject) (billing.CreatePortalSessionResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.StartPortal.Handle(ctx, &app.StartPortal{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	return billing.CreatePortalSession200JSONResponse{Url: url}, nil
}
