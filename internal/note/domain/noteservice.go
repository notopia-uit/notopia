package domain

import "encoding/json"

type NoteService struct{}

func NewNoteService() *NoteService {
	return &NoteService{}
}

func (s *NoteService) UpdateNoteSizeBasedOnDocumentContent(note *Note, content any) error {
	b, err := json.Marshal(content)
	if err != nil {
		return NewErrNoteFailToMarshalDocumentContent(note.ID(), err)
	}
	note.SetSize(uint(len(b)))
	return nil
}
