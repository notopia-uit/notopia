package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CancelNow struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type CancelNowHandler struct {
	authz                 AuthorizationSvc
	subscriptionCanceller *SubscriptionCanceller
}

func NewCancelNowHandler(authz AuthorizationSvc, subscriptionCanceller *SubscriptionCanceller) *CancelNowHandler {
	return &CancelNowHandler{authz: authz, subscriptionCanceller: subscriptionCanceller}
}

var ProvideCancelNowHandler = NewCancelNowHandler

type CancelNowCmd commonhandler.Cmd[CancelNow]

var _ CancelNowCmd = (*CancelNowHandler)(nil)

func (h *CancelNowHandler) Handle(ctx context.Context, cmd *CancelNow) error {
	if err := requirePermission(ctx, h.authz, cmd.UserID, cmd.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return err
	}
	return h.subscriptionCanceller.CancelNow(ctx, cmd.WorkspaceID, CancelNowCauseUser, cmd.UserID)
}
