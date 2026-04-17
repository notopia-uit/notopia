package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	ProvideChangeWorkspaceSlugHandler,
	ProvideCreateFolderHandler,
	ProvideCreateNoteHandler,
	ProvideCreateWorkspaceHandler,
	ProvidePermanentlyDeleteFolderHandler,
	ProvidePermanentlyDeleteNoteHandler,
	ProvideDeleteWorkspaceHandler,
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

var ProviderSetEvent = wire.NewSet(
	ProvideDocumentCommittedHandler,
	ProvideNotifyWorkspaceItemsUpdatedHandler,
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
