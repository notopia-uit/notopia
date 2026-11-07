package persistence

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/common/cli"
	"github.com/pressly/goose/v3"
)

type Migrator struct {
	provider *goose.Provider
	logger   *slog.Logger
}

func NewMigrator(provider *goose.Provider, logger *slog.Logger) *Migrator {
	return &Migrator{provider, logger}
}

func (m *Migrator) Up(ctx context.Context) error {
	_, err := m.provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	_, err := m.provider.Down(ctx)
	if err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func (m *Migrator) Status(ctx context.Context) error {
	statuses, err := m.provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("migration status: %w", err)
	}
	for _, status := range statuses {
		m.logger.InfoContext(
			ctx,
			"migration",
			slog.Int64("version", status.Source.Version),
			slog.Any("state", status.State),
		)
	}
	return nil
}

var _ cli.Migrator = (*Migrator)(nil)
