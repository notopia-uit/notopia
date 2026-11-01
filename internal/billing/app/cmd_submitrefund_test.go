package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSubmitRefundHandler(f *cancellerFixture) *SubmitRefundHandler {
	return NewSubmitRefundHandler(f.store, f.store.Subscription(), f.provider, &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"})
}

func (f *cancellerFixture) requestRefund(t *testing.T) *SubmitRefund {
	t.Helper()
	refunds, err := f.sub.CancelNow(time.Now(), domain.RefundReasonCancelNow, newTestQuote(t, "pi_1", 30000, 10000), "u")
	require.NoError(t, err)
	require.NoError(t, f.store.Subscription().Save(context.Background(), f.sub))
	return &SubmitRefund{SubscriptionID: f.sub.ID(), RefundID: refunds[0].ID()}
}

func (f *cancellerFixture) refund(t *testing.T, id uuid.UUID) *domain.Refund {
	t.Helper()
	r := f.store.subs[f.sub.ID()].Refund(id)
	require.NotNil(t, r)
	return r
}

func TestSubmitRefundRecordsProviderResultAndIsIdempotent(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	cmd := f.requestRefund(t)
	f.provider.EXPECT().CreateRefund(mock.Anything, cmd.RefundID, "pi_1", mock.Anything).
		Return(RefundResult{ProviderRefundID: "re_1", Status: domain.RefundStatusSucceeded}, nil).Once()
	handler := newSubmitRefundHandler(f)

	require.NoError(t, handler.Handle(context.Background(), cmd))
	require.NoError(t, handler.Handle(context.Background(), cmd), "submitted: no second provider call")

	r := f.refund(t, cmd.RefundID)
	assert.Equal(t, domain.RefundStatusSucceeded, r.Status())
	require.NotNil(t, r.ProviderRefundID())
	assert.Equal(t, "re_1", *r.ProviderRefundID())
}

func TestSubmitRefundFailsRefundOnPermanentProviderError(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	cmd := f.requestRefund(t)
	f.provider.EXPECT().CreateRefund(mock.Anything, cmd.RefundID, "pi_1", mock.Anything).
		Return(RefundResult{}, errs.New(errs.CodeInvalid, "stripe create refund rejected", errors.New("amount exceeds charge"))).Once()

	require.NoError(t, newSubmitRefundHandler(f).Handle(context.Background(), cmd))

	r := f.refund(t, cmd.RefundID)
	assert.Equal(t, domain.RefundStatusFailed, r.Status())
	require.NotNil(t, r.FailureMessage())
	assert.Contains(t, *r.FailureMessage(), "amount exceeds charge")
	assert.Nil(t, r.ProviderRefundID())
}

func TestSubmitRefundReturnsTransientProviderErrorAndKeepsRefundPending(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	cmd := f.requestRefund(t)
	f.provider.EXPECT().CreateRefund(mock.Anything, cmd.RefundID, "pi_1", mock.Anything).
		Return(RefundResult{}, errs.New(errs.CodePaymentProvider, "stripe create refund failed", errors.New("timeout"))).Once()

	require.Error(t, newSubmitRefundHandler(f).Handle(context.Background(), cmd))

	assert.True(t, f.refund(t, cmd.RefundID).AwaitingSubmission())
}

func TestSubmitRefundSkipsExternalRefund(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	f.sub.RecordExternalRefund("re_ext", "pi_1", vnd(t, 5000), domain.RefundStatusPending, time.Now(), "u")
	require.NoError(t, f.store.Subscription().Save(context.Background(), f.sub))
	external := f.sub.RefundByProviderID("re_ext")

	require.NoError(t, newSubmitRefundHandler(f).Handle(context.Background(), &SubmitRefund{SubscriptionID: f.sub.ID(), RefundID: external.ID()}))
	// no CreateRefund expectation: the mock fails the test if it is called
}

func TestReconcileSubmitsPendingRefunds(t *testing.T) {
	f := newCancellerFixture(t, domain.SubscriptionStatusActive)
	cmd := f.requestRefund(t)
	f.provider.EXPECT().CancelSubscriptionNow(mock.Anything, "sub_1").Return(nil).Once()
	f.provider.EXPECT().CreateRefund(mock.Anything, cmd.RefundID, "pi_1", mock.Anything).
		Return(RefundResult{ProviderRefundID: "re_1", Status: domain.RefundStatusPending}, nil).Once()
	handler := NewReconcileSubscriptionsHandler(f.store.Subscription(), newSyncHandler(t, f), newCancelProviderHandler(f), newSubmitRefundHandler(f))

	require.NoError(t, handler.Handle(context.Background(), &ReconcileSubscriptions{}))

	assert.False(t, f.refund(t, cmd.RefundID).AwaitingSubmission())
}
