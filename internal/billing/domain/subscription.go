package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type ProviderSnapshot struct {
	ProviderSubscriptionID string
	ProviderPriceID        string
	WorkspaceID            uuid.UUID
	Status                 Status
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
}

// PaymentInfo is one paid invoice payment. CoverageStart/End is the service window it paid
// for (invoice line periods); zero for lifetime.
type PaymentInfo struct {
	Amount        Money
	PaymentRef    string
	PaidAt        time.Time
	CoverageStart time.Time
	CoverageEnd   time.Time
}

type RefundLine struct {
	PaymentRef      string
	PaidAmount      Money
	Amount          Money
	CoverageSeconds int64
	UnusedSeconds   int64
}

// RefundQuote.Amount is the sum of Lines; each line becomes one Refund.
type RefundQuote struct {
	Amount        Money
	UnusedSeconds int64
	PeriodSeconds int64
	Lines         []RefundLine
}

// FullRefundQuote refunds every payment in full (duplicate checkout).
func FullRefundQuote(payments ...PaymentInfo) RefundQuote {
	var q RefundQuote
	for _, p := range payments {
		q.Lines = append(q.Lines, RefundLine{
			PaymentRef:      p.PaymentRef,
			PaidAmount:      p.Amount,
			Amount:          p.Amount,
			CoverageSeconds: 0,
			UnusedSeconds:   0,
		})
		q.Amount = Money{Amount: q.Amount.Amount + p.Amount.Amount, Currency: p.Amount.Currency}
	}
	return q
}

type Subscription struct {
	eventRecorder
	id                     uuid.UUID
	workspaceID            uuid.UUID
	plan                   PlanTier
	interval               Interval
	priceID                uuid.UUID
	status                 Status
	providerSubscriptionID *string
	providerPaymentID      *string
	currentPeriodStart     *time.Time
	currentPeriodEnd       *time.Time
	cancelAtPeriodEnd      bool
	canceledAt             *time.Time
	pastDueSince           *time.Time
	paidAt                 *time.Time
	providerSyncedAt       time.Time
	createdAt              time.Time
	updatedAt              time.Time
	refunds                []*Refund
}

func (s *Subscription) base(userID string) BaseEvent { return NewBaseEvent(s.id.String(), userID) }

func NewSubscription(id, workspaceID uuid.UUID, price *Price, snap ProviderSnapshot, now time.Time, userID string) (*Subscription, error) {
	if price.Interval() == IntervalLifetime {
		return nil, errs.NewInvalid("recurring subscription needs a recurring price")
	}
	providerID := snap.ProviderSubscriptionID
	s := &Subscription{
		eventRecorder:          eventRecorder{events: nil},
		id:                     id,
		workspaceID:            workspaceID,
		plan:                   price.Plan(),
		interval:               price.Interval(),
		priceID:                price.ID(),
		status:                 StatusIncomplete,
		providerSubscriptionID: &providerID,
		providerPaymentID:      nil,
		currentPeriodStart:     nil,
		currentPeriodEnd:       nil,
		cancelAtPeriodEnd:      false,
		canceledAt:             nil,
		pastDueSince:           nil,
		paidAt:                 nil,
		providerSyncedAt:       now,
		createdAt:              now,
		updatedAt:              now,
		refunds:                nil,
	}
	s.record(&SubscriptionCreatedEvent{BaseEvent: s.base(userID), Plan: s.plan, Interval: s.interval})
	s.ApplySnapshot(snap, price, now, userID)
	return s, nil
}

