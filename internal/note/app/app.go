package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type App struct {
	// Command Handlers
	CreateNoteHandler                   *CreateNoteHandler
	CreateFolderHandler                 *CreateFolderHandler
	CreateWorkspaceHandler              *CreateWorkspaceHandler
	DeleteNoteHandler                   *DeleteNoteHandler
	DeleteWorkspaceHandler              *DeleteWorkspaceHandler
	GenerateDailyNoteHandler            *GenerateDailyNoteHandler
	MoveWorkspaceItemsHandler           *MoveWorkspaceItemsHandler
	PublishNoteHandler                  *PublishNoteHandler
	PublishWorkspaceHandler             *PublishWorkspaceHandler
	RenameFolderHandler                 *RenameFolderHandler
	RenameNoteHandler                   *RenameNoteHandler
	RenameWorkspaceHandler              *RenameWorkspaceHandler
	RestoreTrashedWorkspaceItemsHandler *RestoreTrashedWorkspaceItemsHandler
	TrashWorkspaceItemsHandler          *TrashWorkspaceItemsHandler
	UnpublishNoteHandler                *UnpublishNoteHandler
	UnpublishWorkspaceHandler           *UnpublishWorkspaceHandler
	UpdateWorkspaceMembersHandler       *UpdateWorkspaceMembersHandler

	// Query Handlers
	CheckWorkspaceExistsHandler *CheckWorkspaceExistsHandler
	GetNoteGraphHandler         *GetNoteGraphHandler
	GetNoteLinksHandler         *GetNoteLinksHandler
	GetNotesHandler             *GetNotesHandler
	GetWorkspaceHandler         *GetWorkspaceHandler
	GetWorkspaceGraphHandler    *GetWorkspaceGraphHandler
	GetWorkspaceMembersHandler  *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler     *GetWorkspaceTreeHandler
	ShowTrashHandler            *ShowTrashHandler

	// Event Handlers
	DocumentCommittedHandler *DocumentCommittedHandler

	// Dependencies
	workspaceEventPubSub WorkspaceEventPubSub
	persistence          Persistence
}

func NewApp(
	createNoteHandler *CreateNoteHandler,
	createFolderHandler *CreateFolderHandler,
	createWorkspaceHandler *CreateWorkspaceHandler,
	deleteNoteHandler *DeleteNoteHandler,
	deleteWorkspaceHandler *DeleteWorkspaceHandler,
	generateDailyNoteHandler *GenerateDailyNoteHandler,
	moveWorkspaceItemsHandler *MoveWorkspaceItemsHandler,
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
	checkWorkspaceExistsHandler *CheckWorkspaceExistsHandler,
	getNoteGraphHandler *GetNoteGraphHandler,
	getNoteLinksHandler *GetNoteLinksHandler,
	getNotesHandler *GetNotesHandler,
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
		CreateNoteHandler:                   createNoteHandler,
		CreateFolderHandler:                 createFolderHandler,
		CreateWorkspaceHandler:              createWorkspaceHandler,
		DeleteNoteHandler:                   deleteNoteHandler,
		DeleteWorkspaceHandler:              deleteWorkspaceHandler,
		GenerateDailyNoteHandler:            generateDailyNoteHandler,
		MoveWorkspaceItemsHandler:           moveWorkspaceItemsHandler,
		PublishNoteHandler:                  publishNoteHandler,
		PublishWorkspaceHandler:             publishWorkspaceHandler,
		RenameFolderHandler:                 renameFolderHandler,
		RenameNoteHandler:                   renameNoteHandler,
		RenameWorkspaceHandler:              renameWorkspaceHandler,
		RestoreTrashedWorkspaceItemsHandler: restoreTrashedWorkspaceItemsHandler,
		TrashWorkspaceItemsHandler:          trashWorkspaceItemsHandler,
		UnpublishNoteHandler:                unpublishNoteHandler,
		UnpublishWorkspaceHandler:           unpublishWorkspaceHandler,
		UpdateWorkspaceMembersHandler:       updateWorkspaceMembersHandler,
		CheckWorkspaceExistsHandler:         checkWorkspaceExistsHandler,
		GetNoteGraphHandler:                 getNoteGraphHandler,
		GetNoteLinksHandler:                 getNoteLinksHandler,
		GetNotesHandler:                     getNotesHandler,
		GetWorkspaceHandler:                 getWorkspaceHandler,
		GetWorkspaceGraphHandler:            getWorkspaceGraphHandler,
		GetWorkspaceMembersHandler:          getWorkspaceMembersHandler,
		GetWorkspaceTreeHandler:             getWorkspaceTreeHandler,
		ShowTrashHandler:                    showTrashHandler,
		DocumentCommittedHandler:            documentCommittedHandler,
		workspaceEventPubSub:                workspaceEventPubSub,
		persistence:                         persistence,
	}
}

func (a *App) RunMigration(ctx context.Context) error {
	if a.persistence == nil {
		return nil
	}
	return a.persistence.RunMigrations(ctx)
}

func (a *App) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if err := a.RunMigration(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if a.workspaceEventPubSub != nil {
		g.Go(func() error {
			return a.workspaceEventPubSub.Run(ctx)
		})
	}

	return g.Wait()
}

func (a *App) Stop(ctx context.Context) error {
	if a.workspaceEventPubSub != nil {
		return a.workspaceEventPubSub.Close()
	}
	return nil
}

var ProvideApp = NewApp
