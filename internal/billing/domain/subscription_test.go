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
	p, err := domain.NewPrice(uuid.New(), plan, interval, vnd(t, amount), "price_x", t0, "u")
	require.NoError(t, err)
	return p
}

func vnd(t *testing.T, amount int64) domain.Money {
	t.Helper()
	m, err := domain.NewMoney(amount, domain.CurrencyVND)
	require.NoError(t, err)
	return m
}

func snap(status domain.SubscriptionStatus, start, end time.Time) domain.ProviderSnapshot {
	s, err := domain.NewProviderSnapshot(domain.ProviderSnapshotParams{
		ProviderSubscriptionID: "sub_1", Status: status, CurrentPeriodStart: start, CurrentPeriodEnd: end,
	})
	if err != nil {
		panic(err)
	}
	return s
}

func payment(t *testing.T, amount int64, ref string, paidAt, start, end time.Time) domain.PaymentInfo {
	t.Helper()
	p, err := domain.NewPaymentInfo(domain.PaymentInfoParams{
		Amount: vnd(t, amount), PaymentRef: ref, PaidAt: paidAt, CoverageStart: start, CoverageEnd: end,
	})
	require.NoError(t, err)
	return p
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
	case *domain.SubscriptionStateChangedEvent:
		return "state_changed"
	case *domain.SubscriptionCanceledNowEvent:
		return "canceled_now"
	case *domain.RefundRequestedEvent:
		return "refund_requested"
	case *domain.RefundSucceededEvent:
		return "refund_succeeded"
	default:
		return "other"
	}
}

func TestGrantsAccess(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 3 days grace
	end := t0.AddDate(0, 1, 0)
	cases := []struct {
		name   string
		status domain.SubscriptionStatus
		now    time.Time
		want   bool
	}{
		{"active after stored period end (renewal webhook late)", domain.SubscriptionStatusActive, end.Add(2 * time.Hour), true},
		{"past_due within grace", domain.SubscriptionStatusPastDue, t0.Add(48 * time.Hour), true},
		{"past_due after grace", domain.SubscriptionStatusPastDue, t0.Add(96 * time.Hour), false},
		{"canceled", domain.SubscriptionStatusCanceled, t0, false},
		{"unpaid", domain.SubscriptionStatusUnpaid, t0, false},
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
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), plus, snap(domain.SubscriptionStatusActive, t0, end), t0, "u")
	assert.Equal(t, []string{"created", "activated", "state_changed"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.SubscriptionStatusActive, end, end.AddDate(0, 1, 0)), plus, end, "u")
	assert.Equal(t, []string{"renewed"}, eventNames(s.PopEvents()))

	s.ApplySnapshot(snap(domain.SubscriptionStatusPastDue, end, end.AddDate(0, 1, 0)), pro, end, "u")
	events := s.PopEvents()
	assert.Equal(t, []string{"past_due", "plan_changed", "state_changed"}, eventNames(events))
	changed, ok := events[2].(*domain.SubscriptionStateChangedEvent)
	require.True(t, ok)
	assert.Equal(t, s.WorkspaceID(), changed.WorkspaceID)
	assert.Equal(t, domain.PlanTierPro, changed.Plan)
	assert.Equal(t, domain.SubscriptionStatusPastDue, changed.Status)

	ended := snap(domain.SubscriptionStatusCanceled, end, end.AddDate(0, 1, 0))
	s.ApplySnapshot(ended, pro, end, "u")
	assert.Equal(t, []string{"ended", "state_changed"}, eventNames(s.PopEvents()))
}

func TestCancelNowRecordsStateChangeAndRefunds(t *testing.T) {
	end := t0.Add(30 * 24 * time.Hour)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.SubscriptionStatusActive, t0, end), t0, "u")
	s.PopEvents()
	line, err := domain.NewRefundLine(domain.RefundLineParams{PaymentRef: "pi_1", PaidAmount: vnd(t, 30000), Amount: vnd(t, 10000)})
	require.NoError(t, err)
	quote, err := domain.NewRefundQuote([]domain.RefundLine{line}, 0, 0)
	require.NoError(t, err)

	refunds, err := s.CancelNow(t0, domain.RefundReasonCancelNow, quote, "u")
	require.NoError(t, err)

	events := s.PopEvents()
	assert.Equal(t, []string{"canceled_now", "state_changed", "refund_requested"}, eventNames(events))
	requested, ok := events[2].(*domain.RefundRequestedEvent)
	require.True(t, ok)
	assert.Equal(t, refunds[0].ID(), requested.RefundID)
	assert.Equal(t, s.ID().String(), requested.GetAggregateID())
}

