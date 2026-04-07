package domain

import (
	"github.com/google/uuid"
)

type TrashService struct{}

func NewTrashService() *TrashService {
	return &TrashService{}
}

var ProvideTrashService = NewTrashService

func (s *TrashService) TrashNotes(notes []*Note, userID string) error {
	for i := range notes {
		if err := notes[i].Trash(TrashedByPurpose, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) TrashFolders(
	workspaceNotes *[]*Note,
	workspaceFolders *[]*Folder,
	targetFolders []*Folder,
	userID string,
) error {
	for i := range targetFolders {
		if err := targetFolders[i].Trash(TrashedByPurpose, userID); err != nil {
			return err
		}

		if err := s.cascadeTrashChildren(workspaceNotes, workspaceFolders, targetFolders[i].ID(), userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) cascadeTrashChildren(
	workspaceNotes *[]*Note,
	workspaceFolders *[]*Folder,
	folderID uuid.UUID,
	userID string,
) error {
	for i := range *workspaceFolders {
		folder := (*workspaceFolders)[i]
		if folder.ParentID() != uuid.Nil && folder.ParentID() == folderID && !folder.IsTrashed() {
			if err := folder.Trash(TrashedByParent, userID); err != nil {
				return err
			}

			if err := s.cascadeTrashChildren(workspaceNotes, workspaceFolders, folder.ID(), userID); err != nil {
				return err
			}
		}
	}

	for i := range *workspaceNotes {
		note := (*workspaceNotes)[i]
		if note.FolderID() == folderID && !note.IsTrashed() {
			if err := note.Trash(TrashedByParent, userID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TrashService) RestoreNotes(notes []*Note, userID string) error {
	for i := range notes {
		notes[i].Restore(userID)
	}
	return nil
}

func (s *TrashService) RestoreFolders(
	trashedNotes *[]*Note,
	trashedFolders *[]*Folder,
	targetFolders []*Folder,
	userID string,
) error {
	for i := range targetFolders {
		targetFolders[i].Restore(userID)

		if err := s.cascadeRestoreChildrenByParent(trashedNotes, trashedFolders, targetFolders[i].ID(), userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrashService) cascadeRestoreChildrenByParent(
	trashedNotes *[]*Note,
	trashedFolders *[]*Folder,
	folderID uuid.UUID,
	userID string,
) error {
	for i := range *trashedFolders {
		folder := (*trashedFolders)[i]
		if folder.ParentID() != uuid.Nil && folder.ParentID() == folderID && folder.IsTrashed() {
			trashedBy := folder.TrashedBy()
			if trashedBy != TrashedByUnspecified && trashedBy == TrashedByParent {
				folder.Restore(userID)

				if err := s.cascadeRestoreChildrenByParent(trashedNotes, trashedFolders, folder.ID(), userID); err != nil {
					return err
				}
			}
		}
	}

	for i := range *trashedNotes {
		note := (*trashedNotes)[i]
		if note.FolderID() == folderID && note.IsTrashed() {
			trashedBy := note.TrashedBy()
			if trashedBy != TrashedByUnspecified && trashedBy == TrashedByParent {
				note.Restore(userID)
			}
		}
	}

	return nil
}