func NewLifetimeSubscription(id, workspaceID uuid.UUID, price *Price, providerPaymentID string, paidAt, now time.Time, userID string) (*Subscription, error) {
	if price.Interval() != IntervalLifetime {
		return nil, errs.NewInvalid("lifetime subscription needs a lifetime price")
	}
	s := &Subscription{
		eventRecorder:          eventRecorder{events: nil},
		id:                     id,
		workspaceID:            workspaceID,
		plan:                   price.Plan(),
		interval:               IntervalLifetime,
		priceID:                price.ID(),
		status:                 StatusActive,
		providerSubscriptionID: nil,
		providerPaymentID:      &providerPaymentID,
		currentPeriodStart:     nil,
		currentPeriodEnd:       nil,
		cancelAtPeriodEnd:      false,
		canceledAt:             nil,
		pastDueSince:           nil,
		paidAt:                 &paidAt,
		providerSyncedAt:       now,
		createdAt:              now,
		updatedAt:              now,
		refunds:                nil,
	}
	s.record(&SubscriptionCreatedEvent{BaseEvent: s.base(userID), Plan: s.plan, Interval: s.interval})
	s.record(&SubscriptionActivatedEvent{BaseEvent: s.base(userID)})
	return s, nil
}

type UnmarshalSubscriptionParams struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	Plan                   PlanTier
	Interval               Interval
	PriceID                uuid.UUID
	Status                 Status
	ProviderSubscriptionID *string
	ProviderPaymentID      *string
	CurrentPeriodStart     *time.Time
	CurrentPeriodEnd       *time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
	PastDueSince           *time.Time
	PaidAt                 *time.Time
	ProviderSyncedAt       time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Refunds                []*Refund
}

func UnmarshalSubscription(p UnmarshalSubscriptionParams) *Subscription {
	return &Subscription{
		eventRecorder:          eventRecorder{events: nil},
		id:                     p.ID,
		workspaceID:            p.WorkspaceID,
		plan:                   p.Plan,
		interval:               p.Interval,
		priceID:                p.PriceID,
		status:                 p.Status,
		providerSubscriptionID: p.ProviderSubscriptionID,
		providerPaymentID:      p.ProviderPaymentID,
		currentPeriodStart:     p.CurrentPeriodStart,
		currentPeriodEnd:       p.CurrentPeriodEnd,
		cancelAtPeriodEnd:      p.CancelAtPeriodEnd,
		canceledAt:             p.CanceledAt,
		pastDueSince:           p.PastDueSince,
		paidAt:                 p.PaidAt,
		providerSyncedAt:       p.ProviderSyncedAt,
		createdAt:              p.CreatedAt,
		updatedAt:              p.UpdatedAt,
		refunds:                p.Refunds,
	}
}

func (s *Subscription) ApplySnapshot(snap ProviderSnapshot, price *Price, now time.Time, userID string) {
	prevStatus, prevPlan, prevCancel := s.status, s.plan, s.cancelAtPeriodEnd
	prevEnd, prevStart := s.currentPeriodEnd, s.currentPeriodStart
	isNew := prevStatus == StatusIncomplete && prevEnd == nil

	start, end := snap.CurrentPeriodStart, snap.CurrentPeriodEnd
	s.status = snap.Status
	s.plan, s.interval, s.priceID = price.Plan(), price.Interval(), price.ID()
	s.currentPeriodStart, s.currentPeriodEnd = &start, &end
	s.cancelAtPeriodEnd = snap.CancelAtPeriodEnd
	s.canceledAt = snap.CanceledAt
	s.providerSyncedAt, s.updatedAt = now, now
	s.trackPastDue(prevStatus, prevStart, now)

	switch {
	case s.status == StatusActive && prevStatus != StatusActive:
		s.paidAt = &now
		s.record(&SubscriptionActivatedEvent{BaseEvent: s.base(userID)})
	case s.status == StatusActive && prevEnd != nil && end.After(*prevEnd):
		s.paidAt = &now
		s.record(&SubscriptionRenewedEvent{BaseEvent: s.base(userID)})
	}
	if s.status == StatusPastDue && prevStatus != StatusPastDue {
		s.record(&SubscriptionPastDueEvent{BaseEvent: s.base(userID)})
	}
	if !isNew && prevPlan != s.plan {
		s.record(&SubscriptionPlanChangedEvent{BaseEvent: s.base(userID), From: prevPlan, To: s.plan})
	}
	if !prevCancel && s.cancelAtPeriodEnd {
		s.record(&SubscriptionCancelScheduledEvent{BaseEvent: s.base(userID)})
	}
	if prevCancel && !s.cancelAtPeriodEnd && s.IsCurrent() {
		s.record(&SubscriptionCancelResumedEvent{BaseEvent: s.base(userID)})
	}
	if !s.IsCurrent() && (prevStatus != s.status) && !isNew {
		s.record(&SubscriptionEndedEvent{BaseEvent: s.base(userID)})
	}
}

