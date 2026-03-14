package query

type ShowTrash struct {
	Slug       string
	Pagination *PaginationParams
}

type ShowTrashReadModel interface {
	ShowTrash(*ShowTrash) (Trash, error)
}

type ShowTrashHandler struct {
	readModel ShowTrashReadModel
}

func NewShowTrashHandler(readModel ShowTrashReadModel) *ShowTrashHandler {
	return &ShowTrashHandler{readModel: readModel}
}

func (h *ShowTrashHandler) Handle(query *ShowTrash) (Trash, error) {
	return h.readModel.ShowTrash(query)
}
