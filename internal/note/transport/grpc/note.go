package grpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *Handler) CheckNoteExistence(
	ctx context.Context,
	req *connect.Request[pb.CheckNoteExistenceRequest],
) (*connect.Response[pb.CheckNoteExistenceResponse], error) {
	return nil, errors.New("not implemented")
}
