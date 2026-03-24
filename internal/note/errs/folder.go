package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeFolderNotFound       Code = "folder_1"
	CodeEmptyFolderName      Code = "folder_2"
	CodeFolderAlreadyTrashed Code = "folder_3"
)

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
