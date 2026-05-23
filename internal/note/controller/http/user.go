package http

import (
	"context"

	"github.com/google/uuid"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) SearchUsers(ctx context.Context, request note.SearchUsersRequestObject) (note.SearchUsersResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	var activeStatus app.IdentitySvcActiveStatus
	if request.Params.IsActive != nil {
		if *request.Params.IsActive {
			activeStatus = app.IdentitySvcActiveStatusActive
		} else {
			activeStatus = app.IdentitySvcActiveStatusInactive
		}
	}

	var excludeWorkspaceID uuid.UUID
	if request.Params.ExcludeMemberInWorkspaceId != nil {
		excludeWorkspaceID = *request.Params.ExcludeMemberInWorkspaceId
	}

	limit := uint(20)
	if request.Params.Limit != nil {
		limit = uint(*request.Params.Limit)
	}

	query := &app.SearchUsers{
		UserID:                     user.ID,
		Keyword:                    request.Params.Keyword,
		ActiveStatus:               activeStatus,
		Limit:                      limit,
		ExcludeMemberInWorkspaceId: excludeWorkspaceID,
		PaginationParams: app.PaginationParams{
			Page:  1,
			Limit: 20,
		},
	}
	result, err := h.App.Queries.SearchUsers.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dtos := toUserDTOs(result)

	return note.SearchUsers200JSONResponse(dtos), nil
}
