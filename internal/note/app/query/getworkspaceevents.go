package query

import "context"

type GetWorkspaceEvents struct {
	Slug string
}

type GetWorkspaceEventsReadModel interface {
	GetWorkspaceEvents(ctx context.Context, q *GetWorkspaceEvents) (chan interface{}, error)
}

type GetWorkspaceEventsHandler struct {
	readModel GetWorkspaceEventsReadModel
}

func NewGetWorkspaceEventsHandler(readModel GetWorkspaceEventsReadModel) *GetWorkspaceEventsHandler {
	return &GetWorkspaceEventsHandler{readModel: readModel}
}

func (h *GetWorkspaceEventsHandler) Handle(ctx context.Context, query *GetWorkspaceEvents) (chan interface{}, error) {
	return h.readModel.GetWorkspaceEvents(ctx, query)
}
