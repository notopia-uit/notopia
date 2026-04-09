package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeNoteNotFound                     Code = "noteNotFound"
	CodeNoteFailToMarshalDocumentContent Code = "noteFailToMarshalDocumentContent"
	CodeNoteAlreadyTrashed               Code = "noteAlreadyTrashed"
	CodeNotesNotInWorkspace              Code = "notesNotInWorkspace"
)

type NoteNotFound struct {
	Err
	NoteID uuid.UUID
}

func NewNoteNotFound(id uuid.UUID, err error) *NoteNotFound {
	return &NoteNotFound{
		NoteID: id,
		Err: Err{
			message: fmt.Sprintf("note with id %q not found", id.String()),
			code:    CodeNoteNotFound,
			err:     err,
		},
	}
}

type NoteFailToMarshalDocumentContent struct {
	Err
	NoteID  uuid.UUID
	Content any
}

func NewNoteFailToMarshalDocumentContent(id uuid.UUID, content any, err error) *NoteFailToMarshalDocumentContent {
	return &NoteFailToMarshalDocumentContent{
		NoteID:  id,
		Content: content,
		Err: Err{
			message: fmt.Sprintf("failed to marshal document content for note with id %q", id.String()),
			code:    CodeNoteFailToMarshalDocumentContent,
			err:     err,
		},
	}
}

type NoteAlreadyTrashed struct {
	Err
	NoteID uuid.UUID
}

func NewNoteAlreadyTrashed(id uuid.UUID) *NoteAlreadyTrashed {
	return &NoteAlreadyTrashed{
		NoteID: id,
		Err: Err{
			message: fmt.Sprintf("note with id %q is already trashed", id.String()),
			code:    CodeNoteAlreadyTrashed,
			err:     nil,
		},
	}
}

type NotesNotInWorkspace struct {
	Err
	WorkspaceID uuid.UUID
}

func NewNotesNotInWorkspace(workspaceID uuid.UUID) *NotesNotInWorkspace {
	return &NotesNotInWorkspace{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("one or more notes do not belong to workspace %s", workspaceID),
			code:    CodeNotesNotInWorkspace,
		},
	}
}
