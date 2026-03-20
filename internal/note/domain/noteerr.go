package domain

import (
	"fmt"

	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodeNoteNotFound                     = "Note_1"
	ErrCodeNoteFailToMarshalDocumentContent = "Note_2"
)

func NewErrNoteNotFound(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Note with id %q not found", id.String()),
		ErrCodeNoteNotFound,
		err,
	)
}

func NewErrNoteFailToMarshalContent(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewInvalid(
		fmt.Sprintf("Failed to marshal document content for note %q", id.String()),
		ErrCodeNoteFailToMarshalDocumentContent,
		err,
	)
}
