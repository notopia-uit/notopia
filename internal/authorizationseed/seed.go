package authorizationseed

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/notopia-uit/notopia/internal/authorization/policy"
)

type Seeder struct {
	enforcer *casbin.TransactionalEnforcer
	logger   *slog.Logger
}

func NewSeeder(enforcer *casbin.TransactionalEnforcer, logger *slog.Logger) *Seeder {
	return &Seeder{
		enforcer: enforcer,
		logger:   logger,
	}
}

var ProvideSeeder = NewSeeder

// Run is idempotent: missing rules are added and stale p rules are pruned.
func (s *Seeder) Run(ctx context.Context) error {
	if err := policy.Seed(s.enforcer); err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "authorization policy matrix seeded")
	return nil
}
