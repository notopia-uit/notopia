package query

type GetWorkspaceEvents struct {
	Slug string
}

type GetWorkspaceEventsReadModel interface {
	GetWorkspaceEvents(*GetWorkspaceEvents) (chan interface{}, error)
}

type GetWorkspaceEventsHandler struct {
	readModel GetWorkspaceEventsReadModel
}

func NewGetWorkspaceEventsHandler(readModel GetWorkspaceEventsReadModel) *GetWorkspaceEventsHandler {
	return &GetWorkspaceEventsHandler{readModel: readModel}
}

func (h *GetWorkspaceEventsHandler) Handle(query *GetWorkspaceEvents) (chan interface{}, error) {
	return h.readModel.GetWorkspaceEvents(query)
}
