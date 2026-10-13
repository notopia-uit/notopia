package domain

import (
	"time"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type BillingPolicy struct {
	eventRecorder
	pastDueGraceDays         int
	lifetimeRefundWindowDays int
	portalConfigurationID    *string
	updatedAt                time.Time
}

func DefaultBillingPolicy(now time.Time) *BillingPolicy {
	return &BillingPolicy{
		eventRecorder:            eventRecorder{events: nil},
		pastDueGraceDays:         3,
		lifetimeRefundWindowDays: 14,
		portalConfigurationID:    nil,
		updatedAt:                now,
	}
}

func UnmarshalBillingPolicy(grace, window int, portalConfigurationID *string, updatedAt time.Time) *BillingPolicy {
	return &BillingPolicy{
		eventRecorder:            eventRecorder{events: nil},
		pastDueGraceDays:         grace,
		lifetimeRefundWindowDays: window,
		portalConfigurationID:    portalConfigurationID,
		updatedAt:                updatedAt,
	}
}

func (p *BillingPolicy) Update(grace, window *int, now time.Time, userID string) error {
	if (grace != nil && *grace < 0) || (window != nil && *window < 0) {
		return errs.NewInvalid("policy values must be >= 0")
	}
	if grace != nil {
		p.pastDueGraceDays = *grace
	}
	if window != nil {
		p.lifetimeRefundWindowDays = *window
	}
	p.updatedAt = now
	p.record(&BillingPolicyUpdatedEvent{BaseEvent: NewBaseEvent("policy", userID)})
	return nil
}

func (p *BillingPolicy) SetPortalConfiguration(id string, now time.Time, userID string) {
	p.portalConfigurationID = &id
	p.updatedAt = now
	p.record(&PortalConfigurationChangedEvent{BaseEvent: NewBaseEvent("policy", userID), ConfigurationID: id})
}

func (p *BillingPolicy) PastDueGrace() time.Duration {
	return time.Duration(p.pastDueGraceDays) * 24 * time.Hour
}

func (p *BillingPolicy) LifetimeRefundWindow() time.Duration {
	return time.Duration(p.lifetimeRefundWindowDays) * 24 * time.Hour
}

func (p *BillingPolicy) PastDueGraceDays() int          { return p.pastDueGraceDays }
func (p *BillingPolicy) LifetimeRefundWindowDays() int  { return p.lifetimeRefundWindowDays }
func (p *BillingPolicy) PortalConfigurationID() *string { return p.portalConfigurationID }
func (p *BillingPolicy) UpdatedAt() time.Time           { return p.updatedAt }
