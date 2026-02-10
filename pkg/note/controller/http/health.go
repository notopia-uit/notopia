package http

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/pkg/common/controller/http"
)

func NewHealthManager(
	pg *pgxpool.Pool,
) *http.HealthManager {
	postgresCheck := func(ctx context.Context) error {
		return pg.Ping(ctx)
	}
	healthToCheck := map[string]http.HealthToCheckFunc{
		"postgres": postgresCheck,
	}
	return http.NewHealthManager(healthToCheck)
}

var ProvideHealthManager = NewHealthManager
