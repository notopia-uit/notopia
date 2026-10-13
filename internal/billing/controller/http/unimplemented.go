package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
)

func (h *StrictHandler) AdminListPlans(ctx context.Context, _ billing.AdminListPlansRequestObject) (billing.AdminListPlansResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminUpdatePlan(ctx context.Context, _ billing.AdminUpdatePlanRequestObject) (billing.AdminUpdatePlanResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminGetPolicy(ctx context.Context, _ billing.AdminGetPolicyRequestObject) (billing.AdminGetPolicyResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminUpdatePolicy(ctx context.Context, _ billing.AdminUpdatePolicyRequestObject) (billing.AdminUpdatePolicyResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminListPrices(ctx context.Context, _ billing.AdminListPricesRequestObject) (billing.AdminListPricesResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminCreatePrice(ctx context.Context, _ billing.AdminCreatePriceRequestObject) (billing.AdminCreatePriceResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) AdminArchivePrice(ctx context.Context, _ billing.AdminArchivePriceRequestObject) (billing.AdminArchivePriceResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) StripeWebhook(ctx context.Context, _ billing.StripeWebhookRequestObject) (billing.StripeWebhookResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) CreateCheckoutSession(ctx context.Context, _ billing.CreateCheckoutSessionRequestObject) (billing.CreateCheckoutSessionResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) ClaimPayer(ctx context.Context, _ billing.ClaimPayerRequestObject) (billing.ClaimPayerResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) CreatePortalSession(ctx context.Context, _ billing.CreatePortalSessionRequestObject) (billing.CreatePortalSessionResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) CancelNow(ctx context.Context, _ billing.CancelNowRequestObject) (billing.CancelNowResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}

func (h *StrictHandler) PreviewCancelNow(ctx context.Context, _ billing.PreviewCancelNowRequestObject) (billing.PreviewCancelNowResponseObject, error) {
	return nil, errs.New(errs.CodeInternal, "not implemented", nil)
}
