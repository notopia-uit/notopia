package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type permanentDeleteTargets struct {
	notes          []*domain.Note
	folders        []*domain.Folder
	requestedCount int
}

func (t permanentDeleteTargets) requestedNotes() []*domain.Note {
	return t.notes[:t.requestedCount]
}

func (t permanentDeleteTargets) folderChildNotes() []*domain.Note {
	return t.notes[t.requestedCount:]
}

func (t permanentDeleteTargets) delete(trashService *domain.TrashService, userID string) {
	trashService.PermanentlyDeleteNotes(t.requestedNotes(), userID)
	trashService.PermanentlyDeleteFoldersWithChildren(t.folders, t.folderChildNotes(), userID)
}

func loadPermanentDeleteTargets(
	ctx context.Context,
	r domain.RepoRegistry,
	noteIDs, folderIDs []uuid.UUID,
) (permanentDeleteTargets, error) {
	noteRepo := r.Note()
	folderRepo := r.Folder()

	var targets permanentDeleteTargets
	seen := map[uuid.UUID]struct{}{}

	if len(noteIDs) > 0 {
		requested, err := noteRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.NoteRepoGetManyParams{
				IDs:       noteIDs,
				ForUpdate: true,
			})
		if err != nil {
			return permanentDeleteTargets{}, err
		}
		for _, n := range requested {
			if _, ok := seen[n.ID()]; ok {
				continue
			}
			seen[n.ID()] = struct{}{}
			targets.notes = append(targets.notes, n)
		}
	}
	targets.requestedCount = len(targets.notes)

	if len(folderIDs) > 0 {
		folders, err := folderRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.FolderRepoGetManyParams{
				IDs:       folderIDs,
				ForUpdate: true,
			})
		if err != nil {
			return permanentDeleteTargets{}, err
		}
		targets.folders = folders

		for _, folder := range folders {
			children, err := noteRepo.GetRecursiveChildrenFromFolder(ctx, folder.ID(), true)
			if err != nil {
				return permanentDeleteTargets{}, err
			}
			for _, n := range children {
				if _, ok := seen[n.ID()]; ok {
					continue
				}
				seen[n.ID()] = struct{}{}
				targets.notes = append(targets.notes, n)
			}
		}
	}

	return targets, nil
}

func savePermanentDeleteTargets(
	ctx context.Context,
	r domain.RepoRegistry,
	t permanentDeleteTargets,
) error {
	if err := r.Note().SaveMany(ctx, t.notes); err != nil {
		return err
	}
	return r.Folder().SaveMany(ctx, t.folders)
}
