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

func TestSyncRefundOfDashboardRefundOnRecurringPaymentReducesNextQuote(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.provider.EXPECT().GetRefund(mock.Anything, "re_dash").Return(RefundResult{
		ProviderRefundID: "re_dash", RefundID: nil, PaymentRef: "pi_1", Amount: vnd(t, 10000),
		Status: domain.RefundStatusSucceeded, FailureMessage: nil, ProviderSubscriptionID: "sub_1",
	}, nil).Once()
	handler := NewSyncRefundHandler(f.store, f.provider, &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"})

	err := handler.Handle(context.Background(), &SyncRefund{EventID: "evt_1", EventType: "refund.created", ProviderRefundID: "re_dash"})

	require.NoError(t, err)
	refunds := f.store.subs[f.sub.ID()].Refunds()
	require.Len(t, refunds, 1)
	assert.Equal(t, domain.RefundReasonExternal, refunds[0].Reason())
	assert.Equal(t, "pi_1", refunds[0].PaymentRef())

	now := time.Now()
	f.provider.EXPECT().PeriodPayments(mock.Anything, "sub_1", mock.Anything, mock.Anything).Return([]domain.PaymentInfo{newTestPayment(t, 30000, "pi_1", now, now, now.Add(30*24*time.Hour))}, nil).Once()
	quote, err := f.subscriptionCanceller.Quote(context.Background(), f.ws)
	require.NoError(t, err)
	assert.EqualValues(t, 20000, quote.Amount().Amount(), "quote is capped by what is left of the payment")
}
