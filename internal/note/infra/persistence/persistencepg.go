package persistence

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed pgmigration/*
var PgMigrations embed.FS

type Pg struct {
	db            *sql.DB
	pgpool        *pgxpool.Pool
	gooseProvider *goose.Provider
}

var _ app.Persistence = (*Pg)(nil)

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
		"postgres",
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

func NewPg(
	db *sql.DB,
	pgxpool *pgxpool.Pool,
	gooseProvider *goose.Provider,
) (*Pg, error) {
	return &Pg{
		db:            db,
		pgpool:        pgxpool,
		gooseProvider: gooseProvider,
	}, nil
}

func (p *Pg) RunMigrations(ctx context.Context) error {
	_, err := p.gooseProvider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (p *Pg) CheckReadiness(ctx context.Context) error {
	return p.pgpool.Ping(ctx)
}

var ProvidePg = NewPg
