package event

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func (e *Event) documentCommittedHandler(ctx context.Context, event *share.DocumentCommittedEvent) error {
	ev := &app.DocumentCommitted{
		ID:              event.Id,
		Content:         event.Content,
		Tags:            event.Tags,
		OutgoingLinkIDs: event.OutgoingLinkIds,
		UserID:          event.UserId,
	}
	return e.app.IntegrationEvents.DocumentCommittedHandler.Handle(ctx, ev)
}
