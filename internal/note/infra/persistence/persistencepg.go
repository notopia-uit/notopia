package persistence

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func NewGooseProvider(db *sql.DB, logger *slog.Logger) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres session locker: %w", err)
	}
	migrationFiles, err := fs.Sub(PgMigrations, "pgmigration")
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}
	gooseProvider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationFiles,
		goose.WithSlog(logger),
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create goose provider: %w", err)
	}
	return gooseProvider, nil
}

var ProvideGooseProvider = NewGooseProvider

//go:embed pgmigration/*
var PgMigrations embed.FS

type Pg struct {
	pgxPool       *pgxpool.Pool
	gooseProvider *goose.Provider
}

func NewPg(
	pgxPool *pgxpool.Pool,
	gooseProvider *goose.Provider,
) (*Pg, error) {
	return &Pg{
		pgxPool:       pgxPool,
		gooseProvider: gooseProvider,
	}, nil
}

func (p *Pg) IsMigrationDone(ctx context.Context) (bool, error) {
	pending, err := p.gooseProvider.HasPending(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return !pending, nil
}

func (p *Pg) Ping(ctx context.Context) error {
	return p.pgxPool.Ping(ctx)
}

func (p *Pg) RunMigrations(ctx context.Context) error {
	_, err := p.gooseProvider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

var ProvidePg = NewPg
