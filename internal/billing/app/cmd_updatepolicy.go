package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdatePolicy struct {
	PastDueGraceDays         *int
	LifetimeRefundWindowDays *int
	UserID                   string
}

type UpdatePolicyHandler struct{ uow domain.UnitOfWork }

func NewUpdatePolicyHandler(uow domain.UnitOfWork) *UpdatePolicyHandler {
	return &UpdatePolicyHandler{uow: uow}
}

var ProvideUpdatePolicyHandler = NewUpdatePolicyHandler

type UpdatePolicyCmd commonhandler.Cmd[UpdatePolicy]

var _ UpdatePolicyCmd = (*UpdatePolicyHandler)(nil)

func (h *UpdatePolicyHandler) Handle(ctx context.Context, cmd *UpdatePolicy) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		policy, err := r.BillingPolicy().Get(ctx, true)
		if err != nil {
			return err
		}
		if err := policy.Update(cmd.PastDueGraceDays, cmd.LifetimeRefundWindowDays, time.Now(), cmd.UserID); err != nil {
			return err
		}
		return r.BillingPolicy().Save(ctx, policy)
	})
}
