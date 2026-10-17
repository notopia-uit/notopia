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

var testPayer = domain.Payer{ID: "112", Email: "a@b.c", Name: "A"}

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
	plus, _ := domain.NewPrice(uuid.New(), domain.PlanTierPlus, domain.IntervalMonth, domain.Money{Amount: 30000, Currency: "vnd"}, "price_plus", now, "u")
	lifetime, _ := domain.NewPrice(uuid.New(), domain.PlanTierPro, domain.IntervalLifetime, domain.Money{Amount: 2490000, Currency: "vnd"}, "price_life", now, "u")
	store.prices[plus.ID()], store.prices[lifetime.ID()] = plus, lifetime
	ws := uuid.New()
	store.accounts[ws] = domain.OpenBillingAccount(ws, "cus_1", testPayer, now, "112")

	provider := NewMockPaymentProvider(t)
	provider.EXPECT().UpdateCustomerContact(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	notifier := NewMockNotifier(t)
	notifier.EXPECT().SubscriptionChanged(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	settings := &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"}
	refunds := NewRefundRunner(store, provider, settings)
	return &fixture{
		store:    store,
		provider: provider,
		ws:       ws,
		plus:     plus,
		lifetime: lifetime,
		handler:  NewCompleteCheckoutHandler(store, provider, notifier, settings, refunds),
	}
}

// expectSubscribeSession stubs a paid plus-monthly checkout that created Stripe subscription subID.
func (f *fixture) expectSubscribeSession(subID string) {
	now := time.Now()
	end := now.Add(30 * 24 * time.Hour)
	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_"+subID).Return(CheckoutSessionResult{
		ID: "cs_" + subID, Purpose: CheckoutPurposeSubscribe, Complete: true, WorkspaceID: f.ws, PriceID: f.plus.ID(),
		Payer: testPayer, CustomerID: "cus_1", ProviderSubscriptionID: subID,
	}, nil)
	f.provider.EXPECT().GetSubscriptionSnapshot(mock.Anything, subID).Return(domain.ProviderSnapshot{
		ProviderSubscriptionID: subID, ProviderPriceID: "price_plus", WorkspaceID: f.ws, Status: domain.StatusActive,
		CurrentPeriodStart: now, CurrentPeriodEnd: end,
	}, nil).Maybe()
	f.provider.EXPECT().PeriodPayments(mock.Anything, subID, mock.Anything, mock.Anything).Return([]domain.PaymentInfo{{
		Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_" + subID, PaidAt: now,
		CoverageStart: now, CoverageEnd: end,
	}}, nil).Maybe()
}

// expectRefund accepts one refund and reports it succeeded; returns the captured amount.
func (f *fixture) expectRefund() *domain.Money {
	var got domain.Money
	f.provider.EXPECT().CreateRefund(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, id uuid.UUID, ref string, amount domain.Money) (RefundResult, error) {
			got = amount
			return RefundResult{
				ProviderRefundID: "re_" + id.String(), RefundID: &id, PaymentRef: ref,
				Amount: amount, Status: domain.RefundStatusSucceeded, FailureMessage: nil,
			}, nil
		}).Once()
	return &got
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
	f.expectSubscribeSession("sub_1")
	cmd := &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}
	for range 2 {
		require.NoError(t, f.handler.Handle(context.Background(), cmd))
	}
	assert.Len(t, f.store.subs, 1)
	assert.Equal(t, domain.PlanTierPlus, f.current(t).Plan())
	// no CancelSubscriptionNow / CreateRefund expectation: the mock fails the test if either is called
}

func TestDuplicateCheckoutIsCanceledAndRefunded(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession("sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.expectSubscribeSession("sub_2")
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_2").
		RunAndReturn(func(context.Context, string) error {
			dup := f.byProviderSubscription("sub_2")
			require.NotNil(t, dup, "duplicate row is committed before the provider cancel")
			assert.Equal(t, domain.StatusCanceled, dup.Status())
			return nil
		}).Once()
	refunded := f.expectRefund()
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_sub_2"}))

	assert.Equal(t, "sub_1", *f.current(t).ProviderSubscriptionID())
	assert.EqualValues(t, 30000, refunded.Amount)
}

func TestLifetimeReplacesRecurring(t *testing.T) {
	f := newFixture(t)
	f.expectSubscribeSession("sub_1")
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_1", SessionID: "cs_sub_1"}))

	f.provider.EXPECT().GetCheckoutSession(mock.Anything, "cs_life").Return(CheckoutSessionResult{
		ID: "cs_life", Purpose: CheckoutPurposeLifetime, Complete: true, WorkspaceID: f.ws, PriceID: f.lifetime.ID(),
		Payer: testPayer, CustomerID: "cus_1",
		Payment: domain.PaymentInfo{Amount: domain.Money{Amount: 2490000, Currency: "vnd"}, PaymentRef: "pi_life", PaidAt: time.Now()},
	}, nil)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").
		RunAndReturn(func(context.Context, string) error {
			old := f.byProviderSubscription("sub_1")
			require.NotNil(t, old)
			assert.Equal(t, domain.StatusCanceled, old.Status(), "old subscription is canceled locally before the provider cancel")
			assert.Equal(t, domain.IntervalLifetime, f.current(t).Interval(), "lifetime is already committed")
			return nil
		}).Once()
	refunded := f.expectRefund()
	require.NoError(t, f.handler.Handle(context.Background(), &CompleteCheckout{EventID: "evt_2", SessionID: "cs_life"}))

	cur := f.current(t)
	assert.Equal(t, domain.IntervalLifetime, cur.Interval())
	assert.Equal(t, domain.PlanTierPro, cur.Plan())
	assert.Positive(t, refunded.Amount)
}
