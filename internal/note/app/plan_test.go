package app_test

import (
	"testing"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/stretchr/testify/assert"
)

func TestRequiredPlanForLimit(t *testing.T) {
	plans := []app.PlanEntitlements{
		{Tier: app.PlanTierFree, Limits: map[app.PlanLimit]app.PlanQuota{app.PlanLimitMembers: 3}},
		{Tier: app.PlanTierPlus, Limits: map[app.PlanLimit]app.PlanQuota{app.PlanLimitMembers: 10}},
		{Tier: app.PlanTierPro, Limits: map[app.PlanLimit]app.PlanQuota{app.PlanLimitMembers: app.PlanQuotaUnlimited}},
	}

	assert.Equal(t, app.PlanTierPlus, app.RequiredPlanForLimit(plans, app.PlanLimitMembers, 4))
	assert.Equal(t, app.PlanTierPlus, app.RequiredPlanForLimit(plans, app.PlanLimitMembers, 10))
	assert.Equal(t, app.PlanTierPro, app.RequiredPlanForLimit(plans, app.PlanLimitMembers, 11))
	assert.Equal(t, app.PlanTierPro, app.RequiredPlanForLimit(nil, app.PlanLimitMembers, 1))
}

func TestPlanQuotaAllows(t *testing.T) {
	assert.True(t, app.PlanQuotaUnlimited.Allows(1_000_000))
	assert.True(t, app.PlanQuota(5).Allows(5))
	assert.False(t, app.PlanQuota(5).Allows(6))
}
