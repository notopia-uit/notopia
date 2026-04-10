package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ServiceServer) GetNoteName(ctx context.Context, req *pb.GetNoteNameRequest) (*pb.GetNoteNameResponse, error) {
	return nil, errs.Unimplemented
}

func (s *ServiceServer) GetNote(
	ctx context.Context,
	req *pb.GetNoteRequest,
) (*pb.GetNoteResponse, error) {
	noteID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid note id: %v", err))
	}
	note, err := s.app.Queries.GetNoteHandler.Handle(ctx, &app.GetNote{
		ID:             noteID,
		ExcludeTrashed: req.ExcludeTrashed,
		UserID:         req.UserId,
	})
	if err != nil {
		return nil, err
	}
	var icon *string
	if note.Icon != "" {
		icon = &note.Icon
	}
	return &pb.GetNoteResponse{
		Id:        note.ID.String(),
		Name:      note.Name,
		Icon:      icon,
		FolderId:  note.FolderID.String(),
		Tags:      note.Tags,
		UpdatedAt: timestamppb.New(note.UpdatedAt),
		Trashed:   toTrashed(note.Trashed),
	}, nil
}

func (s *ServiceServer) GetWorkspaceIdByNoteId(
	ctx context.Context,
	req *pb.GetWorkspaceIdByNoteIdRequest,
) (*pb.GetWorkspaceIdByNoteIdResponse, error) {
	return nil, errs.Unimplemented
}
