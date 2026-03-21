package domain

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RepoRegistry interface {
	Workspace() WorkspaceRepo
	Folder() FolderRepo
	Note() NoteRepo
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repoRegistry RepoRegistry) errs.Error) errs.Error
}
