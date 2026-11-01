package app

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/require"
)

func vnd(t *testing.T, amount int64) domain.Money {
	t.Helper()
	m, err := domain.NewMoney(amount, domain.CurrencyVND)
	require.NoError(t, err)
	return m
}

func newTestPayment(t *testing.T, amount int64, ref string, paidAt, start, end time.Time) domain.PaymentInfo {
	t.Helper()
	p, err := domain.NewPaymentInfo(domain.PaymentInfoParams{
		Amount: vnd(t, amount), PaymentRef: ref, PaidAt: paidAt, CoverageStart: start, CoverageEnd: end,
	})
	require.NoError(t, err)
	return p
}

func newTestSnapshot(t *testing.T, subID, providerPriceID string, ws uuid.UUID, status domain.SubscriptionStatus, start, end time.Time) domain.ProviderSnapshot {
	t.Helper()
	s, err := domain.NewProviderSnapshot(domain.ProviderSnapshotParams{
		ProviderSubscriptionID: subID, ProviderPriceID: providerPriceID, WorkspaceID: ws, Status: status,
		CurrentPeriodStart: start, CurrentPeriodEnd: end,
	})
	require.NoError(t, err)
	return s
}

func newTestQuote(t *testing.T, paymentRef string, paid, refund int64) domain.RefundQuote {
	t.Helper()
	line, err := domain.NewRefundLine(domain.RefundLineParams{PaymentRef: paymentRef, PaidAmount: vnd(t, paid), Amount: vnd(t, refund)})
	require.NoError(t, err)
	q, err := domain.NewRefundQuote([]domain.RefundLine{line}, 0, 0)
	require.NoError(t, err)
	return q
}
