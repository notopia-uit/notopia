package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type cancellerFixture struct {
	store                 *memStore
	provider              *MockPaymentProvider
	subscriptionCanceller *SubscriptionCanceller
	ws                    uuid.UUID
	sub                   *domain.Subscription
}

func newCancellerFixture(t *testing.T, status domain.SubscriptionStatus) *cancellerFixture {
	t.Helper()
	store := newMemStore()
	now := time.Now()
	plus, _ := domain.NewPrice(uuid.New(), domain.PlanTierPlus, domain.IntervalMonth, vnd(t, 30000), "price_plus", now, "u")
	store.prices[plus.ID()] = plus
	ws := uuid.New()
	snap := newTestSnapshot(t, "sub_1", "price_plus", ws, status, now, now.Add(30*24*time.Hour))
	sub, err := domain.NewSubscription(uuid.New(), ws, plus, snap, now, "u")
	require.NoError(t, err)
	require.NoError(t, store.Subscription().Save(context.Background(), sub))

	provider := NewMockPaymentProvider(t)
	return &cancellerFixture{
		store:                 store,
		provider:              provider,
		subscriptionCanceller: NewSubscriptionCanceller(store, store.Subscription(), store.BillingPolicy(), provider),
		ws:                    ws,
		sub:                   sub,
	}
}

func TestCancelNowPastDueWithoutPeriodPayments(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusPastDue)
	f.provider.EXPECT().PeriodPayments(mock.Anything, "sub_1", mock.Anything, mock.Anything).Return(nil, nil).Once()

	err := f.subscriptionCanceller.CancelNow(context.Background(), f.ws, CancelNowCauseWorkspaceDeleted, "u")

	require.NoError(t, err)
	assert.Equal(t, domain.SubscriptionStatusCanceled, f.store.subs[f.sub.ID()].Status())
	assert.Empty(t, f.store.subs[f.sub.ID()].Refunds())
}

func TestCancelNowEndsAtCommitWithPendingRefund(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	now := time.Now()
	f.provider.EXPECT().PeriodPayments(mock.Anything, "sub_1", mock.Anything, mock.Anything).Return([]domain.PaymentInfo{newTestPayment(t, 30000, "pi_1", now, now, now.Add(30*24*time.Hour))}, nil).Once()
	f.store.published = nil

	err := f.subscriptionCanceller.CancelNow(context.Background(), f.ws, CancelNowCauseUser, "u")

	require.NoError(t, err)
	saved := f.store.subs[f.sub.ID()]
	assert.Equal(t, domain.SubscriptionStatusCanceled, saved.Status())
	assert.True(t, saved.AwaitingProviderCancel())
	require.Len(t, saved.Refunds(), 1)
	refundID := saved.Refunds()[0].ID()
	assert.True(t, saved.Refunds()[0].AwaitingSubmission())
	assert.True(t, f.store.hasPublished(func(e domain.Event) bool {
		_, ok := e.(*domain.SubscriptionCanceledNowEvent)
		return ok
	}))
	assert.True(t, f.store.hasPublished(func(e domain.Event) bool {
		r, ok := e.(*domain.RefundRequestedEvent)
		return ok && r.RefundID == refundID
	}))
}
