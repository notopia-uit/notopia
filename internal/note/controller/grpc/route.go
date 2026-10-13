package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
)

// NOTE: This might be deleted
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
	note, err := s.app.Queries.GetNote.Handle(ctx, &app.GetNote{
		ID:             noteID,
		ExcludeTrashed: req.ExcludeTrashed,
		UserID:         req.UserId,
	})
	if err != nil {
		return nil, err
	}
	noteDTO := toNoteDTO(note)
	return &pb.GetNoteResponse{
		Note: noteDTO,
	}, nil
}

func (s *ServiceServer) GetWorkspaceByNote(
	ctx context.Context,
	req *pb.GetWorkspaceByNoteRequest,
) (*pb.GetWorkspaceByNoteResponse, error) {
	noteID, err := uuid.Parse(req.NoteId)
	if err != nil {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid note id: %v", err))
	}
	workspace, err := s.app.Queries.GetWorkspaceByNote.Handle(ctx, &app.GetWorkspaceByNote{
		NoteID: noteID,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	workspaceDTO := toWorkspaceDTO(&workspace)
	return &pb.GetWorkspaceByNoteResponse{
		Workspace: workspaceDTO,
	}, nil
}

func (s *ServiceServer) BatchGetNoteWorkspaces(
	ctx context.Context,
	req *pb.BatchGetNoteWorkspacesRequest,
) (*pb.BatchGetNoteWorkspacesResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.NoteIds))
	for _, raw := range req.NoteIds {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, errs.NewInvalid(fmt.Sprintf("invalid note id: %v", err))
		}
		ids = append(ids, id)
	}
	result, err := s.app.Queries.BatchGetNoteWorkspaces.Handle(ctx, &app.BatchGetNoteWorkspaces{NoteIDs: ids})
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(result))
	for noteID, workspaceID := range result {
		out[noteID.String()] = workspaceID.String()
	}
	return &pb.BatchGetNoteWorkspacesResponse{WorkspaceIds: out}, nil
}
