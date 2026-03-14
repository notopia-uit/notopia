package query

type GetWorkspaceGraph struct {
	Slug   string
	Orphan *bool
}

type GetWorkspaceGraphReadModel interface {
	GetWorkspaceGraph(*GetWorkspaceGraph) (Graph, error)
}

type GetWorkspaceGraphHandler struct {
	readModel GetWorkspaceGraphReadModel
}

func NewGetWorkspaceGraphHandler(readModel GetWorkspaceGraphReadModel) *GetWorkspaceGraphHandler {
	return &GetWorkspaceGraphHandler{readModel: readModel}
}

func (h *GetWorkspaceGraphHandler) Handle(query *GetWorkspaceGraph) (Graph, error) {
	return h.readModel.GetWorkspaceGraph(query)
}
