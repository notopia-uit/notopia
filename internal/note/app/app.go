package app

import (
	"context"
	"fmt"
)

type App struct {
	CreateNoteHandler                      *CreateNoteHandler
	CreateFolderHandler                    *CreateFolderHandler
	CreateWorkspaceHandler                 *CreateWorkspaceHandler
	DeleteNoteHandler                      *DeleteNoteHandler
	DeleteFolderHandler                    *DeleteFolderHandler
	DeleteWorkspaceHandler                 *DeleteWorkspaceHandler
	GenerateDailyNoteHandler               *GenerateDailyNoteHandler
	MoveWorkspaceItemsHandler              *MoveWorkspaceItemsHandler
	PermanentlyDeleteWorkspaceItemsHandler *PermanentlyDeleteWorkspaceItemsHandler
	PublishNoteHandler                     *PublishNoteHandler
	PublishWorkspaceHandler                *PublishWorkspaceHandler
	RenameFolderHandler                    *RenameFolderHandler
	RenameNoteHandler                      *RenameNoteHandler
	RenameWorkspaceHandler                 *RenameWorkspaceHandler
	RestoreTrashedWorkspaceItemsHandler    *RestoreTrashedWorkspaceItemsHandler
	TrashWorkspaceItemsHandler             *TrashWorkspaceItemsHandler
	UnpublishNoteHandler                   *UnpublishNoteHandler
	UnpublishWorkspaceHandler              *UnpublishWorkspaceHandler
	UpdateWorkspaceMembersHandler          *UpdateWorkspaceMembersHandler

	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler

	DocumentCommittedHandler *DocumentCommittedHandler

	workspaceEventPubSub WorkspaceEventPubSub
	persistence          Persistence
}

func NewApp(
	createNoteHandler *CreateNoteHandler,
	createFolderHandler *CreateFolderHandler,
	createWorkspaceHandler *CreateWorkspaceHandler,
	deleteNoteHandler *DeleteNoteHandler,
	deleteFolderHandler *DeleteFolderHandler,
	deleteWorkspaceHandler *DeleteWorkspaceHandler,
	generateDailyNoteHandler *GenerateDailyNoteHandler,
	moveWorkspaceItemsHandler *MoveWorkspaceItemsHandler,
	permanentlyDeleteWorkspaceItemsHandler *PermanentlyDeleteWorkspaceItemsHandler,
	publishNoteHandler *PublishNoteHandler,
	publishWorkspaceHandler *PublishWorkspaceHandler,
	renameFolderHandler *RenameFolderHandler,
	renameNoteHandler *RenameNoteHandler,
	renameWorkspaceHandler *RenameWorkspaceHandler,
	restoreTrashedWorkspaceItemsHandler *RestoreTrashedWorkspaceItemsHandler,
	trashWorkspaceItemsHandler *TrashWorkspaceItemsHandler,
	unpublishNoteHandler *UnpublishNoteHandler,
	unpublishWorkspaceHandler *UnpublishWorkspaceHandler,
	updateWorkspaceMembersHandler *UpdateWorkspaceMembersHandler,
	checkWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler,
	getNoteGraphHandler *GetNoteGraphHandler,
	getNoteLinksHandler *GetNoteLinksHandler,
	getWorkspaceHandler *GetWorkspaceHandler,
	getWorkspaceGraphHandler *GetWorkspaceGraphHandler,
	getWorkspaceMembersHandler *GetWorkspaceMembersHandler,
	getWorkspaceTreeHandler *GetWorkspaceTreeHandler,
	showTrashHandler *ShowTrashHandler,
	documentCommittedHandler *DocumentCommittedHandler,
	workspaceEventPubSub WorkspaceEventPubSub,
	persistence Persistence,
) *App {
	return &App{
		CreateNoteHandler:                      createNoteHandler,
		CreateFolderHandler:                    createFolderHandler,
		CreateWorkspaceHandler:                 createWorkspaceHandler,
		DeleteNoteHandler:                      deleteNoteHandler,
		DeleteFolderHandler:                    deleteFolderHandler,
		DeleteWorkspaceHandler:                 deleteWorkspaceHandler,
		GenerateDailyNoteHandler:               generateDailyNoteHandler,
		MoveWorkspaceItemsHandler:              moveWorkspaceItemsHandler,
		PermanentlyDeleteWorkspaceItemsHandler: permanentlyDeleteWorkspaceItemsHandler,
		PublishNoteHandler:                     publishNoteHandler,
		PublishWorkspaceHandler:                publishWorkspaceHandler,
		RenameFolderHandler:                    renameFolderHandler,
		RenameNoteHandler:                      renameNoteHandler,
		RenameWorkspaceHandler:                 renameWorkspaceHandler,
		RestoreTrashedWorkspaceItemsHandler:    restoreTrashedWorkspaceItemsHandler,
		TrashWorkspaceItemsHandler:             trashWorkspaceItemsHandler,
		UnpublishNoteHandler:                   unpublishNoteHandler,
		UnpublishWorkspaceHandler:              unpublishWorkspaceHandler,
		UpdateWorkspaceMembersHandler:          updateWorkspaceMembersHandler,
		CheckWorkspaceSlugExistsHandler:        checkWorkspaceSlugExistsHandler,
		GetNoteGraphHandler:                    getNoteGraphHandler,
		GetNoteLinksHandler:                    getNoteLinksHandler,
		GetWorkspaceHandler:                    getWorkspaceHandler,
		GetWorkspaceGraphHandler:               getWorkspaceGraphHandler,
		GetWorkspaceMembersHandler:             getWorkspaceMembersHandler,
		GetWorkspaceTreeHandler:                getWorkspaceTreeHandler,
		ShowTrashHandler:                       showTrashHandler,
		DocumentCommittedHandler:               documentCommittedHandler,
		workspaceEventPubSub:                   workspaceEventPubSub,
		persistence:                            persistence,
	}
}

func (a *App) RunMigration(ctx context.Context) error {
	return a.persistence.RunMigrations(ctx)
}

func (a *App) Start(ctx context.Context) error {
	if err := a.RunMigration(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return a.workspaceEventPubSub.Run(ctx)
}

func (a *App) Stop(ctx context.Context) error {
	return a.workspaceEventPubSub.Close()
}

var ProvideApp = NewApp
