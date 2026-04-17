package event

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

func (e *Event) notifyWorkspaceRenamed(ctx context.Context, event *domain.WorkspaceRenamedEvent) error {
	if err := e.app.Events.NotifyWorkspaceRenamedHandler.Handle(ctx, &app.NotifyWorkspaceRenamed{
		WorkspaceID:   event.AggregateID,
		UserID:        event.UserID,
		Name:          event.Name,
		CorrelationID: event.ID,
	}); err != nil {
		return err
	}
	return nil
}
