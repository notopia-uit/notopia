package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/healthmanager"
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
