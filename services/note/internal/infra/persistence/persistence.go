package persistence

import "context"

type Persistence interface {
	RunMigrations() error
	CheckReadiness(ctx context.Context) error
}
