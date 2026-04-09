package app

import (
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

func deduplicateNotes(notes []*domain.Note) []*domain.Note {
	if len(notes) == 0 {
		return notes
	}

	seen := make(map[uuid.UUID]*domain.Note)
	for _, note := range notes {
		seen[note.ID()] = note
	}

	result := make([]*domain.Note, 0, len(seen))
	for _, note := range seen {
		result = append(result, note)
	}
	return result
}

func deduplicateFolders(folders []*domain.Folder) []*domain.Folder {
	if len(folders) == 0 {
		return folders
	}

	seen := make(map[uuid.UUID]*domain.Folder)
	for _, folder := range folders {
		seen[folder.ID()] = folder
	}

	result := make([]*domain.Folder, 0, len(seen))
	for _, folder := range seen {
		result = append(result, folder)
	}
	return result
}
