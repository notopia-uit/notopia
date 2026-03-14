package query

type GetWorkspaceMembers struct {
	Slug string
}

type GetWorkspaceMembersReadModel interface {
	GetWorkspaceMembers(*GetWorkspaceMembers) ([]WorkspaceMember, error)
}

type GetWorkspaceMembersHandler struct {
	readModel GetWorkspaceMembersReadModel
}

func NewGetWorkspaceMembersHandler(readModel GetWorkspaceMembersReadModel) *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{readModel: readModel}
}

func (h *GetWorkspaceMembersHandler) Handle(query *GetWorkspaceMembers) ([]WorkspaceMember, error) {
	return h.readModel.GetWorkspaceMembers(query)
}
