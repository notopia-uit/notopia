package grpc

import (
	"context"

	"connectrpc.com/connect"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *Handler) CheckNoteExistence(
	ctx context.Context,
	req *connect.Request[pb.CheckNoteExistenceRequest],
) (*connect.Response[pb.CheckNoteExistenceResponse], error) {
	// TODO: Implement gRPC endpoint for checking note existence
	// Consider delegating to a handler similar to HTTP controllers or direct repo access
	// Response model: &pb.CheckNoteExistenceResponse{Exists: bool}
	return nil, commonerror.NewUnimplemented()
}
