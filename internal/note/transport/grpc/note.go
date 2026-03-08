package grpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *Handler) GetLatestNoteContent(
	ctx context.Context,
	req *connect.Request[pb.GetLatestNoteContentRequest],
) (*connect.Response[pb.GetLatestNoteContentResponse], error) {
	return nil, errors.New("not implemented")
}
