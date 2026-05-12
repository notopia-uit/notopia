package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
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
	wire.Struct(new(Cmds), "*"),
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
	wire.Struct(new(Queries), "*"),
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	ProviderSetEvent,
	ProviderSetQuery,
	wire.Struct(new(Server), "*"),
)
