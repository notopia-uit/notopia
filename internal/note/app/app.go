package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "note-app"

type Cmds struct {
	ChangeWorkspaceSlug             ChangeWorkspaceSlugCmd
	CreateFolder                    CreateFolderCmd
	CreateNote                      CreateNoteCmd
	CreateWorkspace                 CreateWorkspaceCmd
	DeleteFolder                    PermanentlyDeleteFolderCmd
	DeleteNote                      PermanentlyDeleteNoteCmd
	DeleteWorkspace                 DeleteWorkspaceCmd
	EmptyTrash                      EmptyTrashCmd
	LeaveWorkspace                  LeaveWorkspaceCmd
	MoveWorkspaceItems              MoveWorkspaceItemsCmd
	PermanentlyDeleteWorkspaceItems PermanentlyDeleteWorkspaceItemsCmd
	PublishNote                     PublishNoteCmd
	PublishWorkspace                PublishWorkspaceCmd
	PurgeExpiredTrash               PurgeExpiredTrashCmd
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
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	changeWorkspaceSlugHandler *ChangeWorkspaceSlugHandler,
	createFolderHandler *CreateFolderHandler,
	createNoteHandler *CreateNoteHandler,
	createWorkspaceHandler *CreateWorkspaceHandler,
	deleteWorkspaceHandler *DeleteWorkspaceHandler,
	emptyTrashHandler *EmptyTrashHandler,
	leaveWorkspaceHandler *LeaveWorkspaceHandler,
	moveWorkspaceItemsHandler *MoveWorkspaceItemsHandler,
	permanentlyDeleteFolderHandler *PermanentlyDeleteFolderHandler,
	permanentlyDeleteNoteHandler *PermanentlyDeleteNoteHandler,
	permanentlyDeleteWorkspaceItemsHandler *PermanentlyDeleteWorkspaceItemsHandler,
	publishNoteHandler *PublishNoteHandler,
	publishWorkspaceHandler *PublishWorkspaceHandler,
	purgeExpiredTrashHandler *PurgeExpiredTrashHandler,
	renameFolderHandler *RenameFolderHandler,
	renameNoteHandler *RenameNoteHandler,
	renameWorkspaceHandler *RenameWorkspaceHandler,
	restoreTrashedWorkspaceItemsHandler *RestoreTrashedWorkspaceItemsHandler,
	trashWorkspaceItemsHandler *TrashWorkspaceItemsHandler,
	unpublishNoteHandler *UnpublishNoteHandler,
	unpublishWorkspaceHandler *UnpublishWorkspaceHandler,
	updateWorkspaceMembersHandler *UpdateWorkspaceMembersHandler,
) *Cmds {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Cmds{
		ChangeWorkspaceSlug:             d.Cmd(changeWorkspaceSlugHandler),
		CreateFolder:                    d.Cmd(createFolderHandler),
		CreateNote:                      d.Cmd(createNoteHandler),
		CreateWorkspace:                 d.Cmd(createWorkspaceHandler),
		DeleteWorkspace:                 d.Cmd(deleteWorkspaceHandler),
		EmptyTrash:                      d.Cmd(emptyTrashHandler),
		LeaveWorkspace:                  d.Cmd(leaveWorkspaceHandler),
		MoveWorkspaceItems:              d.Cmd(moveWorkspaceItemsHandler),
		DeleteFolder:                    d.Cmd(permanentlyDeleteFolderHandler),
		DeleteNote:                      d.Cmd(permanentlyDeleteNoteHandler),
		PermanentlyDeleteWorkspaceItems: d.Cmd(permanentlyDeleteWorkspaceItemsHandler),
		PublishNote:                     d.Cmd(publishNoteHandler),
		PublishWorkspace:                d.Cmd(publishWorkspaceHandler),
		PurgeExpiredTrash:               d.Cmd(purgeExpiredTrashHandler),
		RenameFolder:                    d.Cmd(renameFolderHandler),
		RenameNote:                      d.Cmd(renameNoteHandler),
		RenameWorkspace:                 d.Cmd(renameWorkspaceHandler),
		RestoreTrashedWorkspaceItems:    d.Cmd(restoreTrashedWorkspaceItemsHandler),
		TrashWorkspaceItems:             d.Cmd(trashWorkspaceItemsHandler),
		UnpublishNote:                   d.Cmd(unpublishNoteHandler),
		UnpublishWorkspace:              d.Cmd(unpublishWorkspaceHandler),
		UpdateWorkspaceMembers:          d.Cmd(updateWorkspaceMembersHandler),
	}
}

var ProvideCmds = NewCmds

type Queries struct {
	BatchGetNoteWorkspaces   BatchGetNoteWorkspacesQuery
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
	SearchUsers              SearchUsersQuery
	ShowTrash                ShowTrashQuery
}

func NewQueries(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	batchGetNoteWorkspacesHandler *BatchGetNoteWorkspacesHandler,
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
	SearchUsersHandler *SearchUsersHandler,
	ShowTrashHandler *ShowTrashHandler,
) *Queries {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Queries{
		BatchGetNoteWorkspaces:   d.Query(batchGetNoteWorkspacesHandler),
		CheckWorkspaceSlugExists: d.Query(CheckWorkspaceSlugExistsHandler),
		GetMyWorkspaces:          d.Query(GetMyWorkspacesHandler),
		GetNoteGraph:             d.Query(GetNoteGraphHandler),
		GetNote:                  d.Query(GetNoteHandler),
		GetNoteLinks:             d.Query(GetNoteLinksHandler),
		GetWorkspaceByNote:       d.Query(GetWorkspaceByNoteHandler),
		GetWorkspaceGraph:        d.Query(GetWorkspaceGraphHandler),
		GetWorkspace:             d.Query(GetWorkspaceHandler),
		GetWorkspaceMembers:      d.Query(GetWorkspaceMembersHandler),
		GetWorkspaceSearchToken:  d.Query(GetWorkspaceSearchTokenHandler),
		GetWorkspaceTree:         d.Query(GetWorkspaceTreeHandler),
		SearchUsers:              d.Query(SearchUsersHandler),
		ShowTrash:                d.Query(ShowTrashHandler),
	}
}

var ProvideQueries = NewQueries

type Events struct {
	BillingChanged                           *BillingChangedHandler
	DocumentCommitted                        *DocumentCommittedHandler
	NoteCreatedDomainToIntegrationEvent      *NoteCreatedDomainToIntegrationEventHandler
	NoteDeletedDomainToIntegrationEvent      *NoteDeletedDomainToIntegrationEventHandler
	NoteUpdatedDomainToIntegrationEvent      *NoteUpdatedDomainToIntegrationEventHandler
	NotifyWorkspaceItemsUpdated              *NotifyWorkspaceItemsUpdatedHandler
	NotifyWorkspaceRenamed                   *NotifyWorkspaceRenamedHandler
	NotifyWorkspaceSlugChanged               *NotifyWorkspaceSlugChangedHandler
	WorkspaceDeletedDomainToIntegrationEvent *WorkspaceDeletedDomainToIntegrationEventHandler
}

type Server struct {
	Cmds    *Cmds
	Events  *Events
	Queries *Queries

	WorkspaceEventHub WorkspaceEventHub
}
