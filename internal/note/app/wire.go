package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	ProvideCmds,

	ProvideChangeWorkspaceSlugHandler,
	ProvideCreateFolderHandler,
	ProvideCreateNoteHandler,
	ProvideCreateWorkspaceHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideGetNoteHandler,
	ProvideLeaveWorkspaceHandler,
	ProvideMoveWorkspaceItemsHandler,
	ProvidePermanentlyDeleteFolderHandler,
	ProvidePermanentlyDeleteNoteHandler,
	ProvidePermanentlyDeleteWorkspaceItemsHandler,
	ProvidePublishNoteHandler,
	ProvidePublishWorkspaceHandler,
	ProvideRenameFolderHandler,
	ProvideRenameNoteHandler,
	ProvideRenameWorkspaceHandler,
	ProvideRestoreTrashedWorkspaceItemsHandler,
	ProvideTrashWorkspaceItemsHandler,
	ProvideUnpublishNoteHandler,
	ProvideUnpublishWorkspaceHandler,
	ProvideUpdateWorkspaceMembersHandler,
)

var ProviderSetEvent = wire.NewSet(
	ProvideDocumentCommittedHandler,
	ProvideNoteCreatedDomainToIntegrationEventHandler,
	ProvideNoteUpdatedDomainToIntegrationEventHandler,
	ProvideNotifyWorkspaceItemsUpdatedHandler,
	ProvideNotifyWorkspaceRenamedHandler,
	ProvideNotifyWorkspaceSlugChangedHandler,
	wire.Struct(new(Events), "*"),
)

var ProviderSetQuery = wire.NewSet(
	ProvideQueries,

	ProvideCheckWorkspaceSlugExistsHandler,
	ProvideGetMyWorkspacesHandler,
	ProvideGetNoteGraphHandler,
	ProvideGetNoteLinksHandler,
	ProvideGetWorkspaceByNoteHandler,
	ProvideGetWorkspaceBySlugHandler,
	ProvideGetWorkspaceGraphHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideGetWorkspaceSearchTokenHandler,
	ProvideGetWorkspaceTreeHandler,
	ProvideShowTrashHandler,
)

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProviderSetCommand,
	ProviderSetEvent,
	ProviderSetQuery,
	wire.Struct(new(Server), "*"),
)
