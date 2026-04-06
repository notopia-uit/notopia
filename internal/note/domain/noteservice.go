package domain

import (
	"encoding/json"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type NoteService struct{}

func NewNoteService() *NoteService {
	return &NoteService{}
}

var ProvideNoteService = NewNoteService

func (s *NoteService) UpdateNoteSizeBasedOnContent(note *Note, content any, userID string) error {
	b, err := json.Marshal(content)
	if err != nil {
		return errs.NewNoteFailToMarshalDocumentContent(note.ID(), content, err)
	}
	note.SetSize(uint64(len(b)), userID)
	return nil
}
