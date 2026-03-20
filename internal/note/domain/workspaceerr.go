package domain

import (
	"fmt"

	"github.com/google/uuid"
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

func NewErrWorkspaceByIDNotFound(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Workspace with id %q not found", id.String()),
		ErrCodeWorkspaceNotFound,
		err,
	)
}

func NewErrWorkspaceRootFolderNotFound(id uuid.UUID, err error) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Root folder for workspace with id %q not found", id.String()),
		ErrCodeWorkspaceRootFolderNotFound,
		err,
	)
}

var ErrEmptyWorkspaceName = commonerror.NewInvalid(
	"Workspace name cannot be empty",
	ErrCodeInvalidWorkspaceName,
	nil,
)

var ErrEmptyWorkspaceSlug = commonerror.NewInvalid(
	"Workspace slug cannot be empty",
	ErrCodeInvalidWorkspaceSlug,
	nil,
)
