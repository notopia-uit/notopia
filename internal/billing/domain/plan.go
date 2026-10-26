package domain

import (
	"fmt"
	"maps"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Plan struct {
	eventRecorder
	tier      PlanTier
	features  map[FeatureKey]bool
	limits    map[LimitKey]int64
	updatedAt time.Time
}

func NewPlan(tier PlanTier, features map[FeatureKey]bool, limits map[LimitKey]int64, now time.Time) (*Plan, error) {
	p := &Plan{
		eventRecorder: eventRecorder{events: nil},
		tier:          tier,
		features:      map[FeatureKey]bool{},
		limits:        map[LimitKey]int64{},
		updatedAt:     now,
	}
	if !tier.IsValid() {
		return nil, errs.NewInvalid(fmt.Sprintf("invalid tier %q", tier))
	}
	if err := p.apply(features, limits); err != nil {
		return nil, err
	}
	for _, k := range FeatureKeys {
		if _, ok := p.features[k]; !ok {
			p.features[k] = false
		}
	}
	for _, k := range LimitKeys {
		if _, ok := p.limits[k]; !ok {
			p.limits[k] = 0
		}
	}
	return p, nil
}

func UnmarshalPlan(tier PlanTier, features map[FeatureKey]bool, limits map[LimitKey]int64, updatedAt time.Time) *Plan {
	return &Plan{
		eventRecorder: eventRecorder{events: nil},
		tier:          tier,
		features:      features,
		limits:        limits,
		updatedAt:     updatedAt,
	}
}

func (p *Plan) apply(features map[FeatureKey]bool, limits map[LimitKey]int64) error {
	for k, v := range features {
		if !k.IsValid() {
			return errs.NewInvalid(fmt.Sprintf("unknown feature %q", k))
		}
		p.features[k] = v
	}
	for k, v := range limits {
		if !k.IsValid() {
			return errs.NewInvalid(fmt.Sprintf("unknown limit %q", k))
		}
		if v < Unlimited {
			return errs.NewInvalid(fmt.Sprintf("limit %q must be >= -1", k))
		}
		p.limits[k] = v
	}
	return nil
}

func (p *Plan) Update(features map[FeatureKey]bool, limits map[LimitKey]int64, now time.Time, userID string) error {
	if err := p.apply(features, limits); err != nil {
		return err
	}
	p.updatedAt = now
	p.record(&PlanUpdatedEvent{BaseEvent: NewBaseEvent(p.tier.String(), userID), Features: features, Limits: limits})
	return nil
}

func (p *Plan) Tier() PlanTier                { return p.tier }
func (p *Plan) Feature(k FeatureKey) bool     { return p.features[k] }
func (p *Plan) Limit(k LimitKey) int64        { return p.limits[k] }
func (p *Plan) Features() map[FeatureKey]bool { return maps.Clone(p.features) }
func (p *Plan) Limits() map[LimitKey]int64    { return maps.Clone(p.limits) }
func (p *Plan) UpdatedAt() time.Time          { return p.updatedAt }

func DefaultPlans(now time.Time) []*Plan {
	mk := func(t PlanTier, autoCommit, graph bool, members, published, depth, interval, revisions, trash, ai int64) *Plan {
		p, _ := NewPlan(t,
			map[FeatureKey]bool{FeatureKeyAutoCommit: autoCommit, FeatureKeyWorkspaceGraph: graph},
			map[LimitKey]int64{
				LimitKeyMembers: members, LimitKeyPublishedNotes: published, LimitKeyNoteGraphMaxDepth: depth,
				LimitKeyAutoCommitMinInterval: interval, LimitKeyRevisionHistoryDays: revisions,
				LimitKeyTrashRetentionDays: trash, LimitKeyAICreditsMonthly: ai,
			}, now)
		return p
	}
	return []*Plan{
		mk(PlanTierFree, false, false, 3, 3, 5, 0, 7, 7, 0),
		mk(PlanTierPlus, true, true, 10, Unlimited, Unlimited, 300, 30, 30, 200),
		mk(PlanTierPro, true, true, Unlimited, Unlimited, Unlimited, 60, Unlimited, 90, 2000),
	}
}
