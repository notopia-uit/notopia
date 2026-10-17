package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type BatchGetEntitlements struct {
	WorkspaceIDs []uuid.UUID
}

type BatchGetEntitlementsHandler struct{ read BillingReadModel }

func NewBatchGetEntitlementsHandler(read BillingReadModel) *BatchGetEntitlementsHandler {
	return &BatchGetEntitlementsHandler{read: read}
}

var ProvideBatchGetEntitlementsHandler = NewBatchGetEntitlementsHandler

type BatchGetEntitlementsQuery commonhandler.Query[BatchGetEntitlements, map[uuid.UUID]Entitlements]

var _ BatchGetEntitlementsQuery = (*BatchGetEntitlementsHandler)(nil)

func (h *BatchGetEntitlementsHandler) Handle(ctx context.Context, q *BatchGetEntitlements) (map[uuid.UUID]Entitlements, error) {
	subs, err := h.read.CurrentSubscriptions(ctx, q.WorkspaceIDs)
	if err != nil {
		return nil, err
	}
	plans, err := h.read.Plans(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := h.read.Policy(ctx)
	if err != nil {
		return nil, err
	}
	byWorkspace := make(map[uuid.UUID]*domain.Subscription, len(subs))
	for _, s := range subs {
		byWorkspace[s.WorkspaceID()] = s
	}
	now := time.Now()
	out := make(map[uuid.UUID]Entitlements, len(q.WorkspaceIDs))
	for _, id := range q.WorkspaceIDs {
		out[id] = newEntitlements(domain.ResolveEntitlements(id, byWorkspace[id], plans, policy, now))
	}
	return out, nil
}
