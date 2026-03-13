package domain

import (
	"fmt"

	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var ErrCodeFolderNotFound = "folder_1"

func NewErrFolderNotFound(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Folder with id %q not found", id.String()),
		ErrCodeFolderNotFound,
		err,
	)
}
