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
	ProvideEmptyTrashHandler,
	ProvideGetNoteHandler,
	ProvideLeaveWorkspaceHandler,
	ProvideMoveWorkspaceItemsHandler,
	ProvidePermanentlyDeleteFolderHandler,
	ProvidePermanentlyDeleteNoteHandler,
	ProvidePermanentlyDeleteWorkspaceItemsHandler,
	ProvidePublishNoteHandler,
	ProvidePurgeExpiredTrashHandler,
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
	ProvideBillingChangedHandler,
	ProvideDocumentCommittedHandler,
	ProvideNoteCreatedDomainToIntegrationEventHandler,
	ProvideNoteDeletedDomainToIntegrationEventHandler,
	ProvideNoteUpdatedDomainToIntegrationEventHandler,
	ProvideNotifyWorkspaceItemsUpdatedHandler,
	ProvideNotifyWorkspaceRenamedHandler,
	ProvideNotifyWorkspaceSlugChangedHandler,
	ProvideWorkspaceDeletedDomainToIntegrationEventHandler,
	wire.Struct(new(Events), "*"),
)

var ProviderSetQuery = wire.NewSet(
	ProvideQueries,

	ProvideBatchGetNoteWorkspacesHandler,
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
	ProvideSearchUsersHandler,
	ProvideShowTrashHandler,
)

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProviderSetCommand,
	ProviderSetEvent,
	ProviderSetQuery,
	wire.Struct(new(Server), "*"),
)
