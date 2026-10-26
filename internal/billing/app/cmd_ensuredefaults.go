package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type EnsureDefaults struct{}

type EnsureDefaultsHandler struct{ uow domain.UnitOfWork }

func NewEnsureDefaultsHandler(uow domain.UnitOfWork) *EnsureDefaultsHandler {
	return &EnsureDefaultsHandler{uow: uow}
}

var ProvideEnsureDefaultsHandler = NewEnsureDefaultsHandler

type EnsureDefaultsCmd commonhandler.Cmd[EnsureDefaults]

var _ EnsureDefaultsCmd = (*EnsureDefaultsHandler)(nil)

func (h *EnsureDefaultsHandler) Handle(ctx context.Context, _ *EnsureDefaults) error {
	now := time.Now()
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		plans, err := r.Plan().ListMap(ctx)
		if err != nil {
			return err
		}
		for _, p := range domain.DefaultPlans(now) {
			if _, ok := plans[p.Tier()]; ok {
				continue
			}
			if err := r.Plan().Save(ctx, p); err != nil {
				return err
			}
		}
		if _, err := r.BillingPolicy().Get(ctx, false); err != nil {
			if !hasCode(err, errs.CodeNotFound) {
				return err
			}
			return r.BillingPolicy().Save(ctx, domain.DefaultBillingPolicy(now))
		}
		return nil
	})
}
