package app

import (
	"context"
	"testing"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (f *cancellerFixture) cancelLocally(t *testing.T) {
	t.Helper()
	quote, err := domain.NewRefundQuote(nil, 0, 0)
	require.NoError(t, err)
	_, err = f.sub.CancelNow(time.Now(), domain.RefundReasonCancelNow, quote, "u")
	require.NoError(t, err)
	require.NoError(t, f.store.Subscription().Save(context.Background(), f.sub))
}

func (f *cancellerFixture) providerSnapshot(t *testing.T, status domain.SubscriptionStatus) domain.ProviderSnapshot {
	t.Helper()
	now := time.Now()
	return newTestSnapshot(t, "sub_1", "price_plus", f.ws, status, now, now.Add(30*24*time.Hour))
}

func newSyncHandler(t *testing.T, f *cancellerFixture) *SyncSubscriptionHandler {
	t.Helper()
	return NewSyncSubscriptionHandler(f.store, f.provider, &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"})
}

func TestSyncNeverResurrectsLocallyCanceledSubscription(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, "sub_1").Return(f.providerSnapshot(t, domain.SubscriptionStatusActive), nil).Once()

	err := newSyncHandler(t, f).Handle(context.Background(), &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: "sub_1"})

	require.NoError(t, err)
	assert.Equal(t, domain.SubscriptionStatusCanceled, f.store.subs[f.sub.ID()].Status(), "local cancel is never resurrected")
}

func TestSyncMarksLocallyCanceledSubscriptionSyncedOnceProviderCanceled(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, "sub_1").Return(f.providerSnapshot(t, domain.SubscriptionStatusCanceled), nil).Once()

	err := newSyncHandler(t, f).Handle(context.Background(), &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: "sub_1"})

	require.NoError(t, err)
	saved := f.store.subs[f.sub.ID()]
	assert.Equal(t, domain.SubscriptionStatusCanceled, saved.Status())
	assert.True(t, saved.ProviderSyncedAt().After(*saved.CanceledAt()))
}

func TestReconcileCancelsLocallyCanceledSubscriptionAtProvider(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()
	handler := NewReconcileSubscriptionsHandler(f.store.Subscription(), newSyncHandler(t, f), newCancelProviderHandler(f), newSubmitRefundHandler(f))

	require.NoError(t, handler.Handle(context.Background(), &ReconcileSubscriptions{}))

	assert.False(t, f.store.subs[f.sub.ID()].AwaitingProviderCancel())
}
