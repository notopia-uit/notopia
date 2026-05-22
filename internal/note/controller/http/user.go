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
}
