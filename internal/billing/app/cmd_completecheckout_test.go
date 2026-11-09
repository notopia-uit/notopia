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

var testPayer = func() domain.Payer {
	p, err := domain.NewPayer("112", "a@b.c", "A")
	if err != nil {
		panic(err)
	}
	return p
}()

type fixture struct {
	store    *memStore
	provider *MockPaymentProvider
	handler  *CompleteCheckoutHandler
	ws       uuid.UUID
	plus     *domain.Price
	lifetime *domain.Price
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	store := newMemStore()
	now := time.Now()
	plus, _ := domain.NewPrice(uuid.New(), domain.PlanTierPlus, domain.IntervalMonth, vnd(t, 30000), "price_plus", now, "u")
	lifetime, _ := domain.NewPrice(uuid.New(), domain.PlanTierPro, domain.IntervalLifetime, vnd(t, 2490000), "price_life", now, "u")
	store.prices[plus.ID()], store.prices[lifetime.ID()] = plus, lifetime
	ws := uuid.New()
	store.accounts[ws] = domain.OpenBillingAccount(ws, "cus_1", testPayer, now, "112")

	provider := NewMockPaymentProvider(t)
	provider.EXPECT().UpdateCustomerContact(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	settings := &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"}
	return &fixture{
		store:    store,
		provider: provider,
		ws:       ws,
		plus:     plus,
		lifetime: lifetime,
		handler:  NewCompleteCheckoutHandler(store, store.Subscription(), store.BillingAccount(), provider, settings),
	}
}

// expectSubscribeSession stubs a paid plus-monthly checkout that created Stripe subscription subID.
func (f *fixture) expectSubscribeSession(t *testing.T, subID string) {
	now := time.Now()
	end := now.Add(30 * 24 * time.Hour)
	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_"+subID).Return(CheckoutSessionResult{
		ID: "cs_" + subID, Purpose: CheckoutPurposeSubscribe, Complete: true, WorkspaceID: f.ws, PriceID: f.plus.ID(),
		Payer: testPayer, CustomerID: "cus_1", ProviderSubscriptionID: subID,
	}, nil)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, subID).Return(newTestSnapshot(t, subID, "price_plus", f.ws, domain.SubscriptionStatusActive, now, end), nil).Maybe()
	f.provider.EXPECT().PeriodPayments(mock.Anything, subID, mock.Anything, mock.Anything).Return([]domain.PaymentInfo{newTestPayment(t, 30000, "pi_"+subID, now, now, end)}, nil).Maybe()
}

func (f *fixture) current(t *testing.T) *domain.Subscription {
	t.Helper()
	s, _ := f.store.Subscription().GetCurrentByWorkspace(context.Background(), f.ws, false)
	return s
}

func (f *fixture) byProviderSubscription(id string) *domain.Subscription {
	s, _ := f.store.Subscription().GetByProviderSubscriptionID(context.Background(), id, false)
	return s
}

func TestCompleteCheckoutIsIdempotent(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession(t, "sub_1")
	cmd := &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}
	for range 2 {
		require.NoError(t, f.handler.Handle(context.Background(), cmd))
	}
	assert.Len(t, f.store.subs, 1)
	assert.Equal(t, domain.PlanTierPlus, f.current(t).Plan())
	// no CancelSubscriptionNow / CreateRefund expectation: the mock fails the test if either is called
}

func TestDuplicateCheckoutIsCanceledWithPendingRefund(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession(t, "sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.expectSubscribeSession(t, "sub_2")
	f.store.published = nil
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_sub_2"}))

	assert.Equal(t, "sub_1", *f.current(t).ProviderSubscriptionID())
	dup := f.byProviderSubscription("sub_2")
	require.NotNil(t, dup)
	assert.Equal(t, domain.SubscriptionStatusCanceled, dup.Status())
	require.Len(t, dup.Refunds(), 1)
	assert.True(t, dup.Refunds()[0].AwaitingSubmission())
	assert.EqualValues(t, 30000, dup.Refunds()[0].Amount().Amount())
	assert.True(t, f.store.hasPublished(func(e domain.Event) bool {
		_, ok := e.(*domain.SubscriptionCanceledNowEvent)
		return ok && e.GetAggregateID() == dup.ID().String()
	}))
	assert.False(t, f.store.hasPublished(func(e domain.Event) bool {
		_, ok := e.(*domain.SubscriptionStateChangedEvent)
		return ok
	}), "a duplicate must not report the workspace's subscription state")
	// no CancelSubscriptionNow / CreateRefund expectation: both are driven by the outbox events
}

func TestLifetimeReplacesRecurring(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession(t, "sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_life").Return(CheckoutSessionResult{
		ID: "cs_life", Purpose: CheckoutPurposeLifetime, Complete: true, WorkspaceID: f.ws, PriceID: f.lifetime.ID(),
		Payer: testPayer, CustomerID: "cus_1",
		Payment: newTestPayment(t, 2490000, "pi_life", time.Now(), time.Time{}, time.Time{}),
	}, nil)
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_life"}))

	cur := f.current(t)
	assert.Equal(t, domain.IntervalLifetime, cur.Interval())
	assert.Equal(t, domain.PlanTierPro, cur.Plan())
	old := f.byProviderSubscription("sub_1")
	require.NotNil(t, old)
	assert.Equal(t, domain.SubscriptionStatusCanceled, old.Status())
	require.Len(t, old.Refunds(), 1)
	assert.True(t, old.Refunds()[0].AwaitingSubmission())
	assert.Positive(t, old.Refunds()[0].Amount().Amount())
}
