package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspaceGraph struct {
	ID            uuid.UUID
	IgnoreOrphans bool

	UserID string
}

type GetWorkspaceGraphHandler struct {
	authorizationSvc AuthorizationSvc
	billingSvc       BillingSvc
	readModel        GetWorkspaceGraphReadModel
}

func NewGetWorkspaceGraphHandler(
	authorizationSvc AuthorizationSvc,
	billingSvc BillingSvc,
	readModel GetWorkspaceGraphReadModel,
) *GetWorkspaceGraphHandler {
	return &GetWorkspaceGraphHandler{
		authorizationSvc: authorizationSvc,
		billingSvc:       billingSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceGraphHandler = NewGetWorkspaceGraphHandler

type GetWorkspaceGraphQuery commonhandler.Query[GetWorkspaceGraph, Graph]

var _ GetWorkspaceGraphQuery = (*GetWorkspaceGraphHandler)(nil)

func (h *GetWorkspaceGraphHandler) Handle(ctx context.Context, query *GetWorkspaceGraph) (Graph, error) {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.ID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return Graph{}, err
	}
	if !hasPermission {
		return Graph{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace graph %s", query.UserID, query.ID),
		)
	}
	ent, err := h.billingSvc.GetEntitlements(ctx, query.ID)
	if err != nil {
		return Graph{}, err
	}
	if !ent.Feature(PlanFeatureWorkspaceGraph) {
		plans, err := h.billingSvc.ListPlans(ctx)
		if err != nil {
			return Graph{}, err
		}
		return Graph{}, errs.NewFeatureNotInPlan(
			PlanFeatureWorkspaceGraph.String(),
			RequiredPlanForFeature(plans, PlanFeatureWorkspaceGraph).String(),
		)
	}
	graph, err := h.readModel.Handle(ctx, &GetWorkspaceGraphReadModelParams{
		ID:            query.ID,
		IgnoreOrphans: query.IgnoreOrphans,
	})
	if err != nil {
		return Graph{}, err
	}
	return graph, nil
}
