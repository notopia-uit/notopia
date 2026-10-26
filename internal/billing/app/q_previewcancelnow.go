package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PreviewCancelNow struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type PreviewCancelNowHandler struct {
	authz     AuthorizationSvc
	canceller *Canceller
}

func NewPreviewCancelNowHandler(authz AuthorizationSvc, canceller *Canceller) *PreviewCancelNowHandler {
	return &PreviewCancelNowHandler{authz: authz, canceller: canceller}
}

var ProvidePreviewCancelNowHandler = NewPreviewCancelNowHandler

type PreviewCancelNowQuery commonhandler.Query[PreviewCancelNow, RefundQuote]

var _ PreviewCancelNowQuery = (*PreviewCancelNowHandler)(nil)

func (h *PreviewCancelNowHandler) Handle(ctx context.Context, q *PreviewCancelNow) (RefundQuote, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return RefundQuote{}, err
	}
	quote, err := h.canceller.Quote(ctx, q.WorkspaceID)
	if err != nil {
		return RefundQuote{}, err
	}
	return newRefundQuote(quote), nil
}
