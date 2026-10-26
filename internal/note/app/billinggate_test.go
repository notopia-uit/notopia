package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBilling struct {
	BillingSvc
	ent           Entitlements
	payer         *string
	requiresPayer bool
}

func (f fakeBilling) GetEntitlements(context.Context, uuid.UUID) (Entitlements, error) {
	return f.ent, nil
}

func (f fakeBilling) GetPayer(context.Context, uuid.UUID) (*string, bool, error) {
	return f.payer, f.requiresPayer, nil
}

func (f fakeBilling) ListPlans(context.Context) ([]PlanEntitlements, error) {
	return []PlanEntitlements{
		{Tier: PlanTierPlus, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: 10}},
		{Tier: PlanTierPro, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: PlanQuotaUnlimited}},
	}, nil
}

func TestMembersLimit(t *testing.T) {
	free := fakeBilling{ent: Entitlements{Plan: PlanTierFree, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: 3}}}
	ws := uuid.New()

	members := func(ids ...string) []WorkspaceMemberUpdate {
		out := make([]WorkspaceMemberUpdate, 0, len(ids))
		for _, id := range ids {
			out = append(out, WorkspaceMemberUpdate{ID: id, Role: WorkspaceRoleEditor})
		}
		return out
	}

	err := checkMembersLimit(context.Background(), free, ws, 3, members("1", "2", "3", "4"))
	pr, ok := errors.AsType[*errs.PaymentRequired](err)
	require.True(t, ok, "growing past limit")
	assert.Equal(t, PlanTierPlus.String(), pr.RequiredPlan)
	assert.EqualValues(t, 3, *pr.Max)

	require.NoError(t, checkMembersLimit(context.Background(), free, ws, 5, members("1", "2", "3", "4")), "shrinking an over-limit workspace")
	require.NoError(t, checkMembersLimit(context.Background(), free, ws, 5, members("1", "2", "3", "4", "5")), "same count (role change)")
	require.NoError(
		t,
		checkMembersLimit(context.Background(), free, ws, 3, members("1", "2", "3", "3", "1")),
		"duplicate ids count once",
	)
}

func TestPayerGuard(t *testing.T) {
	payer := "112"
	ws := uuid.New()
	ctx := context.Background()
	g := fakeBilling{payer: &payer, requiresPayer: true}

	demoted := []WorkspaceMemberUpdate{{ID: "112", Role: WorkspaceRoleEditor}, {ID: "111", Role: WorkspaceRoleOwner}}
	assert.True(t, isCode(checkPayerKept(ctx, g, ws, demoted), errs.CodeBillingPayerMustTransfer), "demoting payer")

	removed := []WorkspaceMemberUpdate{{ID: "111", Role: WorkspaceRoleOwner}}
	assert.True(t, isCode(checkPayerKept(ctx, g, ws, removed), errs.CodeBillingPayerMustTransfer), "removing payer")

	kept := []WorkspaceMemberUpdate{{ID: "112", Role: WorkspaceRoleOwner}, {ID: "111", Role: WorkspaceRoleEditor}}
	require.NoError(t, checkPayerKept(ctx, g, ws, kept), "payer stays owner")

	assert.True(t, isCode(checkPayerMayLeave(ctx, g, ws, "112"), errs.CodeBillingPayerMustTransfer), "payer leaving")
	require.NoError(t, checkPayerMayLeave(ctx, g, ws, "111"), "non-payer leaving")

	noCharge := fakeBilling{payer: &payer, requiresPayer: false}
	require.NoError(t, checkPayerMayLeave(ctx, noCharge, ws, "112"), "lifetime or cancelling leaving")
	require.NoError(t, checkPayerKept(ctx, noCharge, ws, demoted), "lifetime or cancelling demotion")
}

func isCode(err error, code errs.Code) bool {
	e, ok := errors.AsType[errs.Error](err)
	return ok && e.Code() == code
}
