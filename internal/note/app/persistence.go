package app

import "context"

type Persistence interface {
	IsMigrationDone(ctx context.Context) (bool, error)
	Ping(ctx context.Context) error
	RunMigrations(ctx context.Context) error
	CheckReadiness(ctx context.Context) error
}
