package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CancelNow struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type CancelNowHandler struct {
	authz     AuthorizationSvc
	canceller *Canceller
}

func NewCancelNowHandler(authz AuthorizationSvc, canceller *Canceller) *CancelNowHandler {
	return &CancelNowHandler{authz: authz, canceller: canceller}
}

var ProvideCancelNowHandler = NewCancelNowHandler

type CancelNowCmd commonhandler.Cmd[CancelNow]

var _ CancelNowCmd = (*CancelNowHandler)(nil)

func (h *CancelNowHandler) Handle(ctx context.Context, cmd *CancelNow) error {
	if err := requirePermission(ctx, h.authz, cmd.UserID, cmd.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return err
	}
	sub, _, err := h.canceller.load(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	reason := domain.RefundReasonCancelNow
	if sub.Interval() == domain.IntervalLifetime {
		reason = domain.RefundReasonLifetimeWindow
	}
	return h.canceller.CancelNow(ctx, cmd.WorkspaceID, reason, cmd.UserID, false)
}
