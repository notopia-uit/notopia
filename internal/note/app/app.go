package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type CommandHandlers struct {
	CreateFolderHandler                    *CreateFolderHandler
	CreateNoteHandler                      *CreateNoteHandler
	CreateWorkspaceHandler                 *CreateWorkspaceHandler
	DeleteFolderHandler                    *PermanentlyDeleteFolderHandler
	DeleteNoteHandler                      *PermanentlyDeleteNoteHandler
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
}

type QueryHandlers struct {
	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteHandler                  *GetNoteHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler
}

type IntegrationEventHandlers struct {
	DocumentCommittedHandler *DocumentCommittedHandler
}

type Server struct {
	CommandHandlers          *CommandHandlers
	IntegrationEventHandlers *IntegrationEventHandlers
	QueryHandlers            *QueryHandlers

	WorkspaceEventHub WorkspaceEventHub
	Persistence       Persistence
}

func NewServer(
	commandHandlers *CommandHandlers,
	integrationEventHandlers *IntegrationEventHandlers,
	queryHandlers *QueryHandlers,
	workspaceEventHub WorkspaceEventHub,
	persistence Persistence,
) *Server {
	return &Server{
		CommandHandlers:          commandHandlers,
		IntegrationEventHandlers: integrationEventHandlers,
		QueryHandlers:            queryHandlers,
		WorkspaceEventHub:        workspaceEventHub,
		Persistence:              persistence,
	}
}

var ProvideServer = NewServer

func (s *Server) RunMigration(ctx context.Context) error {
	return s.Persistence.RunMigrations(ctx)
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.RunMigration(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.WorkspaceEventHub.Run(ctx)
	})
	return g.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.WorkspaceEventHub.Close()
	})
	return g.Wait()
}
