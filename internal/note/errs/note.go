package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeNoteNotFound                     Code = "note_1"
	CodeNoteFailToMarshalDocumentContent Code = "note_2"
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
