package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Price struct {
	eventRecorder
	id              uuid.UUID
	plan            PlanTier
	interval        Interval
	amount          Money
	providerPriceID string
	lookupKey       string
	active          bool
	createdAt       time.Time
	archivedAt      *time.Time
}

func LookupKey(plan PlanTier, interval Interval) string {
	return plan.String() + "_" + interval.String()
}

func NewPrice(id uuid.UUID, plan PlanTier, interval Interval, amount Money, providerPriceID string, now time.Time, userID string) (*Price, error) {
	if !plan.IsValid() || plan == TierFree {
		return nil, errs.NewInvalid("price plan must be plus or pro")
	}
	if !interval.IsValid() {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid interval %q", interval))
	}
	if amount.Amount <= 0 || amount.Currency == "" {
		return nil, errs.NewInvalid("price amount must be > 0 with a currency")
	}
	p := &Price{
		eventRecorder:   eventRecorder{events: nil},
		id:              id,
		plan:            plan,
		interval:        interval,
		amount:          amount,
		providerPriceID: providerPriceID,
		lookupKey:       LookupKey(plan, interval),
		active:          true,
		createdAt:       now,
		archivedAt:      nil,
	}
	p.record(&PriceCreatedEvent{BaseEvent: NewBaseEvent(id.String(), userID)})
	return p, nil
}

type UnmarshalPriceParams struct {
	ID              uuid.UUID
	Plan            PlanTier
	Interval        Interval
	Amount          Money
	ProviderPriceID string
	LookupKey       string
	Active          bool
	CreatedAt       time.Time
	ArchivedAt      *time.Time
}

func UnmarshalPrice(p UnmarshalPriceParams) *Price {
	return &Price{
		eventRecorder:   eventRecorder{events: nil},
		id:              p.ID,
		plan:            p.Plan,
		interval:        p.Interval,
		amount:          p.Amount,
		providerPriceID: p.ProviderPriceID,
		lookupKey:       p.LookupKey,
		active:          p.Active,
		createdAt:       p.CreatedAt,
		archivedAt:      p.ArchivedAt,
	}
}

func (p *Price) Archive(now time.Time, userID string) error {
	if !p.active {
		return nil
	}
	p.active = false
	p.archivedAt = &now
	p.record(&PriceArchivedEvent{BaseEvent: NewBaseEvent(p.id.String(), userID)})
	return nil
}

func (p *Price) ID() uuid.UUID           { return p.id }
func (p *Price) Plan() PlanTier          { return p.plan }
func (p *Price) Interval() Interval      { return p.interval }
func (p *Price) Amount() Money           { return p.amount }
func (p *Price) ProviderPriceID() string { return p.providerPriceID }
func (p *Price) LookupKey() string       { return p.lookupKey }
func (p *Price) Active() bool            { return p.active }
func (p *Price) CreatedAt() time.Time    { return p.createdAt }
func (p *Price) ArchivedAt() *time.Time  { return p.archivedAt }
