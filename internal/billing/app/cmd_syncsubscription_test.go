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
	_, err := f.sub.CancelNow(time.Now(), domain.RefundReasonCancelNow, domain.RefundQuote{}, "u")
	require.NoError(t, err)
	require.NoError(t, f.store.Subscription().Save(context.Background(), f.sub))
}

func (f *cancellerFixture) providerSnapshot(status domain.Status) domain.ProviderSnapshot {
	now := time.Now()
	return domain.ProviderSnapshot{
		ProviderSubscriptionID: "sub_1", ProviderPriceID: "price_plus", WorkspaceID: f.ws, Status: status,
		CurrentPeriodStart: now, CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
	}
}

func newSyncHandler(t *testing.T, f *cancellerFixture) *SyncSubscriptionHandler {
	t.Helper()
	notifier := NewMockNotifier(t)
	notifier.EXPECT().SubscriptionChanged(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	return NewSyncSubscriptionHandler(f.store, f.provider, notifier, &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"})
}

func TestSyncCancelsAtProviderWhenLocallyCanceledButProviderStillActive(t *testing.T) {
	f := newCancellerFixture(t, domain.StatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, "sub_1").Return(f.providerSnapshot(domain.StatusActive), nil).Once()
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()

	err := newSyncHandler(t, f).Handle(context.Background(), &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: "sub_1"})

	require.NoError(t, err)
	assert.Equal(t, domain.StatusCanceled, f.store.subs[f.sub.ID()].Status(), "local cancel is never resurrected")
}

func TestSyncMarksLocallyCanceledSubscriptionSyncedOnceProviderCanceled(t *testing.T) {
	f := newCancellerFixture(t, domain.StatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, "sub_1").Return(f.providerSnapshot(domain.StatusCanceled), nil).Once()

	err := newSyncHandler(t, f).Handle(context.Background(), &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: "sub_1"})

	require.NoError(t, err)
	saved := f.store.subs[f.sub.ID()]
	assert.Equal(t, domain.StatusCanceled, saved.Status())
	assert.True(t, saved.ProviderSyncedAt().After(*saved.CanceledAt()))
}

func TestReconcileCancelsLocallyCanceledSubscriptionAtProvider(t *testing.T) {
	f := newCancellerFixture(t, domain.StatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, "sub_1").Return(f.providerSnapshot(domain.StatusActive), nil).Once()
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()
	settings := &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"}
	handler := NewReconcileSubscriptionsHandler(f.store, newSyncHandler(t, f), NewRefundRunner(f.store, f.provider, settings))

	require.NoError(t, handler.Handle(context.Background(), &ReconcileSubscriptions{}))
}
