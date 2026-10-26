package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) PreviewCancelNow(ctx context.Context, req billing.PreviewCancelNowRequestObject) (billing.PreviewCancelNowResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	q, err := h.App.Queries.PreviewCancelNow.Handle(ctx, &app.PreviewCancelNow{WorkspaceID: req.WorkspaceId, UserID: u.ID})
	if err != nil {
		return nil, err
	}
	payments := make([]billing.RefundQuotePayment, 0, len(q.Lines))
	for _, l := range q.Lines {
		payments = append(payments, billing.RefundQuotePayment{
			PaidAmount:      l.PaidAmount,
			Amount:          l.Amount,
			CoverageSeconds: l.CoverageSeconds,
			UnusedSeconds:   l.UnusedSeconds,
		})
	}
	return billing.PreviewCancelNow200JSONResponse{
		Amount:        q.Amount,
		Currency:      q.Currency,
		UnusedSeconds: q.UnusedSeconds,
		PeriodSeconds: q.PeriodSeconds,
		Payments:      payments,
	}, nil
}

func (h *StrictHandler) CancelNow(ctx context.Context, req billing.CancelNowRequestObject) (billing.CancelNowResponseObject, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.App.Cmds.CancelNow.Handle(ctx, &app.CancelNow{WorkspaceID: req.WorkspaceId, UserID: u.ID}); err != nil {
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
	return billing.CancelNow200JSONResponse(out), nil
}
