package http

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func currentUser(ctx context.Context) (*commonhttp.User, error) {
	u, ok := commonhttp.UserFromContext(ctx)
	if !ok || u.ID == "" {
		return nil, errs.New(errs.CodeUnauthorized, "unauthorized", nil)
	}
	return u, nil
}

func (h *StrictHandler) requireAdmin(ctx context.Context) (*commonhttp.User, error) {
	u, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !u.HasGroup(h.Identity.AdminGroup) {
		return nil, errs.NewForbidden("admin only")
	}
	return u, nil
}
