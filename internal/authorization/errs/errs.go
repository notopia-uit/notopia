package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeCasbinInternalError            Code = "casbinInternalError"
	CodeCasbinEnforcerError            Code = "casbinEnforcerFailed"
	CodeCasbinPolicySignatureInvalid   Code = "casbinPolicySignatureInvalid"
	CodeErrInvalidUserFormat           Code = "invalidUserFormat"
	CodeMemberHasNoPermission          Code = "memberHasNoPermission"
	CodeGetWorkspaceMembersGetFailed   Code = "getWorkspaceMembersGetFailed"
	CodeCreateWorkspaceExists          Code = "createWorkspaceExists"
	CodeInvalidWorkspaceRoleFormat     Code = "invalidWorkspaceRoleFormat"
	CodePublishIntegrationEventsFailed Code = "publishIntegrationEventsFailed"
	CodeUserIsOnlyOwner                Code = "userIsOnlyOwner"
)

type CasbinInternalError struct {
	Err
}

func NewCasbinInternalError(err error) *CasbinInternalError {
	return &CasbinInternalError{
		Err: Err{
			message: "casbin internal error",
			code:    CodeCasbinInternalError,
			err:     err,
		},
	}
}

type CasbinEnforcerError struct {
	Err
}

func NewCasbinEnforcerError(err error) *CasbinEnforcerError {
	return &CasbinEnforcerError{
		Err: Err{
			message: "casbin enforcer error",
			code:    CodeCasbinEnforcerError,
			err:     err,
		},
	}
}

type CasbinPolicySignatureInvalid struct {
	Err
}

func NewCasbinPolicySignatureInvalid(message ...string) *CasbinPolicySignatureInvalid {
	return &CasbinPolicySignatureInvalid{
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

type InvalidUserFormat struct {
	Err
	userID string
}

func NewInvalidUserFormat(userID string, err error) *InvalidUserFormat {
	return &InvalidUserFormat{
		userID: userID,
		Err: Err{
			message: "invalid user format: " + userID,
			code:    CodeErrInvalidUserFormat,
			err:     err,
		},
	}
}

type MemberHasNoPermission struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
	Permission  string
}

func NewMemberHasNoPermission(userID string, workspaceID uuid.UUID, permission string) *MemberHasNoPermission {
	return &MemberHasNoPermission{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Permission:  permission,
		Err: Err{
			message: fmt.Sprintf("user %q does not have %q permission for workspace %q", userID, permission, workspaceID.String()),
			code:    CodeMemberHasNoPermission,
		},
	}
}

type GetWorkspaceMembersGetFailed struct {
	Err
	WorkspaceID uuid.UUID
}

func NewGetWorkspaceMembersGetFailed(workspaceID uuid.UUID, err error) *GetWorkspaceMembersGetFailed {
	return &GetWorkspaceMembersGetFailed{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("failed to get workspace members for workspace %q", workspaceID.String()),
			code:    CodeGetWorkspaceMembersGetFailed,
			err:     err,
		},
	}
}

type CreateWorkspaceExists struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewCreateWorkspaceExists(userID string, workspaceID uuid.UUID) *CreateWorkspaceExists {
	return &CreateWorkspaceExists{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("workspace %q already exists with owner id %q", workspaceID.String(), userID),
			code:    CodeCreateWorkspaceExists,
			err:     nil,
		},
	}
}

type InvalidWorkspaceRoleFormat struct {
	Err
	roleStr string
}

func NewInvalidWorkspaceRoleFormat(roleStr string) *InvalidWorkspaceRoleFormat {
	return &InvalidWorkspaceRoleFormat{
		roleStr: roleStr,
		Err: Err{
			message: fmt.Sprintf("invalid workspace role format: %q", roleStr),
			code:    CodeInvalidWorkspaceRoleFormat,
		},
	}
}

type PublishIntegrationEventsFailed struct {
	Err
	WorkspaceID uuid.UUID
}

func NewPublishIntegrationEventsFailed(workspaceID uuid.UUID, err error) *PublishIntegrationEventsFailed {
	return &PublishIntegrationEventsFailed{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("failed to publish integration events for workspace %q", workspaceID.String()),
			code:    CodePublishIntegrationEventsFailed,
			err:     err,
		},
	}
}

type UserIsOnlyOwner struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewUserIsOnlyOwner(userID string, workspaceID uuid.UUID) *UserIsOnlyOwner {
	return &UserIsOnlyOwner{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("user %q is the only owner of workspace %q and cannot leave", userID, workspaceID.String()),
			code:    CodeUserIsOnlyOwner,
		},
	}
}
