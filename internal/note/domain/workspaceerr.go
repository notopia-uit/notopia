package domain

import (
	"fmt"

	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodeWorkspaceNotFound           = "Workspace_1"
	ErrCodeWorkspaceRootFolderNotFound = "Workspace_2"
	ErrCodeInvalidWorkspaceName        = "Workspace_3"
	ErrCodeInvalidWorkspaceSlug        = "Workspace_4"
)

func NewErrWorkspaceBySlugNotFound(slug string, err error) *commonerror.Err {
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

func NewErrEmptyWorkspaceName() *commonerror.Err {
	return commonerror.NewInvalid(
		"Workspace name cannot be empty",
		ErrCodeInvalidWorkspaceName,
		nil,
	)
}

func NewErrEmptyWorkspaceSlug() *commonerror.Err {
	return commonerror.NewInvalid(
		"Workspace slug cannot be empty",
		ErrCodeInvalidWorkspaceSlug,
		nil,
	)
}
