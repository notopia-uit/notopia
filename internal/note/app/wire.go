package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	ProvideCreateFolderHandler,
	ProvideCreateNoteHandler,
	ProvideCreateWorkspaceHandler,
	ProvidePermanentlyDeleteFolderHandler,
	ProvidePermanentlyDeleteNoteHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideGenerateDailyNoteHandler,
	ProvideGetNoteHandler,
	ProvideMoveWorkspaceItemsHandler,
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

var ProviderSetQuery = wire.NewSet(
	ProvideCheckWorkspaceSlugExistsHandler,
	ProvideGetNoteGraphHandler,
	ProvideGetNoteLinksHandler,
	ProvideGetWorkspaceBySlugHandler,
	ProvideGetWorkspaceGraphHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideGetWorkspaceTreeHandler,
	ProvideShowTrashHandler,
	wire.Struct(new(Queries), "*"),
)

var ProviderSetIntegrationEvent = wire.NewSet(
	ProvideDocumentCommittedHandler,
	wire.Struct(new(IntegrationEvents), "*"),
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	ProviderSetIntegrationEvent,
	ProviderSetQuery,
	wire.Struct(new(Server), "*"),
)
