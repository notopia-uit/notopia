package grpc

import (
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toNoteDTO(note *app.Note) *pb.Note {
	var icon *string
	if note.Icon != "" {
		icon = &note.Icon
	}
	return &pb.Note{
		Id:        note.ID.String(),
		Name:      note.Name,
		Icon:      icon,
		FolderId:  note.FolderID.String(),
		Tags:      note.Tags,
		UpdatedAt: timestamppb.New(note.UpdatedAt),
		Trashed:   toTrashedDTO(note.Trashed),
	}
}

func toWorkspaceDTO(workspace *app.Workspace) *pb.Workspace {
	return &pb.Workspace{
		Id:   workspace.ID.String(),
		Name: workspace.Name,
		Slug: workspace.Slug,
	}
}

func toTrashedDTO(trash app.Trashed) *pb.Trashed {
	if trash.By == app.TrashedByUnspecified && trash.At.IsZero() {
		return nil
	}
	return &pb.Trashed{
		By: toTrashedByDTO(trash.By),
		At: timestamppb.New(trash.At),
	}
}

func toTrashedByDTO(trashedBy app.TrashedBy) pb.TrashedBy {
	switch trashedBy {
	case app.TrashedByPurpose:
		return pb.TrashedBy_TRASHED_BY_PURPOSE
	case app.TrashedByParent:
		return pb.TrashedBy_TRASHED_BY_PARENT
	case app.TrashedByUnspecified:
		return pb.TrashedBy_TRASHED_BY_UNSPECIFIED
	}
	return pb.TrashedBy_TRASHED_BY_UNSPECIFIED
}
