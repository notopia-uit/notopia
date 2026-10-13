package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspaceBilling struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type GetWorkspaceBillingHandler struct {
	read  BillingReadModel
	authz AuthorizationSvc
}

func NewGetWorkspaceBillingHandler(read BillingReadModel, authz AuthorizationSvc) *GetWorkspaceBillingHandler {
	return &GetWorkspaceBillingHandler{read: read, authz: authz}
}

var ProvideGetWorkspaceBillingHandler = NewGetWorkspaceBillingHandler

type GetWorkspaceBillingQuery commonhandler.Query[GetWorkspaceBilling, WorkspaceBillingView]

var _ GetWorkspaceBillingQuery = (*GetWorkspaceBillingHandler)(nil)

func (h *GetWorkspaceBillingHandler) Handle(ctx context.Context, q *GetWorkspaceBilling) (WorkspaceBillingView, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return WorkspaceBillingView{}, err
	}
	var v WorkspaceBillingView
	account, err := h.read.Account(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	if account != nil {
		payer := newPayer(account)
		v.Payer = &payer
	}
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	if sub != nil {
		policy, err := h.read.Policy(ctx)
		if err != nil {
			return v, err
		}
		view := newSubscription(sub, policy)
		v.Subscription = &view
	}
	return v, nil
}
