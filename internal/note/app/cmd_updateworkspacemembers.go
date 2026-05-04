package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type UpdateWorkspaceMembers struct {
	WorkspaceID uuid.UUID
	Members     []WorkspaceMemberUpdate
	UserID      string
}

type UpdateWorkspaceMembersHandler struct {
	workspaceEventPublisher WorkspaceEventPublisher
	authorizationSvc        AuthorizationSvc
}

func NewUpdateWorkspaceMembersHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
	authorizationSvc AuthorizationSvc,
) *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{
		workspaceEventPublisher: workspaceEventPublisher,
		authorizationSvc:        authorizationSvc,
	}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

// FIXME: This maybe need saga? Because it involves 2 external things
// Or we have to persist event and let another handler to consume
func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, cmd *UpdateWorkspaceMembers) error {
	slog.DebugContext(
		ctx, "updating workspace members",
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.Int("member_count", len(cmd.Members)),
		slog.String("user_id", cmd.UserID),
	)
	if len(cmd.Members) == 0 {
		return errs.NewWorkspaceMembersCannotBeEmpty(cmd.WorkspaceID)
	}
	var anyOwner bool
	for _, member := range cmd.Members {
		if member.Role == WorkspaceRoleOwner {
			anyOwner = true
			break
		}
	}
	if !anyOwner {
		return errs.NewWorkspaceMustHaveAtLeastOneOwner(cmd.WorkspaceID)
	}
	slog.DebugContext(
		ctx, "validating members",
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.Int("member_count", len(cmd.Members)),
	)
	if err := h.authorizationSvc.UpdateWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID, cmd.Members); err != nil {
		return err
	}
	slog.DebugContext(ctx, "members updated in authorization service", slog.String("workspace_id", cmd.WorkspaceID.String()))
	eventID, err := uuid.NewV7()
	if err != nil {
		return errs.NewInternalGenerateID(err)
	}
	err = h.workspaceEventPublisher.Publish(ctx, cmd.WorkspaceID, cmd.UserID, &WorkspaceEventMembersUpdated{
		workspaceEvent[note.WorkspaceMembersUpdatedEventEvent]{
			Id:    eventID,
			Event: note.WorkspaceMembersUpdatedEventEventWorkspaceMembersUpdatedEvent,
			Data: note.WorkspaceMembersUpdatedEventData{
				WorkspaceId: &cmd.WorkspaceID,
			},
		},
	})
	if err == nil {
		slog.InfoContext(
			ctx, "workspace members updated successfully",
			slog.String("workspace_id", cmd.WorkspaceID.String()),
			slog.Int("member_count", len(cmd.Members)),
		)
	}
	return err
}
