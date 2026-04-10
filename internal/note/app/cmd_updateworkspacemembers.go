package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type UpdateWorkspaceMembers struct {
	WorkspaceID uuid.UUID
	Members     []*WorkspaceMemberUpdate
	UserID      string
}

type UpdateWorkspaceMembersHandler struct {
	workspaceEventPublisher WorkspaceEventPublisher
	authorizationService    AuthorizationService
}

func NewUpdateWorkspaceMembersHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
	authorizationService AuthorizationService,
) *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{
		workspaceEventPublisher: workspaceEventPublisher,
		authorizationService:    authorizationService,
	}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

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
	// NOTE: We don't check if the user have permission to update workspace members
	// Because the service has implemented it for us
	if err := h.authorizationService.UpdateWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID, cmd.Members); err != nil {
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
