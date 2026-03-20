package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	ProvideCreateNoteHandler,
	ProvideCreateFolderHandler,
	ProvideCreateWorkspaceHandler,
	ProvideDeleteNoteHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideGenerateDailyNoteHandler,
	ProvideMoveWorkspaceItemsHandler,
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
	ProvideCheckWorkspaceExistsHandler,
	ProvideGetNoteGraphHandler,
	ProvideGetNoteLinksHandler,
	ProvideGetNotesHandler,
	ProvideGetWorkspaceHandler,
	ProvideGetWorkspaceGraphHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideGetWorkspaceTreeHandler,
	ProvideShowTrashHandler,
)

var ProviderSetEvent = wire.NewSet(
	ProvideDocumentCommittedHandler,
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	ProviderSetQuery,
	ProviderSetEvent,
	ProvideApp,
)
