package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	wire.Struct(new(CommandHandlers), "*"),
	ProvideCreateNoteHandler,
	ProvideCreateFolderHandler,
	ProvideCreateWorkspaceHandler,
	ProvideDeleteNoteHandler,
	ProvideDeleteFolderHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideGenerateDailyNoteHandler,
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
)

var ProviderSetQuery = wire.NewSet(
	wire.Struct(new(QueryHandlers), "*"),
	ProvideCheckWorkspaceSlugExistsHandler,
	ProvideGetNoteGraphHandler,
	ProvideGetNoteLinksHandler,
	ProvideGetWorkspaceBySlugHandler,
	ProvideGetWorkspaceGraphHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideGetWorkspaceTreeHandler,
	ProvideShowTrashHandler,
)

var ProviderSetIntegrationEvent = wire.NewSet(
	wire.Struct(new(IntegrationEventHandlers), "*"),
	ProvideDocumentCommittedHandler,
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	ProviderSetQuery,
	ProviderSetIntegrationEvent,
	wire.Struct(new(Server), "*"),
)
