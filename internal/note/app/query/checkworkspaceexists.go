package query

type CheckWorkspaceExists struct {
	Slug string
}

type CheckWorkspaceExistsReadModel interface {
	CheckWorkspaceExists(*CheckWorkspaceExists) (CheckWorkspaceExistsResult, error)
}

type CheckWorkspaceExistsHandler struct {
	readModel CheckWorkspaceExistsReadModel
}

func NewCheckWorkspaceExistsHandler(readModel CheckWorkspaceExistsReadModel) *CheckWorkspaceExistsHandler {
	return &CheckWorkspaceExistsHandler{readModel: readModel}
}

func (h *CheckWorkspaceExistsHandler) Handle(query *CheckWorkspaceExists) (CheckWorkspaceExistsResult, error) {
	return h.readModel.CheckWorkspaceExists(query)
}
