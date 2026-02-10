package grpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *Handler) GetNote(ctx context.Context, req *connect.Request[pb.GetNoteRequest]) (*connect.Response[pb.GetNoteResponse], error) {
	panic("implement me")
}
