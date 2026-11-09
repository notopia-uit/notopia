package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCancelProviderHandler(f *cancellerFixture) *CancelProviderSubscriptionHandler {
	return NewCancelProviderSubscriptionHandler(f.store, f.store.Subscription(), f.provider)
}

func TestCancelProviderSubscriptionMarksSyncedOnceAndIsIdempotent(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()
	handler := newCancelProviderHandler(f)
	cmd := &CancelProviderSubscription{SubscriptionID: f.sub.ID()}

	require.NoError(t, handler.Handle(context.Background(), cmd))
	require.NoError(t, handler.Handle(context.Background(), cmd), "already synced: no second provider call")

	assert.False(t, f.store.subs[f.sub.ID()].AwaitingProviderCancel())
}

func TestCancelProviderSubscriptionReturnsProviderErrorWithoutSyncing(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.cancelLocally(t)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(errors.New("stripe down")).Once()

	err := newCancelProviderHandler(f).Handle(context.Background(), &CancelProviderSubscription{SubscriptionID: f.sub.ID()})

	require.Error(t, err)
	assert.True(t, f.store.subs[f.sub.ID()].AwaitingProviderCancel())
}

func TestCancelProviderSubscriptionSkipsLifetimeAndCurrentSubscriptions(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	now := time.Now()
	lifetimePrice, err := domain.NewPrice(uuid.New(), domain.PlanTierPro, domain.IntervalLifetime, vnd(t, 2490000), "price_life", now, "u")
	require.NoError(t, err)
	lifetime, err := domain.NewLifetimeSubscription(uuid.New(), uuid.New(), lifetimePrice, "pi_life", now, now, "u")
	require.NoError(t, err)
	quote, err := domain.NewRefundQuote(nil, 0, 0)
	require.NoError(t, err)
	_, err = lifetime.CancelNow(now, domain.RefundReasonLifetimeWindow, quote, "u")
	require.NoError(t, err)
	require.NoError(t, f.store.Subscription().Save(context.Background(), lifetime))
	handler := newCancelProviderHandler(f)

	require.NoError(t, handler.Handle(context.Background(), &CancelProviderSubscription{SubscriptionID: lifetime.ID()}))
	require.NoError(t, handler.Handle(context.Background(), &CancelProviderSubscription{SubscriptionID: f.sub.ID()}))
	// no CancelSubscriptionNow expectation: the mock fails the test if it is called
}
