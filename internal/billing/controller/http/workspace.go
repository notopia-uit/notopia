package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) GetWorkspaceEntitlements(ctx context.Context, req billing.GetWorkspaceEntitlementsRequestObject) (billing.GetWorkspaceEntitlementsResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	ent, err := h.App.Queries.GetEntitlements.Handle(ctx, &app.GetEntitlements{WorkspaceID: req.WorkspaceId, UserID: u.ID, Internal: false})
	if err != nil {
		return nil, err
	}
	out, err := toEntitlements(ent)
	if err != nil {
		return nil, err
	}
	return billing.GetWorkspaceEntitlements200JSONResponse(out), nil
}

func (h *StrictHandler) GetWorkspaceBilling(ctx context.Context, req billing.GetWorkspaceBillingRequestObject) (billing.GetWorkspaceBillingResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	v, err := h.App.Queries.GetWorkspaceBilling.Handle(ctx, &app.GetWorkspaceBilling{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	out, err := toWorkspaceBilling(v)
	if err != nil {
		return nil, err
	}
	return billing.GetWorkspaceBilling200JSONResponse(out), nil
}
