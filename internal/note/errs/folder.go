package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeFolderAlreadyExisted             Code = "folderAlreadyExisted"
	CodeFolderNotFound                   Code = "folderNotFound"
	CodeEmptyFolderName                  Code = "folderEmptyName"
	CodeCannotMoveFolderToItOwnSubfolder Code = "cannotMoveFolderToItOwnSubfolder"
	CodeFolderAlreadyTrashed             Code = "folderAlreadyTrashed"
	CodeFoldersNotInWorkspace            Code = "foldersNotInWorkspace"
	CodeDestinationFolderNotInWorkspace  Code = "destinationFolderNotInWorkspace"
)

type FolderAlreadyExisted struct {
	Err
	FolderID uuid.UUID
}

func NewFolderAlreadyExisted(id uuid.UUID) *FolderAlreadyExisted {
	return &FolderAlreadyExisted{
		FolderID: id,
		Err: Err{
			message: fmt.Sprintf("folder with id %q already existed", id.String()),
			code:    CodeFolderAlreadyExisted,
		},
	}
}

type FolderNotFound struct {
	Err
	FolderID uuid.UUID
}

func NewFolderNotFound(id uuid.UUID, err error) *FolderNotFound {
	return &FolderNotFound{
		FolderID: id,
		Err: Err{
			message: fmt.Sprintf("folder with id %q not found", id.String()),
			code:    CodeFolderNotFound,
			err:     err,
		},
	}
}

var EmptyFolderName = &Err{
	message: "folder name cannot be empty",
	code:    CodeEmptyFolderName,
}

type FolderAlreadyTrashed struct {
	Err
	FolderID uuid.UUID
}

func NewFolderAlreadyTrashed(id uuid.UUID) *FolderAlreadyTrashed {
	return &FolderAlreadyTrashed{
		FolderID: id,
		Err: Err{
			message: fmt.Sprintf("folder with id %q is already trashed", id.String()),
			code:    CodeFolderAlreadyTrashed,
			err:     nil,
		},
	}
}

type CannotMoveFolderToItOwnSubfolder struct {
	Err
	FolderID            uuid.UUID
	DestinationFolderID uuid.UUID
}

func NewCannotMoveFolderToItOwnSubfolder(folderID, destinationFolderID uuid.UUID) *CannotMoveFolderToItOwnSubfolder {
	return &CannotMoveFolderToItOwnSubfolder{
		FolderID:            folderID,
		DestinationFolderID: destinationFolderID,
		Err: Err{
			message: fmt.Sprintf("cannot move folder %s into its own subfolder %s", folderID, destinationFolderID),
			code:    CodeCannotMoveFolderToItOwnSubfolder,
		},
	}
}

type FoldersNotInWorkspace struct {
	Err
	WorkspaceID uuid.UUID
}

func NewFoldersNotInWorkspace(workspaceID uuid.UUID) *FoldersNotInWorkspace {
	return &FoldersNotInWorkspace{
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("one or more folders do not belong to workspace %s", workspaceID),
			code:    CodeFoldersNotInWorkspace,
		},
	}
}

type DestinationFolderNotInWorkspace struct {
	Err
	FolderID    uuid.UUID
	WorkspaceID uuid.UUID
}

func NewDestinationFolderNotInWorkspace(folderID, workspaceID uuid.UUID) *DestinationFolderNotInWorkspace {
	return &DestinationFolderNotInWorkspace{
		FolderID:    folderID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: fmt.Sprintf("destination folder %s does not belong to workspace %s", folderID, workspaceID),
			code:    CodeDestinationFolderNotInWorkspace,
		},
	}
}
