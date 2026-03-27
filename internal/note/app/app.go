package app

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/app/command"
	"github.com/notopia-uit/notopia/internal/note/app/event"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/app/query"
)

type App struct {
	CreateNoteHandler                      *command.CreateNoteHandler
	CreateFolderHandler                    *command.CreateFolderHandler
	CreateWorkspaceHandler                 *command.CreateWorkspaceHandler
	DeleteNoteHandler                      *command.DeleteNoteHandler
	DeleteFolderHandler                    *command.DeleteFolderHandler
	DeleteWorkspaceHandler                 *command.DeleteWorkspaceHandler
	GenerateDailyNoteHandler               *command.GenerateDailyNoteHandler
	MoveWorkspaceItemsHandler              *command.MoveWorkspaceItemsHandler
	PermanentlyDeleteWorkspaceItemsHandler *command.PermanentlyDeleteWorkspaceItemsHandler
	PublishNoteHandler                     *command.PublishNoteHandler
	PublishWorkspaceHandler                *command.PublishWorkspaceHandler
	RenameFolderHandler                    *command.RenameFolderHandler
	RenameNoteHandler                      *command.RenameNoteHandler
	RenameWorkspaceHandler                 *command.RenameWorkspaceHandler
	RestoreTrashedWorkspaceItemsHandler    *command.RestoreTrashedWorkspaceItemsHandler
	TrashWorkspaceItemsHandler             *command.TrashWorkspaceItemsHandler
	UnpublishNoteHandler                   *command.UnpublishNoteHandler
	UnpublishWorkspaceHandler              *command.UnpublishWorkspaceHandler
	UpdateWorkspaceMembersHandler          *command.UpdateWorkspaceMembersHandler

	CheckWorkspaceSlugExistsHandler *query.CheckWorkspaceSlugExistsHandler
	GetNoteHandler                  *query.GetNoteHandler
	GetNoteGraphHandler             *query.GetNoteGraphHandler
	GetNoteLinksHandler             *query.GetNoteLinksHandler
	GetWorkspaceHandler             *query.GetWorkspaceHandler
	GetWorkspaceGraphHandler        *query.GetWorkspaceGraphHandler
	GetWorkspaceMembersHandler      *query.GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *query.GetWorkspaceTreeHandler
	ShowTrashHandler                *query.ShowTrashHandler

	DocumentCommittedHandler *event.DocumentCommittedHandler

	workspaceEventPubSub pubsub.WorkspaceEvent
	persistence          Persistence
}

func NewApp(
	createNoteHandler *command.CreateNoteHandler,
	createFolderHandler *command.CreateFolderHandler,
	createWorkspaceHandler *command.CreateWorkspaceHandler,
	deleteNoteHandler *command.DeleteNoteHandler,
	deleteFolderHandler *command.DeleteFolderHandler,
	deleteWorkspaceHandler *command.DeleteWorkspaceHandler,
	generateDailyNoteHandler *command.GenerateDailyNoteHandler,
	moveWorkspaceItemsHandler *command.MoveWorkspaceItemsHandler,
	permanentlyDeleteWorkspaceItemsHandler *command.PermanentlyDeleteWorkspaceItemsHandler,
	publishNoteHandler *command.PublishNoteHandler,
	publishWorkspaceHandler *command.PublishWorkspaceHandler,
	renameFolderHandler *command.RenameFolderHandler,
	renameNoteHandler *command.RenameNoteHandler,
	renameWorkspaceHandler *command.RenameWorkspaceHandler,
	restoreTrashedWorkspaceItemsHandler *command.RestoreTrashedWorkspaceItemsHandler,
	trashWorkspaceItemsHandler *command.TrashWorkspaceItemsHandler,
	unpublishNoteHandler *command.UnpublishNoteHandler,
	unpublishWorkspaceHandler *command.UnpublishWorkspaceHandler,
	updateWorkspaceMembersHandler *command.UpdateWorkspaceMembersHandler,
	checkWorkspaceSlugExistsHandler *query.CheckWorkspaceSlugExistsHandler,
	getNoteHandler *query.GetNoteHandler,
	getNoteGraphHandler *query.GetNoteGraphHandler,
	getNoteLinksHandler *query.GetNoteLinksHandler,
	getWorkspaceHandler *query.GetWorkspaceHandler,
	getWorkspaceGraphHandler *query.GetWorkspaceGraphHandler,
	getWorkspaceMembersHandler *query.GetWorkspaceMembersHandler,
	getWorkspaceTreeHandler *query.GetWorkspaceTreeHandler,
	showTrashHandler *query.ShowTrashHandler,
	documentCommittedHandler *event.DocumentCommittedHandler,
	workspaceEventPubSub pubsub.WorkspaceEvent,
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
		GetNoteHandler:                         getNoteHandler,
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
