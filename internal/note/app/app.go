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

type DomainEventHandlers struct {
	bus DomainEventProcessor
}

func NewDomainEventHandlers(bus DomainEventProcessor) (*DomainEventHandlers, error) {
	handlers := &DomainEventHandlers{
		bus: bus,
	}
	if err := handlers.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register domain event handlers: %w", err)
	}
	return handlers, nil
}

var ProvideDomainEventHandlers = NewDomainEventHandlers

func (h *DomainEventHandlers) registerHandlers() error {
	if err := h.bus.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register NoteCreatedHandler: %w", err)
	}
	return nil
}

func (h *DomainEventHandlers) Run(ctx context.Context) error {
	return h.bus.Run(ctx)
}

func (h *DomainEventHandlers) Close() error {
	return h.bus.Close()
}

type IntegrationEventHandlers struct {
	DocumentCommittedHandler *DocumentCommittedHandler
}

type Server struct {
	CommandHandlers          *CommandHandlers
	IntegrationEventHandlers *IntegrationEventHandlers
	QueryHandlers            *QueryHandlers
	DomainEventHandlers      *DomainEventHandlers

	WorkspaceEventHub WorkspaceEventHub
	Persistence       Persistence
}

func NewServer(
	commandHandlers *CommandHandlers,
	integrationEventHandlers *IntegrationEventHandlers,
	queryHandlers *QueryHandlers,
	DomainEventHandlers *DomainEventHandlers,
	workspaceEventHub WorkspaceEventHub,
	persistence Persistence,
) *Server {
	return &Server{
		CommandHandlers:          commandHandlers,
		IntegrationEventHandlers: integrationEventHandlers,
		QueryHandlers:            queryHandlers,
		DomainEventHandlers:      DomainEventHandlers,
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
	g.Go(func() error {
		return s.DomainEventHandlers.Run(ctx)
	})
	return g.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.WorkspaceEventHub.Close()
	})
	g.Go(func() error {
		return s.DomainEventHandlers.Close()
	})
	return g.Wait()
}
