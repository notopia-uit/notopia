package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspacePayer struct {
	WorkspaceID uuid.UUID
}

type GetWorkspacePayerHandler struct{ read BillingReadModel }

func NewGetWorkspacePayerHandler(read BillingReadModel) *GetWorkspacePayerHandler {
	return &GetWorkspacePayerHandler{read: read}
}

var ProvideGetWorkspacePayerHandler = NewGetWorkspacePayerHandler

type GetWorkspacePayerQuery commonhandler.Query[GetWorkspacePayer, PayerView]

var _ GetWorkspacePayerQuery = (*GetWorkspacePayerHandler)(nil)

func (h *GetWorkspacePayerHandler) Handle(ctx context.Context, q *GetWorkspacePayer) (PayerView, error) {
	var v PayerView
	account, err := h.read.Account(ctx, q.WorkspaceID)
	if err != nil || account == nil {
		return v, err
	}
	id := account.Payer().ID
	v.PayerID = &id
	sub, err := h.read.CurrentSubscription(ctx, q.WorkspaceID)
	if err != nil {
		return v, err
	}
	v.RequiresPayer = sub != nil && sub.Interval() != domain.IntervalLifetime && !sub.CancelAtPeriodEnd()
	return v, nil
}
