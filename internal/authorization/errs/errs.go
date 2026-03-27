package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeCasbinInternalError          Code = "casbinInternalError"
	CodeCasbinEnforcerError          Code = "casbinEnforcerFailed"
	CodeCasbinPolicySignatureInvalid Code = "casbinPolicySignatureInvalid"
	CodeErrInvalidUserFormat         Code = "invalidUserFormat"
	CodeMemberHasNoPermission        Code = "memberHasNoPermission"
	CodeGetWorkspaceMembersGetFailed Code = "getWorkspaceMembersGetFailed"
	CodeCreateWorkspaceExists        Code = "createWorkspaceExists"
)

type casbinInternalError struct {
	Err
}

func NewCasbinInternalError(err error) *casbinInternalError {
	return &casbinInternalError{
		Err: Err{
			message: "casbin internal error",
			code:    CodeCasbinInternalError,
			err:     err,
		},
	}
}

type casbinEnforcerError struct {
	Err
}

func NewCasbinEnforcerError(err error) *casbinEnforcerError {
	return &casbinEnforcerError{
		Err: Err{
			message: "casbin enforcer error",
			code:    CodeCasbinEnforcerError,
			err:     err,
		},
	}
}

type casbinPolicySignatureInvalid struct {
	Err
}

func NewCasbinPolicySignatureInvalid(message ...string) *casbinPolicySignatureInvalid {
	return &casbinPolicySignatureInvalid{
		Err: Err{
			message: "casbin policy signature is invalid" + func() string {
				if len(message) > 0 {
					return ": " + message[0]
				}
				return ""
			}(),
			code: CodeCasbinPolicySignatureInvalid,
		},
	}
}

type invalidUserFormat struct {
	Err
	userID string
}

func NewInvalidUserFormat(userID string, err error) *invalidUserFormat {
	return &invalidUserFormat{
		userID: userID,
		Err: Err{
			message: "invalid user format: " + userID,
			code:    CodeErrInvalidUserFormat,
			err:     err,
		},
	}
}

type memberHasNoPermission struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
	Permission  string
}

func NewMemberHasNoPermission(userID string, workspaceID uuid.UUID, permission string) *memberHasNoPermission {
	return &memberHasNoPermission{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Permission:  permission,
		Err: Err{
			message: fmt.Sprintf("user %q does not have %q permission for workspace %q", userID, permission, workspaceID.String()),
			code:    CodeMemberHasNoPermission,
		},
	}
}

type getWorkspaceMembersGetFailed struct {
	Err
	WorkspaceID uuid.UUID
}

func NewGetWorkspaceMembersGetFailed(workspaceID uuid.UUID, err error) *getWorkspaceMembersGetFailed {
	return &getWorkspaceMembersGetFailed{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("failed to get workspace members for workspace %q", workspaceID.String()),
			code:    CodeGetWorkspaceMembersGetFailed,
			err:     err,
		},
	}
}

type createWorkspaceExists struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewCreateWorkspaceExists(userID string, workspaceID uuid.UUID) *createWorkspaceExists {
	return &createWorkspaceExists{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("workspace %q already exists with owner id %q", workspaceID.String(), userID),
			code:    CodeCreateWorkspaceExists,
			err:     nil,
		},
	}
}
