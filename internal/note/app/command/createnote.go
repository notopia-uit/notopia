package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CreateNote struct {
	ID       uuid.UUID
	Name     string
	Icon     *string
	Tags     []string
	FolderID uuid.UUID

	UserID string
}

type CreateNoteHandler struct {
	authorization service.Authorization
	noteRepo      domain.NoteRepo
	folderRepo    domain.FolderRepo
	eventPubSub   pubsub.WorkspaceEvent
}

func NewCreateNoteHandler(
	authorization service.Authorization,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	eventPubSub pubsub.WorkspaceEvent,
) *CreateNoteHandler {
	return &CreateNoteHandler{
		authorization: authorization,
		noteRepo:      noteRepo,
		folderRepo:    folderRepo,
		eventPubSub:   eventPubSub,
	}
}

var ProvideCreateNoteHandler = NewCreateNoteHandler

func (h *CreateNoteHandler) Handle(ctx context.Context, cmd *CreateNote) errs.Error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.FolderID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		service.WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %q does not have permission to create note in workspace %q", cmd.UserID, workspaceID.String()),
		)
	}
	note := domain.NewNote(cmd.ID, cmd.Name, cmd.Icon, cmd.Tags, cmd.FolderID)
	return h.noteRepo.Save(ctx, note)
}
