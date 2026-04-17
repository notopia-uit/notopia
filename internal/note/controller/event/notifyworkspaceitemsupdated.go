package event

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
)

func (e *Event) notifyWorkspaceItemsUpdatedNoteHandler(msg *message.Message) error {
	workspaceIDStr := msg.Metadata.Get(e.domainEventCfg.MessageWorkspaceIDKey)
	if workspaceIDStr == "" {
		return errors.New("missing workspace id in message metadata in notifyWorkspaceItemsUpdatedNoteHandler")
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return errors.New("invalid workspace id in message metadata in notifyWorkspaceItemsUpdatedNoteHandler")
	}
	userID := msg.Metadata.Get(e.domainEventCfg.MessageMetadataUserIDKey)
	if userID == "" {
		return errors.New("missing user id in message metadata in notifyWorkspaceItemsUpdatedNoteHandler")
	}
	if err := e.app.Events.NotifyWorkspaceItemsUpdatedHandler.Handle(&app.NotifyWorkspaceItemsUpdated{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to handle NotifyWorkspaceItemsUpdated in notifyWorkspaceItemsUpdatedNoteHandler: %w", err)
	}
	return nil
}

func (e *Event) notifyWorkspaceItemsUpdatedFolderHandler(msg *message.Message) error {
	workspaceIDStr := msg.Metadata.Get(e.domainEventCfg.MessageWorkspaceIDKey)
	if workspaceIDStr == "" {
		return errors.New("missing workspace id in message metadata in notifyWorkspaceItemsUpdatedFolderHandler")
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return errors.New("invalid workspace id in message metadata in notifyWorkspaceItemsUpdatedFolderHandler")
	}
	userID := msg.Metadata.Get(e.domainEventCfg.MessageMetadataUserIDKey)
	if userID == "" {
		return errors.New("missing user id in message metadata in notifyWorkspaceItemsUpdatedFolderHandler")
	}
	if err := e.app.Events.NotifyWorkspaceItemsUpdatedHandler.Handle(&app.NotifyWorkspaceItemsUpdated{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to handle NotifyWorkspaceItemsUpdated in notifyWorkspaceItemsUpdatedFolderHandler: %w", err)
	}
	return nil
}
