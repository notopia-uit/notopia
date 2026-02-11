package pg

import (
	"context"
	"database/sql"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/services/note/internal/infra/persistence/pgsqlc"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewPgPool(
	ctx context.Context,
	tracerProvider *trace.TracerProvider,
	cfg *commonconfig.SQL,
) (*pgxpool.Pool, func(), error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.GetURL())
	if err != nil {
		return nil, nil, err
	}
	pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTracerProvider(tracerProvider),
	)
	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, nil, err
	}
	return pool, pool.Close, nil
}

var ProvidePgPool = NewPgPool

func NewQueries(db pgsqlc.DBTX) *pgsqlc.Queries {
	return pgsqlc.New(db)
}

var ProvideQueries = NewQueries

func NewStdlib(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDBFromPool(pool)
}

var ProvideStdlib = NewStdlib
