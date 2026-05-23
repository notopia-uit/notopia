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
	Limit                      uint
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

type SearchUsersQuery commonhandler.Query[SearchUsers, []User]

var _ SearchUsersQuery = (*SearchUsersHandler)(nil)

func (h *SearchUsersHandler) Handle(ctx context.Context, query *SearchUsers) ([]User, error) {
	users, err := h.identitySvc.SearchUsers(ctx, &IdentitySvcSearchUsersParams{
		Keyword:      query.Keyword,
		ActiveStatus: query.ActiveStatus,
		Limit:        query.Limit,
	})
	if err != nil {
		return nil, err
	}

	if query.ExcludeMemberInWorkspaceId == uuid.Nil {
		return users, nil
	}

	authorizationMembers, err := h.authorizationSvc.GetWorkspaceMembers(ctx, query.UserID, query.ExcludeMemberInWorkspaceId)
	if err != nil {
		return nil, err
	}

	excludedIDs := make(map[string]struct{}, len(authorizationMembers))
	for _, member := range authorizationMembers {
		excludedIDs[member.ID] = struct{}{}
	}

	filtered := make([]User, 0, len(users))
	for i := range users {
		if _, excluded := excludedIDs[users[i].ID]; !excluded {
			filtered = append(filtered, users[i])
		}
	}

	return filtered, nil
}
