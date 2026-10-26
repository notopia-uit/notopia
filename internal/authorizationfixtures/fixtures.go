package authorizationfixtures

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/notopia-uit/notopia/internal/authorization/policy"
)

type Fixtures struct {
	enforcer *casbin.TransactionalEnforcer
	logger   *slog.Logger
}

func NewFixtures(enforcer *casbin.TransactionalEnforcer, logger *slog.Logger) *Fixtures {
	return &Fixtures{
		enforcer: enforcer,
		logger:   logger,
	}
}

var ProvideFixtures = NewFixtures

// Run is add-only and assumes the policy matrix was already seeded.
func (f *Fixtures) Run(ctx context.Context) error {
	if err := policy.LoadFixtures(f.enforcer); err != nil {
		return err
	}
	f.logger.InfoContext(ctx, "authorization fixture policies loaded")
	return nil
}
