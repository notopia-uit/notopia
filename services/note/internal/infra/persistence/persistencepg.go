package persistence

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

//go:embed pgmigration/*
var PgMigrations embed.FS

type PersistencePg struct {
	db     *sql.DB
	pgpool *pgxpool.Pool
}

var _ Persistence = (*PersistencePg)(nil)

func NewPg(
	db *sql.DB,
	pgxpool *pgxpool.Pool,
) (*PersistencePg, error) {
	goose.SetBaseFS(PgMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	return &PersistencePg{
		db:     db,
		pgpool: pgxpool,
	}, nil
}

func (p *PersistencePg) RunMigrations() error {
	return goose.Up(p.db, "pgmigration")
}

func (p *PersistencePg) CheckReadiness(ctx context.Context) error {
	return p.pgpool.Ping(ctx)
}

var ProvidePg = NewPg
