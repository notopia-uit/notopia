package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

func checkMembersLimit(
	ctx context.Context,
	billingSvc BillingSvc,
	workspaceID uuid.UUID,
	current int,
	members []WorkspaceMemberUpdate,
) error {
	requested := distinctMemberCount(members)
	if requested <= current {
		return nil
	}
	ent, err := billingSvc.GetEntitlements(ctx, workspaceID)
	if err != nil {
		return err
	}
	quota := ent.Limit(PlanLimitMembers)
	if quota.Allows(int64(requested)) {
		return nil
	}
	plans, err := billingSvc.ListPlans(ctx)
	if err != nil {
		return err
	}
	return errs.NewPlanLimitExceeded(
		PlanLimitMembers.String(),
		int64(requested),
		int64(quota),
		RequiredPlanForLimit(plans, PlanLimitMembers, int64(requested)).String(),
	)
}

func distinctMemberCount(members []WorkspaceMemberUpdate) int {
	ids := make(map[string]struct{}, len(members))
	for _, m := range members {
		ids[m.ID] = struct{}{}
	}
	return len(ids)
}

func checkPayerKept(
	ctx context.Context,
	billingSvc BillingSvc,
	workspaceID uuid.UUID,
	members []WorkspaceMemberUpdate,
) error {
	payer, requiresPayer, err := billingSvc.GetPayer(ctx, workspaceID)
	if err != nil {
		return err
	}
	if payer == nil || !requiresPayer {
		return nil
	}
	for _, m := range members {
		if m.ID == *payer && m.Role == WorkspaceRoleOwner {
			return nil
		}
	}
	return errs.NewBillingPayerMustTransfer(workspaceID)
}

func checkPayerMayLeave(
	ctx context.Context,
	billingSvc BillingSvc,
	workspaceID uuid.UUID,
	userID string,
) error {
	payer, requiresPayer, err := billingSvc.GetPayer(ctx, workspaceID)
	if err != nil {
		return err
	}
	if requiresPayer && payer != nil && *payer == userID {
		return errs.NewBillingPayerMustTransfer(workspaceID)
	}
	return nil
}
