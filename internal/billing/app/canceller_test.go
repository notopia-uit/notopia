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

type cancellerFixture struct {
	store     *memStore
	provider  *MockPaymentProvider
	canceller *Canceller
	ws        uuid.UUID
	sub       *domain.Subscription
}

func newCancellerFixture(t *testing.T, status domain.Status) *cancellerFixture {
	t.Helper()
	store := newMemStore()
	now := time.Now()
	plus, _ := domain.NewPrice(uuid.New(), domain.PlanTierPlus, domain.IntervalMonth, domain.Money{Amount: 30000, Currency: "vnd"}, "price_plus", now, "u")
	store.prices[plus.ID()] = plus
	ws := uuid.New()
	snap := domain.ProviderSnapshot{
		ProviderSubscriptionID: "sub_1", ProviderPriceID: "price_plus", WorkspaceID: ws, Status: status,
		CurrentPeriodStart: now, CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
	}
	sub, err := domain.NewSubscription(uuid.New(), ws, plus, snap, now, "u")
	require.NoError(t, err)
	require.NoError(t, store.Subscription().Save(context.Background(), sub))

	provider := NewMockPaymentProvider(t)
	notifier := NewMockNotifier(t)
	notifier.EXPECT().SubscriptionChanged(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	settings := &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"}
	return &cancellerFixture{
		store:     store,
		provider:  provider,
		canceller: NewCanceller(store, provider, notifier, NewRefundRunner(store, provider, settings)),
		ws:        ws,
		sub:       sub,
	}
}

func TestCancelNowPastDueWithoutPeriodPayments(t *testing.T) {
	f := newCancellerFixture(t, domain.StatusPastDue)
	f.provider.EXPECT().PeriodPayments(mock.Anything, "sub_1", mock.Anything, mock.Anything).Return(nil, nil).Once()
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()

	err := f.canceller.CancelNow(context.Background(), f.ws, domain.RefundReasonWorkspaceDeleted, "u", true)

	require.NoError(t, err)
	assert.Equal(t, domain.StatusCanceled, f.store.subs[f.sub.ID()].Status())
	assert.Empty(t, f.store.subs[f.sub.ID()].Refunds())
}

func TestCancelNowCancelsLocallyBeforeProviderAndSurvivesProviderFailure(t *testing.T) {
	f := newCancellerFixture(t, domain.StatusActive)
	now := time.Now()
	f.provider.EXPECT().PeriodPayments(mock.Anything, "sub_1", mock.Anything, mock.Anything).Return([]domain.PaymentInfo{{
		Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_1", PaidAt: now,
		CoverageStart: now, CoverageEnd: now.Add(30 * 24 * time.Hour),
	}}, nil).Once()
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").
		RunAndReturn(func(context.Context, string) error {
			assert.Equal(t, domain.StatusCanceled, f.store.subs[f.sub.ID()].Status(), "local cancel is committed first")
			return errors.New("stripe down")
		}).Once()
	submitted := 0
	f.provider.EXPECT().CreateRefund(mock.Anything, mock.Anything, "pi_1", mock.Anything).
		RunAndReturn(func(_ context.Context, id uuid.UUID, ref string, amount domain.Money) (RefundResult, error) {
			submitted++
			return RefundResult{
				ProviderRefundID: "re_1", RefundID: &id, PaymentRef: ref, Amount: amount,
				Status: domain.RefundStatusSucceeded, FailureMessage: nil, ProviderSubscriptionID: "",
			}, nil
		}).Once()

	err := f.canceller.CancelNow(context.Background(), f.ws, domain.RefundReasonCancelNow, "u", false)

	require.NoError(t, err)
	assert.Equal(t, 1, submitted)
	require.Len(t, f.store.subs[f.sub.ID()].Refunds(), 1)
	assert.Equal(t, domain.RefundStatusSucceeded, f.store.subs[f.sub.ID()].Refunds()[0].Status())
}
