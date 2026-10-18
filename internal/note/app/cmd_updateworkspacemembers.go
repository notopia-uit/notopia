package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdateWorkspaceMembers struct {
	WorkspaceID uuid.UUID
	Members     []WorkspaceMemberUpdate
	UserID      string
}

type UpdateWorkspaceMembersHandler struct {
	workspaceEventPublisher WorkspaceEventPublisher
	authorizationSvc        AuthorizationSvc
	billingSvc              BillingSvc
}

func NewUpdateWorkspaceMembersHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
	authorizationSvc AuthorizationSvc,
	billingSvc BillingSvc,
) *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{
		workspaceEventPublisher: workspaceEventPublisher,
		authorizationSvc:        authorizationSvc,
		billingSvc:              billingSvc,
	}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

type UpdateWorkspaceMembersCmd commonhandler.Cmd[UpdateWorkspaceMembers]

var _ UpdateWorkspaceMembersCmd = (*UpdateWorkspaceMembersHandler)(nil)

// FIXME: This maybe need saga? Because it involves 2 external things
// Or we have to persist event and let another handler to consume
func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, cmd *UpdateWorkspaceMembers) error {
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
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		WorkspacePermissionEdit,
	)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to update members of workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}
	current, err := h.authorizationSvc.GetWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if err := checkMembersLimit(ctx, h.billingSvc, cmd.WorkspaceID, len(current), len(cmd.Members)); err != nil {
		return err
	}
	if err := checkPayerKept(ctx, h.billingSvc, cmd.WorkspaceID, cmd.Members); err != nil {
		return err
	}
	if err := h.authorizationSvc.UpdateWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID, cmd.Members); err != nil {
		return err
	}
	eventID, err := uuid.NewV7()
	if err != nil {
		return errs.NewInternalGenerateID(err)
	}
	return h.workspaceEventPublisher.Publish(ctx, cmd.WorkspaceID, cmd.UserID, &WorkspaceEventMembersUpdated{
		workspaceEvent[note.WorkspaceMembersUpdatedEventEvent]{
			Id:    eventID,
			Event: note.WorkspaceMembersUpdatedEventEventWorkspaceMembersUpdatedEvent,
			Data: note.WorkspaceMembersUpdatedEventData{
				WorkspaceId: &cmd.WorkspaceID,
			},
		},
	})
}
