package app

import (
	"context"
	"fmt"
	"log/slog"

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
	slog.DebugContext(ctx, "Handling note updated integration event", slog.String("note_id", noteID.String()))
	note, err := h.getNoteReadModel.GetNote(ctx, &GetNoteReadModelParams{
		ID:             noteID,
		ExcludeTrashed: false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to get note for note updated event", slog.String("note_id", noteID.String()), slog.Any("error", err))
		return fmt.Errorf("failed to get note for note updated event: %w", err)
	}
	folder, err := h.getFolderReadModel.GetFolder(ctx, &GetFolderReadModelParams{
		ID:             note.FolderID,
		ExcludeTrashed: false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to get folder for note updated event", slog.String("note_id", noteID.String()), slog.Any("error", err))
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
		slog.ErrorContext(ctx, "failed to publish note updated integration event", slog.String("note_id", noteID.String()), slog.Any("error", err))
		return fmt.Errorf("failed to publish note updated integration event: %w", err)
	}
	slog.InfoContext(ctx, "Note updated integration event published", slog.String("note_id", noteID.String()))
	return nil
}
