package domain

import (
	"fmt"

	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var ErrCodeWorkspaceNotFound = "Workspace_1"

func NewErrWorkspaceNotFound(id uuid.UUID) *commonerror.Err {
	return commonerror.NewNotFound(
		fmt.Sprintf("Workspace with id %q not found", id.String()),
		ErrCodeWorkspaceNotFound,
		nil,
	)
}
