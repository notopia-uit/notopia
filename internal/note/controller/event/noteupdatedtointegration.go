package event

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
)

func (e *Event) noteUpdatedToIntegrationHandler(msg *message.Message) error {
	noteIDStr := msg.Metadata.Get(e.domainEventCfg.MessageMetadataAggregateIDKey)
	if noteIDStr == "" {
		return errors.New("missing note id in message metadata in noteUpdatedToIntegrationHandler")
	}
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		return errors.New("invalid note id in message metadata in noteUpdatedToIntegrationHandler")
	}
	ctx := msg.Context()
	if err := e.app.Events.NoteUpdatedDomainToIntegrationEvent.Handle(ctx, noteID); err != nil {
		return fmt.Errorf("failed to handle note updated domain to integration event: %w", err)
	}
	return nil
}
