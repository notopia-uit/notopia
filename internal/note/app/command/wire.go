package command

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
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
