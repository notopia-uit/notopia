package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) ClaimPayer(ctx context.Context, req billing.ClaimPayerRequestObject) (billing.ClaimPayerResponseObject, error) {
	payer, err := currentPayer(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.App.Queries.ClaimPayer.Handle(ctx, &app.ClaimPayer{WorkspaceID: req.WorkspaceId, Payer: payer})
	if err != nil {
		return nil, err
	}
	return billing.ClaimPayer200JSONResponse{Url: url}, nil
}
