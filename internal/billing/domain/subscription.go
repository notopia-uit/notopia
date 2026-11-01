package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Subscription struct {
	eventRecorder
	id                     uuid.UUID
	workspaceID            uuid.UUID
	plan                   PlanTier
	interval               Interval
	priceID                uuid.UUID
	status                 SubscriptionStatus
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

func (s *Subscription) recordStateChanged(userID string) {
	s.record(&SubscriptionStateChangedEvent{BaseEvent: s.base(userID), WorkspaceID: s.workspaceID, Plan: s.plan, Status: s.status})
}

func NewSubscription(id, workspaceID uuid.UUID, price *Price, snap ProviderSnapshot, now time.Time, userID string) (*Subscription, error) {
	if price.Interval() == IntervalLifetime {
		return nil, errs.NewInvalid("recurring subscription needs a recurring price")
	}
	providerID := snap.providerSubscriptionID
	s := &Subscription{
		eventRecorder:          eventRecorder{events: nil},
		id:                     id,
		workspaceID:            workspaceID,
		plan:                   price.Plan(),
		interval:               price.Interval(),
		priceID:                price.ID(),
		status:                 SubscriptionStatusIncomplete,
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
		status:                 SubscriptionStatusActive,
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
	s.recordStateChanged(userID)
	return s, nil
}

// NewDuplicateSubscription records a checkout that duplicated the workspace's current subscription:
// canceled at once with every payment refunded in full.
func NewDuplicateSubscription(
	id, workspaceID uuid.UUID,
	price *Price,
	snap ProviderSnapshot,
	payments []PaymentInfo,
	now time.Time,
	userID string,
) (*Subscription, error) {
	s, err := NewSubscription(id, workspaceID, price, snap, now, userID)
	if err != nil {
		return nil, err
	}
	if err := s.cancelDuplicate(payments, now, userID); err != nil {
		return nil, err
	}
	return s, nil
}

// NewDuplicateLifetimeSubscription is NewDuplicateSubscription for a lifetime checkout.
func NewDuplicateLifetimeSubscription(id, workspaceID uuid.UUID, price *Price, payment PaymentInfo, now time.Time, userID string) (*Subscription, error) {
	s, err := NewLifetimeSubscription(id, workspaceID, price, payment.paymentRef, payment.paidAt, now, userID)
	if err != nil {
		return nil, err
	}
	if err := s.cancelDuplicate([]PaymentInfo{payment}, now, userID); err != nil {
		return nil, err
	}
	return s, nil
}

// cancelDuplicate drops SubscriptionStateChangedEvent: a duplicate never was the workspace's subscription,
// so reporting it would override the workspace's current one downstream.
func (s *Subscription) cancelDuplicate(payments []PaymentInfo, now time.Time, userID string) error {
	quote, err := FullRefundQuote(payments...)
	if err != nil {
		return err
	}
	if _, err := s.CancelNow(now, RefundReasonDuplicateCheckout, quote, userID); err != nil {
		return err
	}
	s.events = slices.DeleteFunc(s.events, func(e Event) bool {
		_, ok := e.(*SubscriptionStateChangedEvent)
		return ok
	})
	return nil
}

type UnmarshalSubscriptionParams struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	Plan                   PlanTier
	Interval               Interval
	PriceID                uuid.UUID
	Status                 SubscriptionStatus
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
	isNew := prevStatus == SubscriptionStatusIncomplete && prevEnd == nil

	start, end := snap.currentPeriodStart, snap.currentPeriodEnd
	s.status = snap.status
	s.plan, s.interval, s.priceID = price.Plan(), price.Interval(), price.ID()
	s.currentPeriodStart, s.currentPeriodEnd = &start, &end
	s.cancelAtPeriodEnd = snap.cancelAtPeriodEnd
	s.canceledAt = snap.CanceledAt()
	s.providerSyncedAt, s.updatedAt = now, now
	s.trackPastDue(prevStatus, prevStart, now)

	switch {
	case s.status == SubscriptionStatusActive && prevStatus != SubscriptionStatusActive:
		s.paidAt = &now
		s.record(&SubscriptionActivatedEvent{BaseEvent: s.base(userID)})
	case s.status == SubscriptionStatusActive && prevEnd != nil && end.After(*prevEnd):
		s.paidAt = &now
		s.record(&SubscriptionRenewedEvent{BaseEvent: s.base(userID)})
	}
	if s.status == SubscriptionStatusPastDue && prevStatus != SubscriptionStatusPastDue {
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
	if isNew || prevStatus != s.status || prevPlan != s.plan {
		s.recordStateChanged(userID)
	}
}

func (s *Subscription) trackPastDue(prevStatus SubscriptionStatus, prevStart *time.Time, now time.Time) {
	switch {
	case s.status != SubscriptionStatusPastDue:
		s.pastDueSince = nil
	case prevStatus != SubscriptionStatusPastDue:
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

// AwaitingProviderCancel reports a locally canceled subscription the provider has not confirmed canceled yet.
func (s *Subscription) AwaitingProviderCancel() bool {
	return s.status == SubscriptionStatusCanceled && s.providerSubscriptionID != nil && s.canceledAt != nil &&
		!s.providerSyncedAt.After(*s.canceledAt)
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
			total += r.amount.amount
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
	var lines []RefundLine
	add := func(p PaymentInfo, amount, coverage, unused int64) error {
		amount = max(min(amount, p.amount.amount-s.refundedFor(p.paymentRef)), 0)
		refundAmount, err := NewMoney(amount, p.amount.currency)
		if err != nil {
			return err
		}
		line, err := NewRefundLine(RefundLineParams{
			PaymentRef:      p.paymentRef,
			PaidAmount:      p.amount,
			Amount:          refundAmount,
			CoverageSeconds: coverage,
			UnusedSeconds:   unused,
		})
		if err != nil {
			return err
		}
		lines = append(lines, line)
		return nil
	}
	if s.interval == IntervalLifetime {
		if s.paidAt == nil || now.Sub(*s.paidAt) > policy.LifetimeRefundWindow() {
			return RefundQuote{}, errs.New(errs.CodeRefundWindowExpired, "lifetime refund window expired", nil)
		}
		for _, p := range payments {
			if err := add(p, p.amount.amount, 0, 0); err != nil {
				return RefundQuote{}, err
			}
		}
		return NewRefundQuote(lines, 0, 0)
	}
	if s.currentPeriodStart == nil || s.currentPeriodEnd == nil {
		return RefundQuote{}, errs.NewInternal("subscription has no period", nil)
	}
	periodSeconds := int64(s.currentPeriodEnd.Sub(*s.currentPeriodStart).Seconds())
	unusedSeconds := max(int64(s.currentPeriodEnd.Sub(now).Seconds()), 0)
	for _, p := range payments {
		coverage := int64(p.coverageEnd.Sub(p.coverageStart).Seconds())
		if coverage <= 0 {
			continue
		}
		from := now
		if p.coverageStart.After(now) {
			from = p.coverageStart
		}
		unused := max(int64(p.coverageEnd.Sub(from).Seconds()), 0)
		if err := add(p, p.amount.amount*unused/coverage, coverage, unused); err != nil {
			return RefundQuote{}, err
		}
	}
	return NewRefundQuote(lines, unusedSeconds, periodSeconds)
}

func (s *Subscription) CancelNow(now time.Time, reason RefundReason, quote RefundQuote, userID string) ([]*Refund, error) {
	if !s.IsCurrent() {
		return nil, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	for _, l := range quote.lines {
		if l.amount.amount > 0 && s.refundedFor(l.paymentRef)+l.amount.amount > l.paidAmount.amount {
			return nil, errs.NewInvalid(fmt.Sprintf("refund %d on %s exceeds refundable amount", l.amount.amount, l.paymentRef))
		}
	}
	s.status, s.canceledAt, s.cancelAtPeriodEnd, s.updatedAt = SubscriptionStatusCanceled, &now, false, now
	s.record(&SubscriptionCanceledNowEvent{BaseEvent: s.base(userID), Reason: reason})
	s.recordStateChanged(userID)
	var refunds []*Refund
	for _, l := range quote.lines {
		if l.amount.amount <= 0 {
			continue
		}
		r := &Refund{
			id:               uuid.Must(uuid.NewV7()),
			providerRefundID: nil,
			paymentRef:       l.paymentRef,
			amount:           l.amount,
			reason:           reason,
			status:           RefundStatusPending,
			failureMessage:   nil,
			createdAt:        now,
			updatedAt:        now,
		}
		s.refunds = append(s.refunds, r)
		s.record(&RefundRequestedEvent{BaseEvent: s.base(userID), RefundID: r.id, Amount: l.amount})
		refunds = append(refunds, r)
	}
	return refunds, nil
}

func (s *Subscription) UpdateRefund(refundID uuid.UUID, providerRefundID string, status RefundStatus, failureMessage *string, now time.Time, userID string) error {
	for _, r := range s.refunds {
		if r.id != refundID {
			continue
		}
		// a stale pending (e.g. the CreateRefund response landing after the webhook) never reopens a settled refund
		if r.status.IsSettled() && !status.IsSettled() {
			return nil
		}
		prev := r.status
		r.providerRefundID, r.status, r.failureMessage, r.updatedAt = &providerRefundID, status, failureMessage, now
		if prev != status {
			s.recordRefundSettled(refundID, status, userID)
		}
		return nil
	}
	return errs.NewNotFound(fmt.Sprintf("refund %s not found", refundID))
}

func (s *Subscription) recordRefundSettled(refundID uuid.UUID, status RefundStatus, userID string) {
	switch status {
	case RefundStatusSucceeded:
		s.record(&RefundSucceededEvent{BaseEvent: s.base(userID), RefundID: refundID})
	case RefundStatusFailed, RefundStatusCanceled:
		s.record(&RefundFailedEvent{BaseEvent: s.base(userID), RefundID: refundID})
	case RefundStatusUnspecified, RefundStatusPending:
	}
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

func (s *Subscription) Refund(refundID uuid.UUID) *Refund {
	for _, r := range s.refunds {
		if r.id == refundID {
			return r
		}
	}
	return nil
}

func (s *Subscription) RefundByProviderID(providerRefundID string) *Refund {
	for _, r := range s.refunds {
		if r.providerRefundID != nil && *r.providerRefundID == providerRefundID {
			return r
		}
	}
	return nil
}

// RecordExternalRefund tracks a refund made outside Notopia (e.g. the provider dashboard). It is already
// submitted, so it records no RefundRequestedEvent (that event drives SubmitRefund), only its settlement.
func (s *Subscription) RecordExternalRefund(
	providerRefundID, paymentRef string,
	amount Money,
	status RefundStatus,
	failureMessage *string,
	now time.Time,
	userID string,
) error {
	if r := s.RefundByProviderID(providerRefundID); r != nil {
		return s.UpdateRefund(r.id, providerRefundID, status, failureMessage, now, userID)
	}
	id := uuid.Must(uuid.NewV7())
	s.refunds = append(s.refunds, &Refund{
		id:               id,
		providerRefundID: &providerRefundID,
		paymentRef:       paymentRef,
		amount:           amount,
		reason:           RefundReasonExternal,
		status:           status,
		failureMessage:   failureMessage,
		createdAt:        now,
		updatedAt:        now,
	})
	s.recordRefundSettled(id, status, userID)
	return nil
}

func (s *Subscription) IsCurrent() bool {
	return s.status != SubscriptionStatusCanceled && s.status != SubscriptionStatusIncompleteExpired
}

func (s *Subscription) GrantsAccess(now time.Time, policy *BillingPolicy) bool {
	switch s.status {
	case SubscriptionStatusActive:
		return true
	case SubscriptionStatusPastDue:
		since := s.pastDueStart()
		return since != nil && now.Before(since.Add(policy.PastDueGrace()))
	case SubscriptionStatusNone, SubscriptionStatusIncomplete, SubscriptionStatusUnpaid, SubscriptionStatusCanceled, SubscriptionStatusIncompleteExpired:
		return false
	default:
		return false
	}
}

func (s *Subscription) AccessUntil(policy *BillingPolicy) *time.Time {
	switch {
	case s.interval == IntervalLifetime || s.currentPeriodEnd == nil:
		return nil
	case s.status == SubscriptionStatusPastDue:
		since := s.pastDueStart()
		if since == nil {
			return nil
		}
		t := since.Add(policy.PastDueGrace())
		return &t
	case s.status == SubscriptionStatusActive:
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
func (s *Subscription) Status() SubscriptionStatus      { return s.status }
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
