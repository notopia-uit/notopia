package pubsub

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type WorkspaceEvent interface {
	Setup()
	Publish(ctx context.Context, workspaceID any, userID string, event domain.WorkspaceEvent) error
	Subscribe(ctx context.Context, workspaceID any, userID string) (<-chan domain.WorkspaceEvent, error)
	Run(ctx context.Context) error
	Close() error
}
