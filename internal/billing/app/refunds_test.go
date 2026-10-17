package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func runRefundWithError(t *testing.T, createErr error) *domain.Refund {
	t.Helper()
	f := newCancellerFixture(t, domain.StatusActive)
	quote := domain.RefundQuote{Lines: []domain.RefundLine{{
		PaymentRef: "pi_1", PaidAmount: domain.Money{Amount: 30000, Currency: "vnd"}, Amount: domain.Money{Amount: 10000, Currency: "vnd"},
	}}}
	_, err := f.sub.CancelNow(time.Now(), domain.RefundReasonCancelNow, quote, "u")
	require.NoError(t, err)
	require.NoError(t, f.store.Subscription().Save(context.Background(), f.sub))
	f.provider.EXPECT().CreateRefund(mock.Anything, mock.Anything, "pi_1", mock.Anything).Return(RefundResult{}, createErr).Once()
	runner := NewRefundRunner(f.store, f.provider, &Settings{WebBaseURL: "http://web.test", ServiceAccountUserID: "1003"})

	require.NoError(t, runner.Run(context.Background(), f.sub.ID()))

	refunds := f.store.subs[f.sub.ID()].Refunds()
	require.Len(t, refunds, 1)
	return refunds[0]
}

func TestRefundRunnerFailsRefundOnPermanentProviderError(t *testing.T) {
	r := runRefundWithError(t, errs.New(errs.CodeInvalid, "stripe create refund failed", errors.New("amount exceeds charge")))

	assert.Equal(t, domain.RefundStatusFailed, r.Status())
	require.NotNil(t, r.FailureMessage())
	assert.Contains(t, *r.FailureMessage(), "amount exceeds charge")
	assert.Nil(t, r.ProviderRefundID())
}

func TestRefundRunnerKeepsRefundPendingOnTransientProviderError(t *testing.T) {
	r := runRefundWithError(t, errs.New(errs.CodePaymentProvider, "stripe create refund failed", errors.New("timeout")))

	assert.Equal(t, domain.RefundStatusPending, r.Status())
	assert.Nil(t, r.ProviderRefundID())
}
