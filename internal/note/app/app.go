package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(
	tracer trace.Tracer,
	logger *slog.Logger,
) *HandlerProvider {
	return (*HandlerProvider)(
		commonhandler.NewHandlerProvider(
			commonhandler.WithTracer(tracer),
			commonhandler.WithLogger(logger),
		),
	)
}

type Cmds struct {
	ChangeWorkspaceSlug             ChangeWorkspaceSlugCmd
	CreateFolder                    CreateFolderCmd
	CreateNote                      CreateNoteCmd
	CreateWorkspace                 CreateWorkspaceCmd
	DeleteFolder                    PermanentlyDeleteFolderCmd
	DeleteNote                      PermanentlyDeleteNoteCmd
	DeleteWorkspace                 DeleteWorkspaceCmd
	LeaveWorkspace                  LeaveWorkspaceCmd
	MoveWorkspaceItems              MoveWorkspaceItemsCmd
	PermanentlyDeleteWorkspaceItems PermanentlyDeleteWorkspaceItemsCmd
	PublishNote                     PublishNoteCmd
	PublishWorkspace                PublishWorkspaceCmd
	RenameFolder                    RenameFolderCmd
	RenameNote                      RenameNoteCmd
	RenameWorkspace                 RenameWorkspaceCmd
	RestoreTrashedWorkspaceItems    RestoreTrashedWorkspaceItemsCmd
	TrashWorkspaceItems             TrashWorkspaceItemsCmd
	UnpublishNote                   UnpublishNoteCmd
	UnpublishWorkspace              UnpublishWorkspaceCmd
	UpdateWorkspaceMembers          UpdateWorkspaceMembersCmd
}

func NewCmds(
	handlerProvider *HandlerProvider,
	changeWorkspaceSlugHandler *ChangeWorkspaceSlugHandler,
) *Cmds {
	return &Cmds{
		ChangeWorkspaceSlug: commonhandler.DecorateCmd((*commonhandler.HandlerProvider)(handlerProvider), changeWorkspaceSlugHandler),
	}
}

var ProvideCmds = NewCmds

type Queries struct {
	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler
	GetMyWorkspacesHandler          *GetMyWorkspacesHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteHandler                  *GetNoteHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceByNoteHandler       *GetWorkspaceByNoteHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceSearchTokenHandler  *GetWorkspaceSearchTokenHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler
}

type Events struct {
	DocumentCommittedHandler                   *DocumentCommittedHandler
	NoteCreatedDomainToIntegrationEventHandler *NoteCreatedDomainToIntegrationEventHandler
	NoteUpdatedDomainToIntegrationEventHandler *NoteUpdatedDomainToIntegrationEventHandler
	NotifyWorkspaceItemsUpdatedHandler         *NotifyWorkspaceItemsUpdatedHandler
	NotifyWorkspaceRenamedHandler              *NotifyWorkspaceRenamedHandler
	NotifyWorkspaceSlugChangedHandler          *NotifyWorkspaceSlugChangedHandler
}

type Server struct {
	Cmds    *Cmds
	Events  *Events
	Queries *Queries

	WorkspaceEventHub WorkspaceEventHub
}
