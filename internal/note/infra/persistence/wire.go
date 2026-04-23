package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgreadmodel"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgrepo"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

var PGReadModelProviderSet = wire.NewSet(
	pgreadmodel.ProvideCheckWorkspaceSlugExists,
	pgreadmodel.ProvideGetFolder,
	pgreadmodel.ProvideGetWorkspaceByNote,
	pgreadmodel.ProvideGetWorkspaces,
	pgreadmodel.ProvideNote,
	pgreadmodel.ProvideNoteGraph,
	pgreadmodel.ProvideNoteLinks,
	pgreadmodel.ProvideShowTrash,
	pgreadmodel.ProvideWorkspaceBySlug,
	pgreadmodel.ProvideWorkspaceGraph,
	pgreadmodel.ProvideWorkspaceTree,
	wire.Bind(new(app.CheckWorkspaceSlugExistsReadModel), new(*pgreadmodel.CheckWorkspaceSlugExists)),
	wire.Bind(new(app.GetFolderReadModel), new(*pgreadmodel.GetFolder)),
	wire.Bind(new(app.GetNoteGraphReadModel), new(*pgreadmodel.NoteGraph)),
	wire.Bind(new(app.GetNoteLinksReadModel), new(*pgreadmodel.NoteLinks)),
	wire.Bind(new(app.GetNoteReadModel), new(*pgreadmodel.Note)),
	wire.Bind(new(app.GetWorkspaceByNoteReadModel), new(*pgreadmodel.GetWorkspaceByNote)),
	wire.Bind(new(app.GetWorkspaceGraphReadModel), new(*pgreadmodel.WorkspaceGraph)),
	wire.Bind(new(app.GetWorkspaceTreeReadModel), new(*pgreadmodel.WorkspaceTree)),
	wire.Bind(new(app.GetWorkspacesReadModel), new(*pgreadmodel.GetWorkspaces)),
	wire.Bind(new(app.ShowTrashReadModel), new(*pgreadmodel.ShowTrash)),
	wire.Bind(new(app.WorkspaceBySlugReadModel), new(*pgreadmodel.WorkspaceBySlug)),
)

var PGRepoProviderSet = wire.NewSet(
	pgrepo.ProvideFolder,
	pgrepo.ProvideNote,
	pgrepo.ProvideUnitOfWork,
	pgrepo.ProvideWorkspace,
	pgrepo.ProvideRunInTx,
	wire.Bind(new(domain.FolderRepo), new(*pgrepo.Folder)),
	wire.Bind(new(domain.NoteRepo), new(*pgrepo.Note)),
	wire.Bind(new(domain.UnitOfWork), new(*pgrepo.UnitOfWork)),
	wire.Bind(new(domain.WorkspaceRepo), new(*pgrepo.Workspace)),
)

var PostgresProviderSet = wire.NewSet(
	PGRepoProviderSet,
	PGReadModelProviderSet,

	ProvideGooseProvider,
	ProvidePg,
	ProvidePgPool,
	ProvideSQLCQueries,
	ProvidePgxPoolStdlib,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
)
