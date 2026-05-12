package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type LeaveWorkspace struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type LeaveWorkspaceHandler struct {
	authorizationSvc AuthorizationSvc
}

func NewLeaveWorkspaceHandler(
	authorizationSvc AuthorizationSvc,
) *LeaveWorkspaceHandler {
	return &LeaveWorkspaceHandler{
		authorizationSvc: authorizationSvc,
	}
}

var ProvideLeaveWorkspaceHandler = NewLeaveWorkspaceHandler

func (h *LeaveWorkspaceHandler) Handle(ctx context.Context, cmd *LeaveWorkspace) error {
	slog.DebugContext(
		ctx, "leaving workspace",
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.String("user_id", cmd.UserID),
	)
	members, err := h.authorizationSvc.GetWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if len(members) == 1 {
		return errs.NewCannotLeaveWorkspaceWithOnlyOneMember(cmd.WorkspaceID)
	}
	return h.authorizationSvc.LeaveWorkspace(ctx, cmd.UserID, cmd.WorkspaceID)
}
