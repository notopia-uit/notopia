package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiredTrashCutoffs(t *testing.T) {
	now := time.Date(2026, 10, 18, 3, 0, 0, 0, time.UTC)
	zero, week, unlimited, missing := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ents := map[uuid.UUID]Entitlements{
		zero:      {Limits: map[PlanLimit]PlanQuota{PlanLimitTrashRetentionDays: 0}},
		week:      {Limits: map[PlanLimit]PlanQuota{PlanLimitTrashRetentionDays: 7}},
		unlimited: {Limits: map[PlanLimit]PlanQuota{PlanLimitTrashRetentionDays: PlanQuotaUnlimited}},
	}

	got := expiredTrashCutoffs(context.Background(), now, []uuid.UUID{zero, week, unlimited, missing}, ents)

	require.Len(t, got, 2)
	assert.Equal(t, workspaceCutoff{workspaceID: zero, cutoff: now.Add(-24 * time.Hour)}, got[0])
	assert.Equal(t, workspaceCutoff{workspaceID: week, cutoff: now.Add(-7 * 24 * time.Hour)}, got[1])
}
