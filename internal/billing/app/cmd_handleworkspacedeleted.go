package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type HandleWorkspaceDeleted struct {
	WorkspaceID uuid.UUID
}

type HandleWorkspaceDeletedHandler struct {
	subscriptionCanceller *SubscriptionCanceller
	settings              *Settings
}

func NewHandleWorkspaceDeletedHandler(subscriptionCanceller *SubscriptionCanceller, settings *Settings) *HandleWorkspaceDeletedHandler {
	return &HandleWorkspaceDeletedHandler{subscriptionCanceller: subscriptionCanceller, settings: settings}
}

var ProvideHandleWorkspaceDeletedHandler = NewHandleWorkspaceDeletedHandler

type HandleWorkspaceDeletedCmd commonhandler.Cmd[HandleWorkspaceDeleted]

var _ HandleWorkspaceDeletedCmd = (*HandleWorkspaceDeletedHandler)(nil)

func (h *HandleWorkspaceDeletedHandler) Handle(ctx context.Context, cmd *HandleWorkspaceDeleted) error {
	err := h.subscriptionCanceller.CancelNow(ctx, cmd.WorkspaceID, domain.RefundReasonWorkspaceDeleted, h.settings.ServiceAccountUserID, true)
	var errsErr errs.Error
	if errors.As(err, &errsErr) && errsErr.Code() == errs.CodeNoActiveSubscription {
		return nil
	}
	return err
}
