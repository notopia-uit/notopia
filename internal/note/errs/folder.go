package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeFolderAlreadyExisted Code = "folderAlreadyExisted"
	CodeFolderNotFound       Code = "folderNotFound"
	CodeEmptyFolderName      Code = "folderEmptyName"
	CodeFolderAlreadyTrashed Code = "folderAlreadyTrashed"
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
