package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type LeaveWorkspace struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type LeaveWorkspaceHandler struct {
	authorizationSvc        AuthorizationSvc
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewLeaveWorkspaceHandler(
	authorizationSvc AuthorizationSvc,
	workspaceEventPublisher WorkspaceEventPublisher,
) *LeaveWorkspaceHandler {
	return &LeaveWorkspaceHandler{
		authorizationSvc:        authorizationSvc,
		workspaceEventPublisher: workspaceEventPublisher,
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
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.String("permission", WorkspaceItemPermissionWrite.String()),
	)
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionWrite)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to move items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}
	slog.DebugContext(
		ctx, "permission granted",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
	)
}
