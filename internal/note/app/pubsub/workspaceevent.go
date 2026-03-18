package pubsub

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type WorkspaceEvent interface {
	Publish(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
		events ...domain.WorkspaceEvent,
	) error
	Subscribe(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
	) (<-chan domain.WorkspaceEvent, error)

	Run(ctx context.Context) error

	Close() error
}
