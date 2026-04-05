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
	pg.ProvideReadModel,
	wire.Bind(new(app.CheckWorkspaceSlugExistsReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetNoteGraphReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetNoteLinksReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetNoteReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetWorkspaceBySlugReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetWorkspaceGraphReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.GetWorkspaceTreeReadModel), new(*pg.ReadModel)),
	wire.Bind(new(app.ShowTrashReadModel), new(*pg.ReadModel)),
)

var PostgresProviderSet = wire.NewSet(
	PostresReadModelProviderSet,

	ProvideGooseProvider,
	ProvidePg,
	pg.ProvideFolder,
	pg.ProvideNote,
	pg.ProvidePgPool,
	pg.ProvideQueries,
	pg.ProvideUnitOfWork,
	pg.ProvideWorkspace,
	wire.Bind(new(app.Persistence), new(*Pg)),
	wire.Bind(new(domain.FolderRepo), new(*pg.Folder)),
	wire.Bind(new(domain.NoteRepo), new(*pg.Note)),
	wire.Bind(new(domain.UnitOfWork), new(*pg.UnitOfWork)),
	wire.Bind(new(domain.WorkspaceRepo), new(*pg.Workspace)),
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
)
