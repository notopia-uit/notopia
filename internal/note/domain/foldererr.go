package domain

import (
	"fmt"

	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodeFolderNotFound                     = "folder_1"
	ErrCodeFolderNameAlreadyExistsInWorkspace = "folder_2"
)

func NewErrFolderNotFound(id uuid.UUID) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Folder with id %q not found", id.String()),
		ErrCodeFolderNotFound,
		nil,
	)
}

func NewErrFolderNameAlreadyExistsInWorkspace(name string, workspaceName string) *commonerror.Err {
	return commonerror.NewConflict(
		fmt.Sprintf("Folder with name %q already exists in workspace %q", name, workspaceName),
		ErrCodeFolderNameAlreadyExistsInWorkspace,
		nil,
	)
}
