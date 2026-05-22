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

	var keyword string
	if request.Params.Keyword != nil {
		keyword = *request.Params.Keyword
	}

	query := &app.SearchUsers{
		UserID:                     user.ID,
		Keyword:                    keyword,
		ActiveStatus:               activeStatus,
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

	dtos := make([]note.User, len(result.Data))
	for i := range result.Data {
		dtos[i] = toUserDTO(&result.Data[i])
	}
	paginationDTO := toPaginationDTO(&result.Pagination)

	return note.SearchUsers200JSONResponse{
		Data:       dtos,
		Pagination: paginationDTO,
	}, nil
}
