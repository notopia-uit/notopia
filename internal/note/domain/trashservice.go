package domain

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

func (s *TrashService) TrashFoldersWithChildren(
	targetFolders []*Folder,
	childFolders []*Folder,
	childNotes []*Note,
	userID string,
) error {
	// Trash target folders first with purpose
	for i := range targetFolders {
		if err := targetFolders[i].Trash(TrashedByPurpose, userID); err != nil {
			return err
		}
	}

	// Trash all child folders that aren't already trashed
	for i := range childFolders {
		if !childFolders[i].IsTrashed() {
			if err := childFolders[i].Trash(TrashedByParent, userID); err != nil {
				return err
			}
		}
	}

	// Trash all child notes that aren't already trashed
	for i := range childNotes {
		if !childNotes[i].IsTrashed() {
			if err := childNotes[i].Trash(TrashedByParent, userID); err != nil {
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

func (s *TrashService) RestoreFoldersWithChildren(
	targetFolders []*Folder,
	childFolders []*Folder,
	childNotes []*Note,
	userID string,
) error {
	// Restore target folders first
	for i := range targetFolders {
		targetFolders[i].Restore(userID)
	}

	// Restore child folders only if they were trashed by parent
	for i := range childFolders {
		if childFolders[i].IsTrashed() && childFolders[i].TrashedBy() == TrashedByParent {
			childFolders[i].Restore(userID)
		}
	}

	// Restore child notes only if they were trashed by parent
	for i := range childNotes {
		if childNotes[i].IsTrashed() && childNotes[i].TrashedBy() == TrashedByParent {
			childNotes[i].Restore(userID)
		}
	}

	return nil
}
