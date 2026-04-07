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
	pgreadmodel.ProvideGetWorkspaceTree,
	pgreadmodel.ProvideShowTrash,
	pgreadmodel.ProvideGetNoteLinks,
	pgreadmodel.ProvideGetWorkspaceBySlug,
	pgreadmodel.ProvideCheckWorkspaceSlugExists,
	pgreadmodel.ProvideGetWorkspaceGraph,
	pgreadmodel.ProvideGetNoteGraph,
	pgreadmodel.ProvideGetNote,
	wire.Bind(new(app.CheckWorkspaceSlugExistsReadModel), new(*pgreadmodel.CheckWorkspaceSlugExists)),
	wire.Bind(new(app.GetNoteGraphReadModel), new(*pgreadmodel.GetNoteGraph)),
	wire.Bind(new(app.GetNoteLinksReadModel), new(*pgreadmodel.GetNoteLinks)),
	wire.Bind(new(app.GetNoteReadModel), new(*pgreadmodel.GetNote)),
	wire.Bind(new(app.GetWorkspaceBySlugReadModel), new(*pgreadmodel.GetWorkspaceBySlug)),
	wire.Bind(new(app.GetWorkspaceGraphReadModel), new(*pgreadmodel.GetWorkspaceGraph)),
	wire.Bind(new(app.GetWorkspaceTreeReadModel), new(*pgreadmodel.GetWorkspaceTree)),
	wire.Bind(new(app.ShowTrashReadModel), new(*pgreadmodel.ShowTrash)),
)

var PGRepoProviderSet = wire.NewSet(
	pgrepo.ProvideFolder,
	pgrepo.ProvideNote,
	pgrepo.ProvideUnitOfWork,
	pgrepo.ProvideWorkspace,
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
