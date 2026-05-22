package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SearchUsers struct {
	UserID                     string
	Keyword                    string
	ActiveStatus               IdentitySvcActiveStatus
	ExcludeMemberInWorkspaceId uuid.UUID // Emptyable
	PaginationParams
}

type SearchUsersHandler struct {
	identitySvc      IdentitySvc
	authorizationSvc AuthorizationSvc
}

func NewSearchUsersHandler(
	identitySvc IdentitySvc,
	authorizationSvc AuthorizationSvc,
) *SearchUsersHandler {
	return &SearchUsersHandler{
		identitySvc:      identitySvc,
		authorizationSvc: authorizationSvc,
	}
}

var ProvideSearchUsersHandler = NewSearchUsersHandler

type SearchUsersQuery commonhandler.Query[SearchUsers, Paginated[User]]

var _ SearchUsersQuery = (*SearchUsersHandler)(nil)

func (h *SearchUsersHandler) Handle(ctx context.Context, query *SearchUsers) (Paginated[User], error) {
	result, err := h.identitySvc.SearchUsers(ctx, &IdentitySvcSearchUsersParams{
		Keyword:      query.Keyword,
		ActiveStatus: query.ActiveStatus,
		PaginationParams: PaginationParams{
			Page:  query.Page,
			Limit: query.Limit,
		},
	})
	if err != nil {
		return Paginated[User]{}, err
	}

	if query.ExcludeMemberInWorkspaceId == uuid.Nil {
		return result, nil
	}

	authorizationMembers, err := h.authorizationSvc.GetWorkspaceMembers(ctx, query.UserID, query.ExcludeMemberInWorkspaceId)
	if err != nil {
		return Paginated[User]{}, err
	}

	excludedIDs := make(map[string]struct{}, len(authorizationMembers))
	for _, member := range authorizationMembers {
		excludedIDs[member.ID] = struct{}{}
	}

	filtered := make([]User, 0, len(result.Data))
	for i := range result.Data {
		if _, excluded := excludedIDs[result.Data[i].ID]; !excluded {
			filtered = append(filtered, result.Data[i])
		}
	}

	total := uint(len(filtered))
	limit := uint(query.Limit)
	page := uint(query.Page)
	var totalPages uint
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	currentTotal := limit
	if page*limit > total {
		currentTotal = total - (page-1)*limit
	}

	return Paginated[User]{
		Data: filtered,
		Pagination: Pagination{
			Page:         page,
			CurrentTotal: currentTotal,
			Total:        total,
			TotalPages:   totalPages,
			HasNext:      page < totalPages,
			HasPrev:      page > 1,
		},
	}, nil
}
