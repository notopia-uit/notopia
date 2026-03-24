package domain

import (
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type TrashService struct{}

func NewTrashService() *TrashService {
	return &TrashService{}
}

var ProvideTrashService = NewTrashService

func (s *TrashService) TrashNotes(notes []*Note) errs.Error {
	for i := range notes {
		if err := notes[i].Trash(TrashedByPurpose); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) TrashFolders(
	workspaceNotes *[]*Note,
	workspaceFolders *[]*Folder,
	targetFolders []*Folder,
) errs.Error {
	for i := range targetFolders {
		if err := targetFolders[i].Trash(TrashedByPurpose); err != nil {
			return err
		}

		if err := s.cascadeTrashChildren(workspaceNotes, workspaceFolders, targetFolders[i].ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) cascadeTrashChildren(
	workspaceNotes *[]*Note,
	workspaceFolders *[]*Folder,
	folderID uuid.UUID,
) errs.Error {
	for i := range *workspaceFolders {
		folder := (*workspaceFolders)[i]
		if folder.ParentID() != nil && *folder.ParentID() == folderID && !folder.IsTrashed() {
			if err := folder.Trash(TrashedByParent); err != nil {
				return err
			}

			if err := s.cascadeTrashChildren(workspaceNotes, workspaceFolders, folder.ID()); err != nil {
				return err
			}
		}
	}

	for i := range *workspaceNotes {
		note := (*workspaceNotes)[i]
		if note.FolderID() == folderID && !note.IsTrashed() {
			if err := note.Trash(TrashedByParent); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TrashService) RestoreNotes(notes []*Note) errs.Error {
	for i := range notes {
		notes[i].Restore()
	}
	return nil
}

func (s *TrashService) RestoreFolders(
	trashedNotes *[]*Note,
	trashedFolders *[]*Folder,
	targetFolders []*Folder,
) errs.Error {
	for i := range targetFolders {
		targetFolders[i].Restore()

		if err := s.cascadeRestoreChildrenByParent(trashedNotes, trashedFolders, targetFolders[i].ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) cascadeRestoreChildrenByParent(
	trashedNotes *[]*Note,
	trashedFolders *[]*Folder,
	folderID uuid.UUID,
) errs.Error {
	for i := range *trashedFolders {
		folder := (*trashedFolders)[i]
		if folder.ParentID() != nil && *folder.ParentID() == folderID && folder.IsTrashed() {
			trashedBy := folder.TrashedBy()
			if trashedBy != nil && *trashedBy == TrashedByParent {
				folder.Restore()

				if err := s.cascadeRestoreChildrenByParent(trashedNotes, trashedFolders, folder.ID()); err != nil {
					return err
				}
			}
		}
	}

	for i := range *trashedNotes {
		note := (*trashedNotes)[i]
		if note.FolderID() == folderID && note.IsTrashed() {
			trashedBy := note.TrashedBy()
			if trashedBy != nil && *trashedBy == TrashedByParent {
				note.Restore()
			}
		}
	}

	return nil
}
