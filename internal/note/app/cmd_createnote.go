package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CreateNote struct {
	ID       uuid.UUID
	Name     string
	Icon     string
	FolderID uuid.UUID

	UserID string
}

type CreateNoteHandler struct {
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewCreateNoteHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *CreateNoteHandler {
	return &CreateNoteHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvideCreateNoteHandler = NewCreateNoteHandler

func (h *CreateNoteHandler) Handle(ctx context.Context, cmd *CreateNote) error {
	slog.DebugContext(
		ctx, "creating note",
		slog.String("note_id", cmd.ID.String()),
		slog.String("name", cmd.Name),
		slog.String("folder_id", cmd.FolderID.String()),
		slog.String("user_id", cmd.UserID),
	)
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		noteRepo := r.Note()
		folderExists, err := folderRepo.CheckExists(ctx, cmd.FolderID)
		if err != nil {
			return err
		}
		if !folderExists {
			return errs.NewFolderNotFound(cmd.FolderID, err)
		}
		slog.DebugContext(ctx, "folder exists", slog.String("folder_id", cmd.FolderID.String()))
		workspaceID, err := folderRepo.GetWorkspaceIDByID(ctx, cmd.FolderID)
		if err != nil {
			return err
		}
		slog.DebugContext(ctx, "checking permission", slog.String("user_id", cmd.UserID), slog.String("workspace_id", workspaceID.String()), slog.String("permission", "write"))
		hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionWrite)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %q does not have permission to create note in workspace %q", cmd.UserID, workspaceID.String()),
			)
		}
		slog.DebugContext(ctx, "permission granted", slog.String("user_id", cmd.UserID), slog.String("workspace_id", workspaceID.String()))
		note := domain.NewNote(cmd.ID, cmd.Name, cmd.Icon, cmd.FolderID)
		err = noteRepo.Save(ctx, note)
		if err == nil {
			slog.InfoContext(ctx, "note created successfully", slog.String("note_id", note.ID().String()))
		}
		return err
	})
}
