package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeWorkspaceNotFound                Code = "workspaceNotFound"
	CodeWorkspaceBySlugNotFound          Code = "workspaceBySlugNotFound"
	CodeWorkspaceRootFolderNotFound      Code = "workspaceRootFolderNotFound"
	CodeInvalidWorkspaceName             Code = "invalidWorkspaceName"
	CodeInvalidWorkspaceSlug             Code = "invalidWorkspaceSlug"
	CodeWorkspaceSlugAlreadyExists       Code = "workspaceSlugAlreadyExists"
	CodeWorkspaceMembersCannotBeEmpty    Code = "workspaceMembersCannotBeEmpty"
	CodeWorkspaceMustHaveAtLeastOneOwner Code = "workspaceMustHaveAtLeastOneOwner"
)

type WorkspaceNotFound struct {
	Err
	WorkspaceID uuid.UUID
}

func NewWorkspaceNotFound(id uuid.UUID, err error) *WorkspaceNotFound {
	return &WorkspaceNotFound{
		WorkspaceID: id,
		Err: Err{
			message: fmt.Sprintf("workspace with id %q not found", id.String()),
			code:    CodeWorkspaceNotFound,
			err:     err,
		},
	}
}

type WorkspaceBySlugNotFound struct {
	Err
	Slug string
}

func NewWorkspaceBySlugNotFound(slug string, err error) *WorkspaceBySlugNotFound {
	return &WorkspaceBySlugNotFound{
		Slug: slug,
		Err: Err{
			message: fmt.Sprintf("workspace with slug %q not found", slug),
			code:    CodeWorkspaceBySlugNotFound,
			err:     err,
		},
	}
}

type WorkspaceRootFolderNotFound struct {
	Err
	WorkspaceID uuid.UUID
}

func NewWorkspaceRootFolderNotFound(id uuid.UUID, err error) *WorkspaceRootFolderNotFound {
	return &WorkspaceRootFolderNotFound{
		WorkspaceID: id,
		Err: Err{
			message: fmt.Sprintf("root folder for workspace with id %q not found", id.String()),
			code:    CodeWorkspaceRootFolderNotFound,
			err:     err,
		},
	}
}

var InvalidWorkspaceName = &Err{
	message: "workspace name cannot be empty",
	code:    CodeInvalidWorkspaceName,
}

var InvalidWorkspaceSlug = &Err{
	message: "workspace slug cannot be empty",
	code:    CodeInvalidWorkspaceSlug,
}

type WorkspaceSlugAlreadyExists struct {
	Err
	Slug string
}

func NewWorkspaceSlugAlreadyExists(slug string, err error) *WorkspaceSlugAlreadyExists {
	return &WorkspaceSlugAlreadyExists{
		Slug: slug,
		Err: Err{
			message: fmt.Sprintf("workspace slug %q already exists", slug),
			code:    CodeWorkspaceSlugAlreadyExists,
			err:     err,
		},
	}
}

type WorkspaceMembersCannotBeEmpty struct {
	Err
	ID uuid.UUID
}

func NewWorkspaceMembersCannotBeEmpty(id uuid.UUID) *WorkspaceMembersCannotBeEmpty {
	return &WorkspaceMembersCannotBeEmpty{
		ID: id,
		Err: Err{
			message: fmt.Sprintf("workspace with id %q cannot have empty members", id.String()),
			code:    CodeWorkspaceMembersCannotBeEmpty,
		},
	}
}

type WorkspaceMustHaveAtLeastOneOwner struct {
	Err
	ID uuid.UUID
}

func NewWorkspaceMustHaveAtLeastOneOwner(id uuid.UUID) *WorkspaceMustHaveAtLeastOneOwner {
	return &WorkspaceMustHaveAtLeastOneOwner{
		ID: id,
		Err: Err{
			message: fmt.Sprintf("workspace with id %q must have at least one owner", id.String()),
			code:    CodeWorkspaceMustHaveAtLeastOneOwner,
		},
	}
}
