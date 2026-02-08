package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pgsqlc"
)

func NewDBTXPool(ctx context.Context, cfg *config.SQL) (pgsqlc.DBTX, error) {
	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

var ProvideDBTXPool = NewDBTXPool

func NewQueries(db pgsqlc.DBTX) *pgsqlc.Queries {
	return pgsqlc.New(db)
}

var ProvideQueries = NewQueries
