package persistence

import (
	"database/sql"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/app/query"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pg"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

var PostgresProviderSet = wire.NewSet(
	ProvidePg,
	ProvideGooseProvider,
	pg.ProvidePgPool,
	pg.ProvideQueries,
	pg.ProvideStdlib,
	pg.ProvideReadModel,
	pg.ProvideNote,
	pg.ProvideFolder,
	pg.ProvideWorkspace,
	pg.ProvideUnitOfWork,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
	wire.Bind(new(qrm.DB), new(*sql.DB)),
	wire.Bind(new(app.Persistence), new(*Pg)),
	wire.Bind(new(domain.NoteRepo), new(*pg.Note)),
	wire.Bind(new(domain.FolderRepo), new(*pg.Folder)),
	wire.Bind(new(domain.WorkspaceRepo), new(*pg.Workspace)),
	wire.Bind(new(domain.UnitOfWork), new(*pg.UnitOfWork)),
	wire.Bind(new(query.CheckWorkspaceSlugExistsReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetNoteGraphReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetNoteLinksReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetNoteReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetWorkspaceBySlugReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetWorkspaceGraphReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.GetWorkspaceTreeReadModel), new(*pg.ReadModel)),
	wire.Bind(new(query.ShowTrashReadModel), new(*pg.ReadModel)),
)
