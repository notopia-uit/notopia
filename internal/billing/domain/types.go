package domain

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

var PlanTiers = []PlanTier{PlanTierFree, PlanTierPlus, PlanTierPro}

func (t PlanTier) IsValid() bool {
	switch t {
	case PlanTierFree, PlanTierPlus, PlanTierPro:
		return true
	case PlanTierUnspecified:
		return false
	}
	return false
}

func (t PlanTier) String() string {
	switch t {
	case PlanTierUnspecified:
		return "unspecified"
	case PlanTierFree:
		return "free"
	case PlanTierPlus:
		return "plus"
	case PlanTierPro:
		return "pro"
	default:
		return "unknown"
	}
}

type Interval uint8

const (
	IntervalUnspecified Interval = iota
	IntervalMonth
	IntervalYear
	IntervalLifetime
)

func (i Interval) IsValid() bool {
	switch i {
	case IntervalMonth, IntervalYear, IntervalLifetime:
		return true
	case IntervalUnspecified:
		return false
	}
	return false
}

func (i Interval) String() string {
	switch i {
	case IntervalUnspecified:
		return "unspecified"
	case IntervalMonth:
		return "month"
	case IntervalYear:
		return "year"
	case IntervalLifetime:
		return "lifetime"
	default:
		return "unknown"
	}
}

// StatusNone is the zero value and a real state: the workspace has no subscription.
type Status uint8

const (
	StatusNone Status = iota
	StatusIncomplete
	StatusActive
	StatusPastDue
	StatusUnpaid
	StatusCanceled
	StatusIncompleteExpired
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNone, StatusIncomplete, StatusActive, StatusPastDue, StatusUnpaid, StatusCanceled, StatusIncompleteExpired:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	switch s {
	case StatusNone:
		return "none"
	case StatusIncomplete:
		return "incomplete"
	case StatusActive:
		return "active"
	case StatusPastDue:
		return "past_due"
	case StatusUnpaid:
		return "unpaid"
	case StatusCanceled:
		return "canceled"
	case StatusIncompleteExpired:
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

func (r RefundReason) IsValid() bool {
	switch r {
	case RefundReasonCancelNow, RefundReasonLifetimeWindow, RefundReasonReplacedByLifetime,
		RefundReasonWorkspaceDeleted, RefundReasonDuplicateCheckout, RefundReasonExternal:
		return true
	case RefundReasonUnspecified:
		return false
	}
	return false
}

func (r RefundReason) String() string {
	switch r {
	case RefundReasonUnspecified:
		return "unspecified"
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
	default:
		return "unknown"
	}
}

type RefundStatus uint8

const (
	RefundStatusUnspecified RefundStatus = iota
	RefundStatusPending
	RefundStatusSucceeded
	RefundStatusFailed
	RefundStatusCanceled
)

func (s RefundStatus) IsValid() bool {
	switch s {
	case RefundStatusPending, RefundStatusSucceeded, RefundStatusFailed, RefundStatusCanceled:
		return true
	case RefundStatusUnspecified:
		return false
	}
	return false
}

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusUnspecified:
		return "unspecified"
	case RefundStatusPending:
		return "pending"
	case RefundStatusSucceeded:
		return "succeeded"
	case RefundStatusFailed:
		return "failed"
	case RefundStatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type FeatureKey uint8

const (
	FeatureKeyUnspecified FeatureKey = iota
	FeatureKeyAutoCommit
	FeatureKeyWorkspaceGraph
)

var FeatureKeys = []FeatureKey{FeatureKeyAutoCommit, FeatureKeyWorkspaceGraph}

func (k FeatureKey) IsValid() bool {
	switch k {
	case FeatureKeyAutoCommit, FeatureKeyWorkspaceGraph:
		return true
	case FeatureKeyUnspecified:
		return false
	}
	return false
}

func (k FeatureKey) String() string {
	switch k {
	case FeatureKeyUnspecified:
		return "unspecified"
	case FeatureKeyAutoCommit:
		return "auto_commit"
	case FeatureKeyWorkspaceGraph:
		return "workspace_graph"
	default:
		return "unknown"
	}
}

func ParseFeatureKey(s string) (FeatureKey, error) {
	for _, k := range FeatureKeys {
		if k.String() == s {
			return k, nil
		}
	}
	return FeatureKeyUnspecified, errs.NewInvalid(fmt.Sprintf("unknown feature %q", s))
}

type LimitKey uint8

const (
	LimitKeyUnspecified LimitKey = iota
	LimitKeyMembers
	LimitKeyPublishedNotes
	LimitKeyNoteGraphMaxDepth
	LimitKeyAutoCommitMinInterval
	LimitKeyRevisionHistoryDays
	LimitKeyTrashRetentionDays
	LimitKeyAICreditsMonthly
)

var LimitKeys = []LimitKey{
	LimitKeyMembers, LimitKeyPublishedNotes, LimitKeyNoteGraphMaxDepth, LimitKeyAutoCommitMinInterval,
	LimitKeyRevisionHistoryDays, LimitKeyTrashRetentionDays, LimitKeyAICreditsMonthly,
}

func (k LimitKey) IsValid() bool {
	switch k {
	case LimitKeyMembers, LimitKeyPublishedNotes, LimitKeyNoteGraphMaxDepth, LimitKeyAutoCommitMinInterval,
		LimitKeyRevisionHistoryDays, LimitKeyTrashRetentionDays, LimitKeyAICreditsMonthly:
		return true
	case LimitKeyUnspecified:
		return false
	}
	return false
}

func (k LimitKey) String() string {
	switch k {
	case LimitKeyUnspecified:
		return "unspecified"
	case LimitKeyMembers:
		return "members"
	case LimitKeyPublishedNotes:
		return "published_notes"
	case LimitKeyNoteGraphMaxDepth:
		return "note_graph_max_depth"
	case LimitKeyAutoCommitMinInterval:
		return "auto_commit_min_interval_seconds"
	case LimitKeyRevisionHistoryDays:
		return "revision_history_days"
	case LimitKeyTrashRetentionDays:
		return "trash_retention_days"
	case LimitKeyAICreditsMonthly:
		return "ai_credits_monthly"
	default:
		return "unknown"
	}
}

func ParseLimitKey(s string) (LimitKey, error) {
	for _, k := range LimitKeys {
		if k.String() == s {
			return k, nil
		}
	}
	return LimitKeyUnspecified, errs.NewInvalid(fmt.Sprintf("unknown limit %q", s))
}

const Unlimited int64 = -1

type Money struct {
	Amount   int64
	Currency string
}

type Payer struct {
	ID    string
	Email string
	Name  string
}
