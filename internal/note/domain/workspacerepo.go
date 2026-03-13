package domain

import "github.com/google/uuid"

type WorkspaceRepo interface {
	GetBySlug(slug string) (*Workspace, error)
	GetIDBySlug(slug string) (uuid.UUID, error)
	CheckSlugExists(slug string) (bool, error)
	Save(workspace *Workspace) error
}
