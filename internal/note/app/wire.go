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
	wire.Struct(new(CommandHandlers), "*"),
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
	wire.Struct(new(QueryHandlers), "*"),
)

var ProviderSetIntegrationEvent = wire.NewSet(
	ProvideDocumentCommittedHandler,
	wire.Struct(new(IntegrationEventHandlers), "*"),
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	ProviderSetIntegrationEvent,
	ProviderSetQuery,
	wire.Struct(new(Server), "*"),
)
