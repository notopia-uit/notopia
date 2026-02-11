package persistence

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed pgmigration/*
var PgMigrations embed.FS

type PersistencePg struct {
	db     *sql.DB
	pgpool *pgxpool.Pool
	goose  *goose.Provider
}

var _ Persistence = (*PersistencePg)(nil)

func NewGooseProvider(db *sql.DB, logger *slog.Logger) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres session locker: %w", err)
	}
	gooseProvider, err := goose.NewProvider(
		"postgres",
		db,
		PgMigrations,
		goose.WithSlog(logger),
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create goose provider: %w", err)
	}
	return gooseProvider, nil
}

var ProvideGooseProvider = NewGooseProvider

func NewPg(
	db *sql.DB,
	pgxpool *pgxpool.Pool,
	gooseProvider *goose.Provider,
) (*PersistencePg, error) {
	return &PersistencePg{
		db:     db,
		pgpool: pgxpool,
		goose:  gooseProvider,
	}, nil
}

func (p *PersistencePg) RunMigrations(ctx context.Context) error {
	_, err := p.goose.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (p *PersistencePg) CheckReadiness(ctx context.Context) error {
	return p.pgpool.Ping(ctx)
}

var ProvidePg = NewPg
