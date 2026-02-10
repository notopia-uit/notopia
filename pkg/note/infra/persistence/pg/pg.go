package pg

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pgsqlc"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewDBTXPool(
	ctx context.Context,
	tracerProvider *trace.TracerProvider,
	cfg *config.SQL,
) (pgsqlc.DBTX, func(), error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.URL)
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

var ProvideDBTXPool = NewDBTXPool

func NewQueries(db pgsqlc.DBTX) *pgsqlc.Queries {
	return pgsqlc.New(db)
}

var ProvideQueries = NewQueries
