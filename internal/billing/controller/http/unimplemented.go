package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) ClaimPayer(ctx context.Context, _ billing.ClaimPayerRequestObject) (billing.ClaimPayerResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) CancelNow(ctx context.Context, _ billing.CancelNowRequestObject) (billing.CancelNowResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) PreviewCancelNow(ctx context.Context, _ billing.PreviewCancelNowRequestObject) (billing.PreviewCancelNowResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}