func (s *Subscription) trackPastDue(prevStatus Status, prevStart *time.Time, now time.Time) {
	switch {
	case s.status != StatusPastDue:
		s.pastDueSince = nil
	case prevStatus != StatusPastDue:
		s.pastDueSince = &now
	case s.pastDueSince == nil:
		since := now
		if prevStart != nil {
			since = *prevStart
		}
		s.pastDueSince = &since
	}
}

// pastDueStart is when the grace window began. Rows saved before past_due_since existed fall back
// to the period start, which never cuts paid time.
func (s *Subscription) pastDueStart() *time.Time {
	if s.pastDueSince != nil {
		return s.pastDueSince
	}
	return s.currentPeriodStart
}

// MarkProviderSynced records that the provider confirmed this (locally canceled) subscription ended.
func (s *Subscription) MarkProviderSynced(now time.Time) {
	s.providerSyncedAt, s.updatedAt = now, now
}

func (s *Subscription) ScheduleCancel(now time.Time, userID string) error {
	if s.interval == IntervalLifetime {
		return errs.NewInvalid("lifetime subscriptions cannot be scheduled for cancellation")
	}
	if s.cancelAtPeriodEnd {
		return nil
	}
	s.cancelAtPeriodEnd, s.updatedAt = true, now
	s.record(&SubscriptionCancelScheduledEvent{BaseEvent: s.base(userID)})
	return nil
}

func (s *Subscription) ResumeCancel(now time.Time, userID string) {
	if !s.cancelAtPeriodEnd {
		return
	}
	s.cancelAtPeriodEnd, s.updatedAt = false, now
	s.record(&SubscriptionCancelResumedEvent{BaseEvent: s.base(userID)})
}

func (s *Subscription) refundedFor(paymentRef string) int64 {
	var total int64
	for _, r := range s.refunds {
		if r.paymentRef == paymentRef && r.status != RefundStatusFailed && r.status != RefundStatusCanceled {
			total += r.amount.Amount
		}
	}
	return total
}

