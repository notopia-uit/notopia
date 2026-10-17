package app

import (
	"time"

	"github.com/google/uuid"
)

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

func (t PlanTier) IsValid() bool { return t >= PlanTierFree && t <= PlanTierPro }

func (t PlanTier) String() string {
	switch t {
	case PlanTierFree:
		return "free"
	case PlanTierPlus:
		return "plus"
	case PlanTierPro:
		return "pro"
	case PlanTierUnspecified:
	}
	return "unspecified"
}

type BillingInterval uint8

const (
	BillingIntervalUnspecified BillingInterval = iota
	BillingIntervalMonth
	BillingIntervalYear
	BillingIntervalLifetime
)

func (i BillingInterval) IsValid() bool {
	return i >= BillingIntervalMonth && i <= BillingIntervalLifetime
}

func (i BillingInterval) String() string {
	switch i {
	case BillingIntervalMonth:
		return "month"
	case BillingIntervalYear:
		return "year"
	case BillingIntervalLifetime:
		return "lifetime"
	case BillingIntervalUnspecified:
	}
	return "unspecified"
}

// SubscriptionStatusNone is the zero value and a real state: the workspace has no subscription.
type SubscriptionStatus uint8

const (
	SubscriptionStatusNone SubscriptionStatus = iota
	SubscriptionStatusIncomplete
	SubscriptionStatusActive
	SubscriptionStatusPastDue
	SubscriptionStatusUnpaid
	SubscriptionStatusCanceled
	SubscriptionStatusIncompleteExpired
)

func (s SubscriptionStatus) IsValid() bool { return s <= SubscriptionStatusIncompleteExpired }

func (s SubscriptionStatus) String() string {
	switch s {
	case SubscriptionStatusNone:
		return "none"
	case SubscriptionStatusIncomplete:
		return "incomplete"
	case SubscriptionStatusActive:
		return "active"
	case SubscriptionStatusPastDue:
		return "past_due"
	case SubscriptionStatusUnpaid:
		return "unpaid"
	case SubscriptionStatusCanceled:
		return "canceled"
	case SubscriptionStatusIncompleteExpired:
		return "incomplete_expired"
	default:
		return "unknown"
	}
}

type RefundReason uint8

const (
	RefundReasonUnspecified RefundReason = iota
	RefundReasonCancelNow
	RefundReasonLifetimeWindow
	RefundReasonReplacedByLifetime
	RefundReasonWorkspaceDeleted
	RefundReasonDuplicateCheckout
	RefundReasonExternal
)

func (r RefundReason) IsValid() bool { return r >= RefundReasonCancelNow && r <= RefundReasonExternal }

func (r RefundReason) String() string {
	switch r {
	case RefundReasonCancelNow:
		return "cancel_now"
	case RefundReasonLifetimeWindow:
		return "lifetime_window"
	case RefundReasonReplacedByLifetime:
		return "replaced_by_lifetime"
	case RefundReasonWorkspaceDeleted:
		return "workspace_deleted"
	case RefundReasonDuplicateCheckout:
		return "duplicate_checkout"
	case RefundReasonExternal:
		return "external"
	case RefundReasonUnspecified:
	}
	return "unspecified"
}

type RefundStatus uint8

const (
	RefundStatusUnspecified RefundStatus = iota
	RefundStatusPending
	RefundStatusSucceeded
	RefundStatusFailed
	RefundStatusCanceled
)

func (s RefundStatus) IsValid() bool { return s >= RefundStatusPending && s <= RefundStatusCanceled }

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusPending:
		return "pending"
	case RefundStatusSucceeded:
		return "succeeded"
	case RefundStatusFailed:
		return "failed"
	case RefundStatusCanceled:
		return "canceled"
	case RefundStatusUnspecified:
	}
	return "unspecified"
}

// Feature and limit keys are plain strings here: both wires (proto map, OpenAPI object) key them by string.
type Plan struct {
	Tier     PlanTier
	Features map[string]bool
	Limits   map[string]int64
}

type Price struct {
	ID         uuid.UUID
	Plan       PlanTier
	Interval   BillingInterval
	Amount     int64
	Currency   string
	LookupKey  string
	Active     bool
	CreatedAt  time.Time
	ArchivedAt *time.Time
}

type PlanView struct {
	Plan   Plan
	Prices []Price
}

type BillingPolicy struct {
	PastDueGraceDays         int
	LifetimeRefundWindowDays int
}

type Entitlements struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
	Status      SubscriptionStatus
	AccessUntil *time.Time
	Features    map[string]bool
	Limits      map[string]int64
}

type Refund struct {
	ID        uuid.UUID
	Amount    int64
	Currency  string
	Reason    RefundReason
	Status    RefundStatus
	CreatedAt time.Time
}

type Subscription struct {
	ID                 uuid.UUID
	Plan               PlanTier
	Interval           BillingInterval
	Status             SubscriptionStatus
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	AccessUntil        *time.Time
	Refunds            []Refund
}

type Payer struct {
	ID           string
	Email        string
	Name         string
	MissingSince *time.Time
}

type WorkspaceBillingView struct {
	Subscription *Subscription
	Payer        *Payer
}

type PayerView struct {
	PayerID       *string
	RequiresPayer bool
}

type RefundQuoteLine struct {
	PaidAmount      int64
	Amount          int64
	CoverageSeconds int64
	UnusedSeconds   int64
}

type RefundQuote struct {
	Amount        int64
	Currency      string
	UnusedSeconds int64
	PeriodSeconds int64
	Lines         []RefundQuoteLine
}
