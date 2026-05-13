package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
) *HandlerProvider {
	tracer := traceProvider.Tracer("note-app")
	return (*HandlerProvider)(
		commonhandler.NewHandlerProvider(
			commonhandler.WithTracer(tracer),
			commonhandler.WithLogger(logger),
		),
	)
}

var ProvideHandlerProvider = NewHandlerProvider

type (
	ChangeWorkspaceSlugCmd             commonhandler.Cmd[ChangeWorkspaceSlug]
	CreateFolderCmd                    commonhandler.Cmd[CreateFolder]
	CreateNoteCmd                      commonhandler.Cmd[CreateNote]
	CreateWorkspaceCmd                 commonhandler.Cmd[CreateWorkspace]
	DeleteWorkspaceCmd                 commonhandler.Cmd[DeleteWorkspace]
	LeaveWorkspaceCmd                  commonhandler.Cmd[LeaveWorkspace]
	MoveWorkspaceItemsCmd              commonhandler.Cmd[MoveWorkspaceItems]
	PermanentlyDeleteFolderCmd         commonhandler.Cmd[PermanentlyDeleteFolder]
	PermanentlyDeleteNoteCmd           commonhandler.Cmd[PermanentlyDeleteNote]
	PermanentlyDeleteWorkspaceItemsCmd commonhandler.Cmd[PermanentlyDeleteWorkspaceItems]
	PublishNoteCmd                     commonhandler.Cmd[PublishNote]
	PublishWorkspaceCmd                commonhandler.Cmd[PublishWorkspace]
	RenameFolderCmd                    commonhandler.Cmd[RenameFolder]
	RenameNoteCmd                      commonhandler.Cmd[RenameNote]
	RenameWorkspaceCmd                 commonhandler.Cmd[RenameWorkspace]
	RestoreTrashedWorkspaceItemsCmd    commonhandler.Cmd[RestoreTrashedWorkspaceItems]
	TrashWorkspaceItemsCmd             commonhandler.Cmd[TrashWorkspaceItems]
	UnpublishNoteCmd                   commonhandler.Cmd[UnpublishNote]
	UnpublishWorkspaceCmd              commonhandler.Cmd[UnpublishWorkspace]
	UpdateWorkspaceMembersCmd          commonhandler.Cmd[UpdateWorkspaceMembers]
)

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
	createFolderHandler *CreateFolderHandler,
	createNoteHandler *CreateNoteHandler,
	createWorkspaceHandler *CreateWorkspaceHandler,
	deleteWorkspaceHandler *DeleteWorkspaceHandler,
	leaveWorkspaceHandler *LeaveWorkspaceHandler,
	moveWorkspaceItemsHandler *MoveWorkspaceItemsHandler,
	permanentlyDeleteFolderHandler *PermanentlyDeleteFolderHandler,
	permanentlyDeleteNoteHandler *PermanentlyDeleteNoteHandler,
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
) *Cmds {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Cmds{
		ChangeWorkspaceSlug:             commonhandler.DecorateCmd(hp, changeWorkspaceSlugHandler),
		CreateFolder:                    commonhandler.DecorateCmd(hp, createFolderHandler),
		CreateNote:                      commonhandler.DecorateCmd(hp, createNoteHandler),
		CreateWorkspace:                 commonhandler.DecorateCmd(hp, createWorkspaceHandler),
		DeleteWorkspace:                 commonhandler.DecorateCmd(hp, deleteWorkspaceHandler),
		LeaveWorkspace:                  commonhandler.DecorateCmd(hp, leaveWorkspaceHandler),
		MoveWorkspaceItems:              commonhandler.DecorateCmd(hp, moveWorkspaceItemsHandler),
		DeleteFolder:                    commonhandler.DecorateCmd(hp, permanentlyDeleteFolderHandler),
		DeleteNote:                      commonhandler.DecorateCmd(hp, permanentlyDeleteNoteHandler),
		PermanentlyDeleteWorkspaceItems: commonhandler.DecorateCmd(hp, permanentlyDeleteWorkspaceItemsHandler),
		PublishNote:                     commonhandler.DecorateCmd(hp, publishNoteHandler),
		PublishWorkspace:                commonhandler.DecorateCmd(hp, publishWorkspaceHandler),
		RenameFolder:                    commonhandler.DecorateCmd(hp, renameFolderHandler),
		RenameNote:                      commonhandler.DecorateCmd(hp, renameNoteHandler),
		RenameWorkspace:                 commonhandler.DecorateCmd(hp, renameWorkspaceHandler),
		RestoreTrashedWorkspaceItems:    commonhandler.DecorateCmd(hp, restoreTrashedWorkspaceItemsHandler),
		TrashWorkspaceItems:             commonhandler.DecorateCmd(hp, trashWorkspaceItemsHandler),
		UnpublishNote:                   commonhandler.DecorateCmd(hp, unpublishNoteHandler),
		UnpublishWorkspace:              commonhandler.DecorateCmd(hp, unpublishWorkspaceHandler),
		UpdateWorkspaceMembers:          commonhandler.DecorateCmd(hp, updateWorkspaceMembersHandler),
	}
}

var ProvideCmds = NewCmds

