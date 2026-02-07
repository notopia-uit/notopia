package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func (h *RPCHandler) GetNote(ctx context.Context, req *connect.Request[pb.GetNoteRequest]) (*connect.Response[pb.GetNoteResponse], error) {
	panic("implement me")
}
