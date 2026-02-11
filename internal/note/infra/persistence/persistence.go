package persistence

import "context"

type Persistence interface {
	RunMigrations(ctx context.Context) error
	CheckReadiness(ctx context.Context) error
}
