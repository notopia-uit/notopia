package domain

import (
	"fmt"

	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodeFolderNotFound  = "folder_1"
	ErrCodeEmptyFolderName = "folder_2"
)

func NewErrFolderNotFound(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Folder with id %q not found", id.String()),
		ErrCodeFolderNotFound,
		err,
	)
}

var ErrEmptyFolderName = commonerror.NewInvalid(
	"Folder name cannot be empty",
	ErrCodeEmptyFolderName,
	nil,
)
