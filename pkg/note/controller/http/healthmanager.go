package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/common/controller/http"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence"
)

func NewHealthManager(
	persistence persistence.Persistence,
) *http.HealthManager {
	postgresCheck := func(ctx context.Context) error {
		return persistence.CheckReadiness(ctx)
	}
	healthToCheck := map[string]http.HealthToCheckFunc{
		"postgres": postgresCheck,
	}
	return http.NewHealthManager(healthToCheck)
}

var ProvideHealthManager = NewHealthManager
