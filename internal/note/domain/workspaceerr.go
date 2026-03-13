package domain

import (
	"fmt"

	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodeWorkspaceNotFound           = "Workspace_1"
	ErrCodeWorkspaceRootFolderNotFound = "Workspace_2"
)

func NewErrWorkspaceNotFound(slug string, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Workspace with slug %q not found", slug),
		ErrCodeWorkspaceNotFound,
		err,
	)
}

func NewErrWorkspaceRootFolderNotFound(slug string, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Root folder for workspace with slug %q not found", slug),
		ErrCodeWorkspaceRootFolderNotFound,
		err,
	)
}