// QuoteCancelNow quotes each paid payment on its own coverage window, not clipped to the current period:
// an immediately invoiced upgrade does not hide the earlier payment, and a month-to-year change
// (which resets the period and credits the unused month into the yearly invoice) still refunds the
// original monthly payment's unused part.
func (s *Subscription) QuoteCancelNow(now time.Time, policy *BillingPolicy, payments []PaymentInfo) (RefundQuote, error) {
	if !s.IsCurrent() {
		return RefundQuote{}, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	var q RefundQuote
	add := func(p PaymentInfo, amount, coverage, unused int64) {
		amount = max(min(amount, p.Amount.Amount-s.refundedFor(p.PaymentRef)), 0)
		q.Lines = append(q.Lines, RefundLine{
			PaymentRef:      p.PaymentRef,
			PaidAmount:      p.Amount,
			Amount:          Money{Amount: amount, Currency: p.Amount.Currency},
			CoverageSeconds: coverage,
			UnusedSeconds:   unused,
		})
		q.Amount = Money{Amount: q.Amount.Amount + amount, Currency: p.Amount.Currency}
	}
	if s.interval == IntervalLifetime {
		if s.paidAt == nil || now.Sub(*s.paidAt) > policy.LifetimeRefundWindow() {
			return RefundQuote{}, errs.New(errs.CodeRefundWindowExpired, "lifetime refund window expired", nil)
		}
		for _, p := range payments {
			add(p, p.Amount.Amount, 0, 0)
		}
		return q, nil
	}
	if s.currentPeriodStart == nil || s.currentPeriodEnd == nil {
		return RefundQuote{}, errs.NewInternal("subscription has no period", nil)
	}
	q.PeriodSeconds = int64(s.currentPeriodEnd.Sub(*s.currentPeriodStart).Seconds())
	q.UnusedSeconds = max(int64(s.currentPeriodEnd.Sub(now).Seconds()), 0)
	for _, p := range payments {
		coverage := int64(p.CoverageEnd.Sub(p.CoverageStart).Seconds())
		if coverage <= 0 {
			continue
		}
		unused := max(int64(p.CoverageEnd.Sub(latest(now, p.CoverageStart)).Seconds()), 0)
		add(p, p.Amount.Amount*unused/coverage, coverage, unused)
	}
	return q, nil
}

func latest(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// CancelNow adds one pending Refund per quote line with a positive amount.
func (s *Subscription) CancelNow(now time.Time, reason RefundReason, quote RefundQuote, userID string) ([]*Refund, error) {
	if !s.IsCurrent() {
		return nil, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	for _, l := range quote.Lines {
		if l.Amount.Amount > 0 && s.refundedFor(l.PaymentRef)+l.Amount.Amount > l.PaidAmount.Amount {
			return nil, errs.NewInvalid(fmt.Sprintf("refund %d on %s exceeds refundable amount", l.Amount.Amount, l.PaymentRef))
		}
	}
	s.status, s.canceledAt, s.cancelAtPeriodEnd, s.updatedAt = StatusCanceled, &now, false, now
	s.record(&SubscriptionCanceledNowEvent{BaseEvent: s.base(userID), Reason: reason})
	var refunds []*Refund
	for _, l := range quote.Lines {
		if l.Amount.Amount <= 0 {
			continue
		}
		r := &Refund{
			id:               uuid.Must(uuid.NewV7()),
			providerRefundID: nil,
			paymentRef:       l.PaymentRef,
			amount:           l.Amount,
			reason:           reason,
			status:           RefundStatusPending,
			failureMessage:   nil,
			createdAt:        now,
			updatedAt:        now,
		}
		s.refunds = append(s.refunds, r)
		s.record(&RefundRequestedEvent{BaseEvent: s.base(userID), RefundID: r.id, Amount: l.Amount})
		refunds = append(refunds, r)
	}
	return refunds, nil
}

func (s *Subscription) UpdateRefund(refundID uuid.UUID, providerRefundID string, status RefundStatus, failureMessage *string, now time.Time, userID string) error {
	for _, r := range s.refunds {
		if r.id != refundID {
			continue
		}
		prev := r.status
		r.providerRefundID, r.status, r.failureMessage, r.updatedAt = &providerRefundID, status, failureMessage, now
		if prev != status {
			switch status {
			case RefundStatusSucceeded:
				s.record(&RefundSucceededEvent{BaseEvent: s.base(userID), RefundID: refundID})
			case RefundStatusFailed, RefundStatusCanceled:
				s.record(&RefundFailedEvent{BaseEvent: s.base(userID), RefundID: refundID})
			case RefundStatusUnspecified, RefundStatusPending:
			}
		}
		return nil
	}
	return errs.NewNotFound(fmt.Sprintf("refund %s not found", refundID))
}

// FailRefund marks a refund the provider will never accept; it keeps no provider refund id.
func (s *Subscription) FailRefund(refundID uuid.UUID, failureMessage string, now time.Time, userID string) error {
	for _, r := range s.refunds {
		if r.id != refundID {
			continue
		}
		if r.status == RefundStatusFailed {
			return nil
		}
		r.status, r.failureMessage, r.updatedAt = RefundStatusFailed, &failureMessage, now
		s.record(&RefundFailedEvent{BaseEvent: s.base(userID), RefundID: refundID})
		return nil
	}
	return errs.NewNotFound(fmt.Sprintf("refund %s not found", refundID))
}

func (s *Subscription) RefundByProviderID(providerRefundID string) *Refund {
	for _, r := range s.refunds {
		if r.providerRefundID != nil && *r.providerRefundID == providerRefundID {
			return r
		}
	}
	return nil
}

func (s *Subscription) RecordExternalRefund(providerRefundID, paymentRef string, amount Money, status RefundStatus, now time.Time, userID string) {
	if r := s.RefundByProviderID(providerRefundID); r != nil {
		_ = s.UpdateRefund(r.id, providerRefundID, status, nil, now, userID)
		return
	}
	id := uuid.Must(uuid.NewV7())
	s.refunds = append(s.refunds, &Refund{
		id:               id,
		providerRefundID: &providerRefundID,
		paymentRef:       paymentRef,
		amount:           amount,
		reason:           RefundReasonExternal,
		status:           status,
		failureMessage:   nil,
		createdAt:        now,
		updatedAt:        now,
	})
	s.record(&RefundRequestedEvent{BaseEvent: s.base(userID), RefundID: id, Amount: amount})
}

func (s *Subscription) IsCurrent() bool {
	return s.status != StatusCanceled && s.status != StatusIncompleteExpired
}

func (s *Subscription) GrantsAccess(now time.Time, policy *BillingPolicy) bool {
	switch s.status {
	case StatusActive:
		return true
	case StatusPastDue:
		since := s.pastDueStart()
		return since != nil && now.Before(since.Add(policy.PastDueGrace()))
	case StatusNone, StatusIncomplete, StatusUnpaid, StatusCanceled, StatusIncompleteExpired:
		return false
	default:
		return false
	}
}

func (s *Subscription) AccessUntil(policy *BillingPolicy) *time.Time {
	switch {
	case s.interval == IntervalLifetime || s.currentPeriodEnd == nil:
		return nil
	case s.status == StatusPastDue:
		since := s.pastDueStart()
		if since == nil {
			return nil
		}
		t := since.Add(policy.PastDueGrace())
		return &t
	case s.status == StatusActive:
		return s.currentPeriodEnd
	default:
		return nil
	}
}

func (s *Subscription) ID() uuid.UUID                   { return s.id }
func (s *Subscription) WorkspaceID() uuid.UUID          { return s.workspaceID }
func (s *Subscription) Plan() PlanTier                  { return s.plan }
func (s *Subscription) Interval() Interval              { return s.interval }
func (s *Subscription) PriceID() uuid.UUID              { return s.priceID }
func (s *Subscription) Status() Status                  { return s.status }
func (s *Subscription) ProviderSubscriptionID() *string { return s.providerSubscriptionID }
func (s *Subscription) ProviderPaymentID() *string      { return s.providerPaymentID }
func (s *Subscription) CurrentPeriodStart() *time.Time  { return s.currentPeriodStart }
func (s *Subscription) CurrentPeriodEnd() *time.Time    { return s.currentPeriodEnd }
func (s *Subscription) CancelAtPeriodEnd() bool         { return s.cancelAtPeriodEnd }
func (s *Subscription) CanceledAt() *time.Time          { return s.canceledAt }
func (s *Subscription) PastDueSince() *time.Time        { return s.pastDueSince }
func (s *Subscription) PaidAt() *time.Time              { return s.paidAt }
func (s *Subscription) ProviderSyncedAt() time.Time     { return s.providerSyncedAt }
func (s *Subscription) CreatedAt() time.Time            { return s.createdAt }
func (s *Subscription) UpdatedAt() time.Time            { return s.updatedAt }
func (s *Subscription) Refunds() []*Refund              { return s.refunds }
