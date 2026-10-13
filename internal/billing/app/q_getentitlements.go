package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetEntitlements struct {
	WorkspaceID uuid.UUID
	UserID      string
	Internal    bool // gRPC from another service: skip the permission check
}

type GetEntitlementsHandler struct {
	read  BillingReadModel
	authz AuthorizationSvc
}

func NewGetEntitlementsHandler(read BillingReadModel, authz AuthorizationSvc) *GetEntitlementsHandler {
	return &GetEntitlementsHandler{read: read, authz: authz}
}

var ProvideGetEntitlementsHandler = NewGetEntitlementsHandler

type GetEntitlementsQuery commonhandler.Query[GetEntitlements, Entitlements]

var _ GetEntitlementsQuery = (*GetEntitlementsHandler)(nil)

func (h *GetEntitlementsHandler) Handle(ctx context.Context, q *GetEntitlements) (Entitlements, error) {
	if !q.Internal {
		if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingRead); err != nil {
			return Entitlements{}, err
		}
	}
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return Entitlements{}, err
	}
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return Entitlements{}, err
	}
	policy, err := h.read.Policy(ctx)
	if err != nil {
		return Entitlements{}, err
	}
	return newEntitlements(domain.ResolveEntitlements(q.WorkspaceID, sub, plans, policy, time.Now())), nil
}
