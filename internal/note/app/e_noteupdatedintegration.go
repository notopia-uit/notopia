package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// This is used for register multiple domain event
type NoteUpdatedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
	getNoteReadModel     GetNoteReadModel
	getFolderReadModel   GetFolderReadModel
}

func NewNoteUpdatedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
	getNoteReadModel GetNoteReadModel,
	getFolderReadModel GetFolderReadModel,
) *NoteUpdatedDomainToIntegrationEventHandler {
	return &NoteUpdatedDomainToIntegrationEventHandler{
		integrationPublisher: integrationPublisher,
		getNoteReadModel:     getNoteReadModel,
		getFolderReadModel:   getFolderReadModel,
	}
}

var ProvideNoteUpdatedDomainToIntegrationEventHandler = NewNoteUpdatedDomainToIntegrationEventHandler

func (h *NoteUpdatedDomainToIntegrationEventHandler) Handle(ctx context.Context, noteID uuid.UUID) error {
	note, err := h.getNoteReadModel.GetNote(ctx, &GetNoteReadModelParams{
		ID:             noteID,
		ExcludeTrashed: false,
	})
	if err != nil {
		return fmt.Errorf("failed to get note for note updated event: %w", err)
	}
	folder, err := h.getFolderReadModel.GetFolder(ctx, &GetFolderReadModelParams{
		ID:             note.FolderID,
		ExcludeTrashed: false,
	})
	if err != nil {
		return fmt.Errorf("failed to get note or folder for note updated event: %w", err)
	}
	integrationEvent := IntegrationEventNoteUpdated{
		ID:          noteID,
		WorkspaceID: folder.WorkspaceID,
		Icon:        note.Icon,
		Name:        note.Name,
		Size:        uint64(note.Size),
		FolderID:    note.FolderID,
		FolderName:  folder.Name,
		Trashed:     note.Trashed,
		UpdatedAt:   note.UpdatedAt,
	}
	if err := h.integrationPublisher.Publish(ctx, integrationEvent); err != nil {
		return fmt.Errorf("failed to publish note updated integration event: %w", err)
	}
	return nil
}
