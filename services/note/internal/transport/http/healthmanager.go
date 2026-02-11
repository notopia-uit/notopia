package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/healthmanager"
	"github.com/notopia-uit/notopia/services/note/internal/infra/persistence"
)

func NewHealthManager(
	persistence persistence.Persistence,
) *healthmanager.HealthManager {
	persistenceCheck := func(ctx context.Context) error {
		return persistence.CheckReadiness(ctx)
	}
	healthToCheck := map[string]healthmanager.ToCheckFunc{
		"persistence": persistenceCheck,
	}
	return healthmanager.New(healthToCheck)
}

var ProvideHealthManager = NewHealthManager
