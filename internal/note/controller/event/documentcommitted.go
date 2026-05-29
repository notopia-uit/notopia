package event

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func (e *Event) documentCommittedHandler(msg *message.Message) error {
	var event share.DocumentCommittedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal DocumentCommittedEvent: %w", err)
	}
	ev := &app.DocumentCommitted{
		ID:              event.Id,
		Content:         event.Content,
		Tags:            event.Tags,
		OutgoingLinkIDs: event.OutgoingLinkIds,
		UserID:          event.UserId,
	}
	return e.app.Events.DocumentCommitted.Handle(msg.Context(), ev)
}
