package query

type GetWorkspaceTree struct {
	Slug string
}

type GetWorkspaceTreeReadModel interface {
	GetWorkspaceTree(*GetWorkspaceTree) (WorkspaceTreeFolder, error)
}

type GetWorkspaceTreeHandler struct {
	readModel GetWorkspaceTreeReadModel
}

func NewGetWorkspaceTreeHandler(readModel GetWorkspaceTreeReadModel) *GetWorkspaceTreeHandler {
	return &GetWorkspaceTreeHandler{readModel: readModel}
}

func (h *GetWorkspaceTreeHandler) Handle(query *GetWorkspaceTree) (WorkspaceTreeFolder, error) {
	return h.readModel.GetWorkspaceTree(query)
}
