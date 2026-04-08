package event

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
)

func (e *Event) notifyWorkspaceItemsUpdatedNoteHandler(msg *message.Message) error {
	noteIDStr := msg.Metadata.Get(e.domainEventCfg.MessageMetadataAggregateIDKey)
	if noteIDStr == "" {
		return errors.New("missing note id in message metadata in notifyWorkspaceItemsUpdatedNoteHandler")
	}
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		return errors.New("invalid note id in message metadata in notifyWorkspaceItemsUpdatedNoteHandler")
	}
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
	if err := e.app.Events.NotifyWorkspaceItemsUpdated.Handle(&app.NotifyWorkspaceItemsUpdated{
		WorkspaceItemID: noteID,
		WorkspaceID:     workspaceID,
		UserID:          userID,
		Type:            app.NotifyWorkspaceItemsUpdatedTypeNote,
	}); err != nil {
		return fmt.Errorf("failed to handle NotifyWorkspaceItemsUpdated in notifyWorkspaceItemsUpdatedNoteHandler: %w", err)
	}
	return nil
}

func (e *Event) notifyWorkspaceItemsUpdatedFolderHandler(msg *message.Message) error {
	folderIDStr := msg.Metadata.Get(e.domainEventCfg.MessageMetadataAggregateIDKey)
	if folderIDStr == "" {
		return errors.New("missing folder id in message metadata in notifyWorkspaceItemsUpdatedFolderHandler")
	}
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		return errors.New("invalid folder id in message metadata in notifyWorkspaceItemsUpdatedFolderHandler")
	}
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
	if err := e.app.Events.NotifyWorkspaceItemsUpdated.Handle(&app.NotifyWorkspaceItemsUpdated{
		WorkspaceItemID: folderID,
		WorkspaceID:     workspaceID,
		UserID:          userID,
		Type:            app.NotifyWorkspaceItemsUpdatedTypeFolder,
	}); err != nil {
		return fmt.Errorf("failed to handle NotifyWorkspaceItemsUpdated in notifyWorkspaceItemsUpdatedFolderHandler: %w", err)
	}
	return nil
}