func TestQuoteCancelNow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	end := t0.Add(30 * 24 * time.Hour)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.SubscriptionStatusActive, t0, end), t0, "u")
	pay := payment(t, 30000, "pi_1", t0, t0, end)

	q, err := s.QuoteCancelNow(t0.Add(10*24*time.Hour), policy, []domain.PaymentInfo{pay})
	require.NoError(t, err)
	assert.EqualValues(t, 20000, q.Amount().Amount())

	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	// A second refund against the same payment must not exceed what was paid.
	overLine, err := domain.NewRefundLine(domain.RefundLineParams{PaymentRef: "pi_1", PaidAmount: pay.Amount(), Amount: vnd(t, 20000)})
	require.NoError(t, err)
	over, err := domain.NewRefundQuote([]domain.RefundLine{overLine}, 0, 0)
	require.NoError(t, err)
	_, err = s.CancelNow(t0.Add(10*24*time.Hour), domain.RefundReasonCancelNow, over, "u")
	assert.Error(t, err, "non-current subscription / over-refund")
}

// Plus 30000 paid for day 0-30; upgrade on day 10 invoiced 40000 for day 10-30; cancel on day 20.
func TestQuoteCancelNowAfterUpgrade(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	day := func(n int) time.Time { return t0.Add(time.Duration(n) * 24 * time.Hour) }
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalMonth, 90000), snap(domain.SubscriptionStatusActive, t0, day(30)), t0, "u")
	payments := []domain.PaymentInfo{
		payment(t, 30000, "pi_a", t0, t0, day(30)),
		payment(t, 40000, "pi_b", day(10), day(10), day(30)),
	}

	q, err := s.QuoteCancelNow(day(20), policy, payments)
	require.NoError(t, err)
	lines := q.Lines()
	require.Len(t, lines, 2)
	assert.EqualValues(t, 10000, lines[0].Amount().Amount())
	assert.EqualValues(t, 20000, lines[1].Amount().Amount())
	assert.EqualValues(t, 30000, q.Amount().Amount())

	refunds, err := s.CancelNow(day(20), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	require.Len(t, refunds, 2)
	assert.Equal(t, "pi_b", refunds[1].PaymentRef())
}

func TestLifetimeRefundWindow(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 14 days
	s, _ := domain.NewLifetimeSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalLifetime, 2490000), "pi_l", t0, t0, "u")
	pay := []domain.PaymentInfo{payment(t, 2490000, "pi_l", t0, time.Time{}, time.Time{})}

	q, err := s.QuoteCancelNow(t0.Add(13*24*time.Hour), policy, pay)
	require.NoError(t, err)
	assert.EqualValues(t, 2490000, q.Amount().Amount())

	_, err = s.QuoteCancelNow(t0.Add(15*24*time.Hour), policy, pay)
	require.Error(t, err, "refundWindowExpired after 14 days")
	assert.True(t, s.GrantsAccess(t0.AddDate(5, 0, 0), policy), "lifetime grants access forever")
}

func TestQuoteCancelNowWithoutPayments(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	end := t0.AddDate(0, 1, 0)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.SubscriptionStatusPastDue, t0, end), t0, "u")

	q, err := s.QuoteCancelNow(t0.Add(time.Hour), policy, nil)
	require.NoError(t, err)
	assert.Zero(t, q.Amount().Amount())
	assert.Empty(t, q.Lines())

	refunds, err := s.CancelNow(t0.Add(time.Hour), domain.RefundReasonCancelNow, q, "u")
	require.NoError(t, err)
	assert.Empty(t, refunds)
	assert.Equal(t, domain.SubscriptionStatusCanceled, s.Status())
}

func TestPastDueGraceStartsWhenSubscriptionBecomesPastDue(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0) // 3 days grace
	plus := price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000)
	end := t0.AddDate(0, 1, 0)
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), plus, snap(domain.SubscriptionStatusActive, t0, end), t0, "u")

	s.ApplySnapshot(snap(domain.SubscriptionStatusPastDue, end, end.AddDate(0, 1, 0)), plus, end, "u")

	assert.True(t, s.GrantsAccess(end.Add(48*time.Hour), policy))
	assert.False(t, s.GrantsAccess(end.Add(96*time.Hour), policy))
	require.NotNil(t, s.AccessUntil(policy))
	assert.True(t, s.AccessUntil(policy).Equal(end.Add(72*time.Hour)))

	s.ApplySnapshot(snap(domain.SubscriptionStatusPastDue, end, end.AddDate(0, 1, 0)), plus, end.Add(24*time.Hour), "u")
	assert.True(t, s.AccessUntil(policy).Equal(end.Add(72*time.Hour)), "a repeated past_due snapshot keeps the original start")

	recovered := end.Add(48 * time.Hour)
	s.ApplySnapshot(snap(domain.SubscriptionStatusActive, end, end.AddDate(0, 1, 0)), plus, recovered, "u")
	assert.Nil(t, s.PastDueSince())
	assert.True(t, s.AccessUntil(policy).Equal(end.AddDate(0, 1, 0)))
}

