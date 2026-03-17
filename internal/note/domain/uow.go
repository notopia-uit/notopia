package domain

import "context"

type RepoRegistry interface {
	Workspace() WorkspaceRepo
	Folder() FolderRepo
	Note() NoteRepo
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repoRegistry RepoRegistry) error) error
}
