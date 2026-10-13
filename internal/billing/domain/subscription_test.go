package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

func price(t *testing.T, plan domain.PlanTier, interval domain.Interval, amount int64) *domain.Price {
	t.Helper()
	p, err := domain.NewPrice(uuid.New(), plan, interval, domain.Money{Amount: amount, Currency: "vnd"}, "price_x", t0, "u")
	require.NoError(t, err)
	return p
}

func snap(status domain.Status, start, end time.Time) domain.ProviderSnapshot {
	return domain.ProviderSnapshot{ProviderSubscriptionID: "sub_1", Status: status, CurrentPeriodStart: start, CurrentPeriodEnd: end}
}

func eventNames(events []domain.Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, fmtType(e))
	}
	return out
}

func fmtType(e domain.Event) string {
	switch e.(type) {
	case *domain.SubscriptionCreatedEvent:
		return "created"
	case *domain.SubscriptionActivatedEvent:
		return "activated"
	case *domain.SubscriptionRenewedEvent:
		return "renewed"
	case *domain.SubscriptionPastDueEvent:
		return "past_due"
	case *domain.SubscriptionPlanChangedEvent:
		return "plan_changed"
	case *domain.SubscriptionCancelScheduledEvent:
		return "cancel_scheduled"
	case *domain.SubscriptionEndedEvent:
		return "ended"
	default:
		return "other"
	}
}

func TestGrantsAccess(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 3 days grace
	end := t0.AddDate(0, 1, 0)
	cases := []struct {
		name   string
		status domain.Status
		now    time.Time
		want   bool
	}{
		{"active after stored period end (renewal webhook late)", domain.StatusActive, end.Add(2 * time.Hour), true},
		{"past_due within grace", domain.StatusPastDue, end.Add(48 * time.Hour), true},
		{"past_due after grace", domain.StatusPastDue, end.Add(96 * time.Hour), false},
		{"canceled", domain.StatusCanceled, t0, false},
		{"unpaid", domain.StatusUnpaid, t0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(c.status, t0, end), t0, "u")
			require.NoError(t, err)
			assert.Equal(t, c.want, s.GrantsAccess(c.now, policy))
		})
	}
}

func TestApplySnapshotEvents(t *testing.T) {
	plus := price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000)
	pro := price(t, domain.PlanTierPro, domain.IntervalMonth, 99000)
	end := t0.AddDate(0, 1, 0)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), plus, snap(domain.StatusActive, t0, end), t0, "u")
	assert.Equal(t, []string{"created", "activated"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.StatusActive, end, end.AddDate(0, 1, 0)), plus, end, "u")
	assert.Equal(t, []string{"renewed"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.StatusPastDue, end, end.AddDate(0, 1, 0)), pro, end, "u")
	assert.Equal(t, []string{"past_due", "plan_changed"}, eventNames(s.PopEvents()))

	ended := snap(domain.StatusCanceled, end, end.AddDate(0, 1, 0))
	s.ApplySnapshot(ended, pro, end, "u")
	assert.Equal(t, []string{"ended"}, eventNames(s.PopEvents()))
}

func TestQuoteCancelNow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	end := t0.Add(30 * 24 * time.Hour)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.StatusActive, t0, end), t0, "u")
	pay := domain.PaymentInfo{Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_1", PaidAt: t0, CoverageStart: t0, CoverageEnd: end}

	q, err := s.QuoteCancelNow(t0.Add(10*24*time.Hour), policy, []domain.PaymentInfo{pay})
	require.NoError(t, err)
	assert.EqualValues(t, 20000, q.Amount.Amount)

	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	// A second refund against the same payment must not exceed what was paid.
	over := domain.RefundQuote{Amount: domain.Money{Amount: 20000, Currency: "vnd"}, Lines: []domain.RefundLine{{PaymentRef: "pi_1", PaidAmount: pay.Amount, Amount: domain.Money{Amount: 20000, Currency: "vnd"}}}}
	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, over, "u")
	assert.Error(t, err, "non-current subscription / over-refund")
}

// Plus 30000 paid for day 0-30; upgrade on day 10 invoiced 40000 for day 10-30; cancel on day 20.
func TestQuoteCancelNowAfterUpgrade(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	day := func(n int) time.Time { return t0.Add(time.Duration(n) * 24 * time.Hour) }
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalMonth, 90000), snap(domain.StatusActive, t0, day(30)), t0, "u")
	payments := []domain.PaymentInfo{
		{Amount: domain.Money{Amount: 30000, Currency: "vnd"}, PaymentRef: "pi_a", PaidAt: t0, CoverageStart: t0, CoverageEnd: day(30)},
		{Amount: domain.Money{Amount: 40000, Currency: "vnd"}, PaymentRef: "pi_b", PaidAt: day(10), CoverageStart: day(10), CoverageEnd: day(30)},
	}

	q, err := s.QuoteCancelNow(day(20), policy, payments)
	require.NoError(t, err)
	require.Len(t, q.Lines, 2)
	assert.EqualValues(t, 10000, q.Lines[0].Amount.Amount)
	assert.EqualValues(t, 20000, q.Lines[1].Amount.Amount)
	assert.EqualValues(t, 30000, q.Amount.Amount)

	refunds, err := s.CancelNow(day(20), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	require.Len(t, refunds, 2)
	assert.Equal(t, "pi_b", refunds[1].PaymentRef())
}

func TestLifetimeRefundWindow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 14 days
	s, _ := domain.NewLifetimeSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalLifetime, 2490000), "pi_l", t0, t0, "u")
	pay := []domain.PaymentInfo{{Amount: domain.Money{Amount: 2490000, Currency: "vnd"}, PaymentRef: "pi_l", PaidAt: t0}}

	q, err := s.QuoteCancelNow(t0.Add(13*24*time.Hour), policy, pay)
	require.NoError(t, err)
	assert.EqualValues(t, 2490000, q.Amount.Amount)

	_, err = s.QuoteCancelNow(t0.Add(15*24*time.Hour), policy, pay)
	require.Error(t, err, "refundWindowExpired after 14 days")
	assert.True(t, s.GrantsAccess(t0.AddDate(5, 0, 0), policy), "lifetime grants access forever")
}
