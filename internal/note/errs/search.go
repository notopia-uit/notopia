package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeGenerateWorkspaceSearchTokenFailed Code = "generateWorkspaceSearchTokenFailed"
)

type GenerateWorkspaceSearchTokenFailed struct {
	Err
	WorkspaceID uuid.UUID
}

func NewFailedToGenerateWorkspaceSearchToken(workspaceID uuid.UUID, err error) *GenerateWorkspaceSearchTokenFailed {
	return &GenerateWorkspaceSearchTokenFailed{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("failed to generate workspace search token for workspace %s: %v", workspaceID.String(), err),
			code:    CodeGenerateWorkspaceSearchTokenFailed,
			err:     err,
		},
	}
}