type (
	CheckWorkspaceSlugExistsQuery commonhandler.Query[CheckWorkspaceSlugExists, bool]
	GetMyWorkspacesQuery          commonhandler.Query[GetMyWorkspaces, []UserWorkspace]
	GetNoteGraphQuery             commonhandler.Query[GetNoteGraph, Graph]
	GetNoteQuery                  commonhandler.Query[GetNote, *Note]
	GetNoteLinksQuery             commonhandler.Query[GetNoteLinks, NoteLinkResult]
	GetWorkspaceByNoteQuery       commonhandler.Query[GetWorkspaceByNote, Workspace]
	GetWorkspaceGraphQuery        commonhandler.Query[GetWorkspaceGraph, Graph]
	GetWorkspaceBySlugQuery       commonhandler.Query[GetWorkspaceBySlug, Workspace]
	GetWorkspaceMembersQuery      commonhandler.Query[GetWorkspaceMembers, []WorkspaceMember]
	GetWorkspaceSearchTokenQuery  commonhandler.Query[GetWorkspaceSearchToken, SearchToken]
	GetWorkspaceTreeQuery         commonhandler.Query[GetWorkspaceTree, WorkspaceTreeFolder]
	ShowTrashQuery                commonhandler.Query[ShowTrash, Trash]
)

type Queries struct {
	CheckWorkspaceSlugExists CheckWorkspaceSlugExistsQuery
	GetMyWorkspaces          GetMyWorkspacesQuery
	GetNoteGraph             GetNoteGraphQuery
	GetNote                  GetNoteQuery
	GetNoteLinks             GetNoteLinksQuery
	GetWorkspaceByNote       GetWorkspaceByNoteQuery
	GetWorkspaceGraph        GetWorkspaceGraphQuery
	GetWorkspace             GetWorkspaceBySlugQuery
	GetWorkspaceMembers      GetWorkspaceMembersQuery
	GetWorkspaceSearchToken  GetWorkspaceSearchTokenQuery
	GetWorkspaceTree         GetWorkspaceTreeQuery
	ShowTrash                ShowTrashQuery
}

func NewQueries(
	handlerProvider *HandlerProvider,
	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler,
	GetMyWorkspacesHandler *GetMyWorkspacesHandler,
	GetNoteGraphHandler *GetNoteGraphHandler,
	GetNoteHandler *GetNoteHandler,
	GetNoteLinksHandler *GetNoteLinksHandler,
	GetWorkspaceByNoteHandler *GetWorkspaceByNoteHandler,
	GetWorkspaceGraphHandler *GetWorkspaceGraphHandler,
	GetWorkspaceHandler *GetWorkspaceBySlugHandler,
	GetWorkspaceMembersHandler *GetWorkspaceMembersHandler,
	GetWorkspaceSearchTokenHandler *GetWorkspaceSearchTokenHandler,
	GetWorkspaceTreeHandler *GetWorkspaceTreeHandler,
	ShowTrashHandler *ShowTrashHandler,
) *Queries {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Queries{
		CheckWorkspaceSlugExists: CheckWorkspaceSlugExistsQuery(commonhandler.DecorateQuery(hp, CheckWorkspaceSlugExistsHandler)),
		GetMyWorkspaces:          GetMyWorkspacesQuery(commonhandler.DecorateQuery(hp, GetMyWorkspacesHandler)),
		GetNoteGraph:             GetNoteGraphQuery(commonhandler.DecorateQuery(hp, GetNoteGraphHandler)),
		GetNote:                  GetNoteQuery(commonhandler.DecorateQuery(hp, GetNoteHandler)),
		GetNoteLinks:             GetNoteLinksQuery(commonhandler.DecorateQuery(hp, GetNoteLinksHandler)),
		GetWorkspaceByNote:       GetWorkspaceByNoteQuery(commonhandler.DecorateQuery(hp, GetWorkspaceByNoteHandler)),
		GetWorkspaceGraph:        GetWorkspaceGraphQuery(commonhandler.DecorateQuery(hp, GetWorkspaceGraphHandler)),
		GetWorkspace:             GetWorkspaceBySlugQuery(commonhandler.DecorateQuery(hp, GetWorkspaceHandler)),
		GetWorkspaceMembers:      GetWorkspaceMembersQuery(commonhandler.DecorateQuery(hp, GetWorkspaceMembersHandler)),
		GetWorkspaceSearchToken:  GetWorkspaceSearchTokenQuery(commonhandler.DecorateQuery(hp, GetWorkspaceSearchTokenHandler)),
		GetWorkspaceTree:         GetWorkspaceTreeQuery(commonhandler.DecorateQuery(hp, GetWorkspaceTreeHandler)),
		ShowTrash:                ShowTrashQuery(commonhandler.DecorateQuery(hp, ShowTrashHandler)),
	}
}

var ProvideQueries = NewQueries

type Events struct {
	DocumentCommitted                   *DocumentCommittedHandler
	NoteCreatedDomainToIntegrationEvent *NoteCreatedDomainToIntegrationEventHandler
	NoteUpdatedDomainToIntegrationEvent *NoteUpdatedDomainToIntegrationEventHandler
	NotifyWorkspaceItemsUpdated         *NotifyWorkspaceItemsUpdatedHandler
	NotifyWorkspaceRenamed              *NotifyWorkspaceRenamedHandler
	NotifyWorkspaceSlugChanged          *NotifyWorkspaceSlugChangedHandler
}

type Server struct {
	Cmds    *Cmds
	Events  *Events
	Queries *Queries

	WorkspaceEventHub WorkspaceEventHub
}