func TestPastDueWithoutRecordedStartFallsBackToPeriodStart(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	end := t0.AddDate(0, 1, 0)
	s := domain.UnmarshalSubscription(domain.UnmarshalSubscriptionParams{
		ID: uuid.New(), WorkspaceID: uuid.New(), Plan: domain.PlanTierPlus, Interval: domain.IntervalMonth, PriceID: uuid.New(),
		Status: domain.SubscriptionStatusPastDue, CurrentPeriodStart: &t0, CurrentPeriodEnd: &end,
		ProviderSyncedAt: t0, CreatedAt: t0, UpdatedAt: t0,
	})

	assert.True(t, s.AccessUntil(policy).Equal(t0.Add(72*time.Hour)))
	assert.True(t, s.GrantsAccess(t0.Add(48*time.Hour), policy))
	assert.False(t, s.GrantsAccess(t0.Add(96*time.Hour), policy))

	s.ApplySnapshot(snap(domain.SubscriptionStatusPastDue, t0, end), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), t0.Add(24*time.Hour), "u")
	require.NotNil(t, s.PastDueSince())
	assert.True(t, s.PastDueSince().Equal(t0))
}

// Plus monthly 30000 paid day 0-30; on day 15 the portal moves to yearly (300000): Stripe resets the
// period to [day 15, +365d] and credits the unused half month, so the yearly invoice is 285000.
func TestQuoteCancelNowAfterMonthToYearUpgrade(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	day := func(n int) time.Time { return t0.Add(time.Duration(n) * 24 * time.Hour) }
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPro, domain.IntervalYear, 300000), snap(domain.SubscriptionStatusActive, day(15), day(15+365)), day(15), "u")
	payments := []domain.PaymentInfo{
		payment(t, 30000, "pi_m", t0, t0, day(30)),
		payment(t, 285000, "pi_y", day(15), day(15), day(15+365)),
	}

	q, err := s.QuoteCancelNow(day(15), policy, payments)

	require.NoError(t, err)
	lines := q.Lines()
	require.Len(t, lines, 2)
	assert.EqualValues(t, 15000, lines[0].Amount().Amount())
	assert.EqualValues(t, 285000, lines[1].Amount().Amount())
	assert.EqualValues(t, 300000, q.Amount().Amount())
}

func TestQuoteCancelNowIgnoresPaymentOfPreviousPeriod(t *testing.T) {
	policy := domain.DefaultBillingPolicy(t0)
	day := func(n int) time.Time { return t0.Add(time.Duration(n) * 24 * time.Hour) }
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 30000), snap(domain.SubscriptionStatusActive, t0, day(30)), t0, "u")
	payments := []domain.PaymentInfo{
		payment(t, 30000, "pi_prev", day(-30), day(-30), t0),
		payment(t, 30000, "pi_cur", t0, t0, day(30)),
	}

	q, err := s.QuoteCancelNow(day(5), policy, payments)

	require.NoError(t, err)
	assert.EqualValues(t, 25000, q.Amount().Amount())
	for _, l := range q.Lines() {
		if l.PaymentRef() == "pi_prev" {
			assert.Zero(t, l.Amount().Amount())
		}
	}
}

func TestRecordExternalRefundRecordsOnlyItsSettlement(t *testing.T) {
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.SubscriptionStatusActive, t0, t0.AddDate(0, 1, 0)), t0, "u")
	s.PopEvents()

	require.NoError(t, s.RecordExternalRefund("re_ext", "pi_1", vnd(t, 5000), domain.RefundStatusSucceeded, nil, t0, "u"))

	assert.Equal(t, []string{"refund_succeeded"}, eventNames(s.PopEvents()))
}

func TestUpdateRefundNeverReopensSettledRefund(t *testing.T) {
	s, _ := domain.NewSubscription(uuid.New(), uuid.New(), price(t, domain.PlanTierPlus, domain.IntervalMonth, 49000), snap(domain.SubscriptionStatusActive, t0, t0.AddDate(0, 1, 0)), t0, "u")
	require.NoError(t, s.RecordExternalRefund("re_ext", "pi_1", vnd(t, 5000), domain.RefundStatusSucceeded, nil, t0, "u"))
	rf := s.RefundByProviderID("re_ext")

	require.NoError(t, s.UpdateRefund(rf.ID(), "re_ext", domain.RefundStatusPending, nil, t0, "u"))

	assert.Equal(t, domain.RefundStatusSucceeded, rf.Status())
}
