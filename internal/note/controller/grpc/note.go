package grpc

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (s *ServiceServer) GetNoteName(ctx context.Context, req *pb.GetNoteNameRequest) (*pb.GetNoteNameResponse, error) {
	return nil, errs.NewUnimplemented()
}

func (s *ServiceServer) CheckNoteExistence(
	ctx context.Context,
	req *pb.CheckNoteExistenceRequest,
) (*pb.CheckNoteExistenceResponse, error) {
	return nil, errs.NewUnimplemented()
}

func (s *ServiceServer) GetWorkspaceIdByNoteId(
	ctx context.Context,
	req *pb.GetWorkspaceIdByNoteIdRequest,
) (*pb.GetWorkspaceIdByNoteIdResponse, error) {
	return nil, errs.NewUnimplemented()
}
