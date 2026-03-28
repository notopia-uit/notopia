package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type CommandHandlers struct {
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
}

type QueryHandlers struct {
	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler
}

type IntegrationEventHandlers struct {
	DocumentCommittedHandler *DocumentCommittedHandler
}

type Server struct {
	CommandHandlers          *CommandHandlers
	QueryHandlers            *QueryHandlers
	IntegrationEventHandlers *IntegrationEventHandlers

	IntegrationPubSub    *IntegrationPubSub
	WorkspaceEventPubSub WorkspaceEventPubSub
	Persistence          Persistence
}

func (s *Server) RunMigration(ctx context.Context) error {
	return s.Persistence.RunMigrations(ctx)
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.RunMigration(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.WorkspaceEventPubSub.Run(ctx)
	})
	g.Go(func() error {
		return s.IntegrationPubSub.Run(ctx)
	})
	return g.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.WorkspaceEventPubSub.Close()
	})
	g.Go(func() error {
		return s.IntegrationPubSub.Close()
	})
	return g.Wait()
}
