package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pg"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

var PostresReadModelProviderSet = wire.NewSet(
	pg.ProvideGetWorkspaceTreeReadModel,
	pg.ProvideShowTrashReadModel,
	pg.ProvideGetNoteLinksReadModel,
	pg.ProvideGetWorkspaceBySlugReadModel,
	pg.ProvideCheckWorkspaceSlugExistsReadModel,
	pg.ProvideGetWorkspaceGraphReadModel,
	pg.ProvideGetNoteGraphReadModel,
	pg.ProvideGetNoteReadModel,
	wire.Bind(new(app.CheckWorkspaceSlugExistsReadModel), new(*pg.CheckWorkspaceSlugExistsReadModel)),
	wire.Bind(new(app.GetNoteGraphReadModel), new(*pg.GetNoteGraphReadModel)),
	wire.Bind(new(app.GetNoteLinksReadModel), new(*pg.GetNoteLinksReadModel)),
	wire.Bind(new(app.GetNoteReadModel), new(*pg.GetNoteReadModel)),
	wire.Bind(new(app.GetWorkspaceBySlugReadModel), new(*pg.GetWorkspaceBySlugReadModel)),
	wire.Bind(new(app.GetWorkspaceGraphReadModel), new(*pg.GetWorkspaceGraphReadModel)),
	wire.Bind(new(app.GetWorkspaceTreeReadModel), new(*pg.GetWorkspaceTreeReadModel)),
	wire.Bind(new(app.ShowTrashReadModel), new(*pg.ShowTrashReadModel)),
)

var PostgresRepoProviderSet = wire.NewSet(
	pg.ProvideFolderRepo,
	pg.ProvideNoteRepo,
	pg.ProvideUnitOfWork,
	pg.ProvideWorkspaceRepo,
	wire.Bind(new(domain.FolderRepo), new(*pg.FolderRepo)),
	wire.Bind(new(domain.NoteRepo), new(*pg.NoteRepo)),
	wire.Bind(new(domain.UnitOfWork), new(*pg.UnitOfWork)),
	pg.ProvidePublisherFactory,
	wire.Bind(new(domain.WorkspaceRepo), new(*pg.WorkspaceRepo)),
)

var PostgresProviderSet = wire.NewSet(
	PostresReadModelProviderSet,
	PostgresRepoProviderSet,

	ProvideGooseProvider,
	ProvidePg,
	pg.ProvidePgPool,
	pg.ProvideQueries,
	pg.ProvideStdlib,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
)
