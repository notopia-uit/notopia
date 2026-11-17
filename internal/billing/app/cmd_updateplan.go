package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdatePlan struct {
	Tier     PlanTier
	Features Features
	Limits   Limits
	UserID   string
}

type UpdatePlanHandler struct{ uow domain.UnitOfWork }

func NewUpdatePlanHandler(uow domain.UnitOfWork) *UpdatePlanHandler {
	return &UpdatePlanHandler{uow: uow}
}

var ProvideUpdatePlanHandler = NewUpdatePlanHandler

type UpdatePlanCmd commonhandler.Cmd[UpdatePlan]

var _ UpdatePlanCmd = (*UpdatePlanHandler)(nil)

func (h *UpdatePlanHandler) Handle(ctx context.Context, cmd *UpdatePlan) error {
	features := toDomainFeatureUpdates(cmd.Features)
	limits := toDomainLimitUpdates(cmd.Limits)
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		plan, err := r.Plan().Get(ctx, toDomainPlanTier(cmd.Tier), true)
		if err != nil {
			return err
		}
		if err := plan.Update(features, limits, time.Now(), cmd.UserID); err != nil {
			return err
		}
		return r.Plan().Save(ctx, plan)
	})
}
