package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type HandleWorkspaceDeleted struct {
	WorkspaceID uuid.UUID
}

type HandleWorkspaceDeletedHandler struct {
	canceller *Canceller
	settings  *Settings
}

func NewHandleWorkspaceDeletedHandler(canceller *Canceller, settings *Settings) *HandleWorkspaceDeletedHandler {
	return &HandleWorkspaceDeletedHandler{canceller: canceller, settings: settings}
}

var ProvideHandleWorkspaceDeletedHandler = NewHandleWorkspaceDeletedHandler

type HandleWorkspaceDeletedCmd commonhandler.Cmd[HandleWorkspaceDeleted]

var _ HandleWorkspaceDeletedCmd = (*HandleWorkspaceDeletedHandler)(nil)

func (h *HandleWorkspaceDeletedHandler) Handle(ctx context.Context, cmd *HandleWorkspaceDeleted) error {
	err := h.canceller.CancelNow(ctx, cmd.WorkspaceID, domain.RefundReasonWorkspaceDeleted, h.settings.ServiceAccountUserID, true)
	if hasCode(err, errs.CodeNoActiveSubscription) {
		return nil
	}
	return err
}
