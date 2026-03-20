package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
	workspaceEventPubSub WorkspaceEventPubSub
}

func NewCreateNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	workspaceEventPubSub WorkspaceEventPubSub,
) *CreateNoteHandler {
	return &CreateNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
		workspaceEventPubSub: workspaceEventPubSub,
	}
}

var ProvideCreateNoteHandler = NewCreateNoteHandler

func (h *CreateNoteHandler) Handle(ctx context.Context, cmd *CreateNote) error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.FolderID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}
	if !hasPermission {
		return newErrCreateNoteForbidden(cmd.UserID, workspaceID)
	}
	note := domain.NewNote(cmd.ID, cmd.Name, cmd.Icon, cmd.Tags, cmd.FolderID)
	return h.noteRepo.Save(ctx, note)
}

var ErrCodeCreateNoteForbidden = "CreateNote_1"

func newErrCreateNoteForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to create note in workspace %q", userID, workspaceID.String()),
		ErrCodeCreateNoteForbidden,
		nil,
	)
}
