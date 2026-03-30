package grpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *Handler) CheckNoteExistence(
	ctx context.Context,
	req *connect.Request[pb.CheckNoteExistenceRequest],
) (*connect.Response[pb.CheckNoteExistenceResponse], error) {
	// TODO: Implement gRPC endpoint for checking note existence
	// Consider delegating to a handler similar to HTTP controllers or direct repo access
	// Response model: &pb.CheckNoteExistenceResponse{Exists: bool}
	return nil, errs.NewUnimplemented()
}

func (h *Handler) GetWorkspaceIdByNoteId(
	ctx context.Context,
	req *connect.Request[pb.GetWorkspaceIdByNoteIdRequest],
) (*connect.Response[pb.GetWorkspaceIdByNoteIdResponse], error) {
	// TODO: Implement gRPC endpoint for getting workspace ID by note ID
	// Consider delegating to a handler similar to HTTP controllers or direct repo access
	// Response model: &pb.GetWorkspaceIdByNoteIdResponse{WorkspaceId: string}
	return nil, errs.NewUnimplemented()
}

func (h *Handler) GetNoteName(ctx context.Context, req *connect.Request[pb.GetNoteNameRequest]) (*connect.Response[pb.GetNoteNameResponse], error) {
	// TODO: Implement gRPC endpoint for getting note name
	// Consider delegating to a handler similar to HTTP controllers or direct repo access
	// Response model: &pb.GetNoteNameResponse{NoteName: string}
	return nil, errs.NewUnimplemented()
}
