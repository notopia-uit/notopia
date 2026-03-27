package query

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideCheckWorkspaceSlugExistsHandler,
	ProvideGetNoteGraphHandler,
	ProvideGetNoteLinksHandler,
	ProvideGetWorkspaceBySlugHandler,
	ProvideGetWorkspaceGraphHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideGetWorkspaceTreeHandler,
	ProvideShowTrashHandler,
	ProvideGetNoteHandler,
)
