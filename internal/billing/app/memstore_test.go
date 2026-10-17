package app

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type memStore struct {
	mu       sync.Mutex
	subs     map[uuid.UUID]*domain.Subscription
	accounts map[uuid.UUID]*domain.BillingAccount
	prices   map[uuid.UUID]*domain.Price
	plans    map[domain.PlanTier]*domain.Plan
	policy   *domain.BillingPolicy
	events   map[string]bool
}

func newMemStore() *memStore {
	now := time.Now()
	s := &memStore{
		subs: map[uuid.UUID]*domain.Subscription{}, accounts: map[uuid.UUID]*domain.BillingAccount{},
		prices: map[uuid.UUID]*domain.Price{}, plans: map[domain.PlanTier]*domain.Plan{},
		policy: domain.DefaultBillingPolicy(now), events: map[string]bool{},
	}
	for _, p := range domain.DefaultPlans(now) {
		s.plans[p.Tier()] = p
	}
	return s
}

func (m *memStore) Execute(ctx context.Context, fn func(domain.RepoRegistry) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(m)
}

func (m *memStore) Subscription() domain.SubscriptionRepo     { return memSubs{m} }
func (m *memStore) BillingAccount() domain.BillingAccountRepo { return memAccounts{m} }
func (m *memStore) Price() domain.PriceRepo                   { return memPrices{m} }
func (m *memStore) Plan() domain.PlanRepo                     { return memPlans{m} }
func (m *memStore) BillingPolicy() domain.BillingPolicyRepo   { return memPolicy{m} }
func (m *memStore) ProviderEventLog() domain.ProviderEventLog { return memEvents{m} }

type memSubs struct{ m *memStore }

func (r memSubs) find(pred func(*domain.Subscription) bool) *domain.Subscription {
	for _, s := range r.m.subs {
		if pred(s) {
			return s
		}
	}
	return nil
}

func (r memSubs) GetByID(_ context.Context, id uuid.UUID, _ bool) (*domain.Subscription, error) {
	if s, ok := r.m.subs[id]; ok {
		return s, nil
	}
	return nil, errs.NewNotFound("sub")
}

func (r memSubs) GetCurrentByWorkspace(_ context.Context, ws uuid.UUID, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { return s.WorkspaceID() == ws && s.IsCurrent() }), nil
}

func (r memSubs) GetByProviderSubscriptionID(_ context.Context, id string, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { p := s.ProviderSubscriptionID(); return p != nil && *p == id }), nil
}

func (r memSubs) GetByProviderPaymentID(_ context.Context, id string, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool { p := s.ProviderPaymentID(); return p != nil && *p == id }), nil
}

func (r memSubs) GetByRefundID(_ context.Context, id uuid.UUID, _ bool) (*domain.Subscription, error) {
	return r.find(func(s *domain.Subscription) bool {
		for _, rf := range s.Refunds() {
			if rf.ID() == id {
				return true
			}
		}
		return false
	}), nil
}

func (r memSubs) ListCurrentByWorkspaces(_ context.Context, ids []uuid.UUID) ([]*domain.Subscription, error) {
	var out []*domain.Subscription
	for _, s := range r.m.subs {
		for _, id := range ids {
			if s.WorkspaceID() == id && s.IsCurrent() {
				out = append(out, s)
			}
		}
	}
	return out, nil
}

func (r memSubs) ListNeedingReconcile(context.Context, time.Time, int) ([]*domain.Subscription, error) {
	return nil, nil
}

func (r memSubs) ListCanceledAwaitingProviderCancel(_ context.Context, limit int) ([]*domain.Subscription, error) {
	var out []*domain.Subscription
	for _, s := range r.m.subs {
		if len(out) == limit {
			break
		}
		if s.Status() == domain.StatusCanceled && s.ProviderSubscriptionID() != nil && s.CanceledAt() != nil &&
			!s.ProviderSyncedAt().After(*s.CanceledAt()) {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r memSubs) ListIDsWithUnsubmittedRefunds(context.Context, int) ([]uuid.UUID, error) {
	return nil, nil
}

func (r memSubs) Save(_ context.Context, s *domain.Subscription) error {
	if s.IsCurrent() {
		if cur := r.find(func(o *domain.Subscription) bool {
			return o.WorkspaceID() == s.WorkspaceID() && o.IsCurrent() && o.ID() != s.ID()
		}); cur != nil {
			return errs.NewInternal("unique current subscription violated", nil)
		}
	}
	s.PopEvents()
	r.m.subs[s.ID()] = s
	return nil
}

type memAccounts struct{ m *memStore }

func (r memAccounts) GetByWorkspace(_ context.Context, ws uuid.UUID, _ bool) (*domain.BillingAccount, error) {
	return r.m.accounts[ws], nil
}

func (r memAccounts) ListByPayer(_ context.Context, payerID string) ([]*domain.BillingAccount, error) {
	var out []*domain.BillingAccount
	for _, a := range r.m.accounts {
		if a.Payer().ID == payerID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r memAccounts) Save(_ context.Context, a *domain.BillingAccount) error {
	a.PopEvents()
	r.m.accounts[a.WorkspaceID()] = a
	return nil
}

type memPrices struct{ m *memStore }

func (r memPrices) GetByID(_ context.Context, id uuid.UUID, _ bool) (*domain.Price, error) {
	if p, ok := r.m.prices[id]; ok {
		return p, nil
	}
	return nil, errs.NewNotFound("price")
}

func (r memPrices) GetByProviderPriceID(_ context.Context, id string) (*domain.Price, error) {
	for _, p := range r.m.prices {
		if p.ProviderPriceID() == id {
			return p, nil
		}
	}
	return nil, errs.NewNotFound("price")
}

func (r memPrices) GetActive(_ context.Context, plan domain.PlanTier, iv domain.Interval, cur string, _ bool) (*domain.Price, error) {
	for _, p := range r.m.prices {
		if p.Active() && p.Plan() == plan && p.Interval() == iv && p.Amount().Currency == cur {
			return p, nil
		}
	}
	return nil, nil
}

func (r memPrices) ListActive(context.Context) ([]*domain.Price, error) {
	var out []*domain.Price
	for _, p := range r.m.prices {
		if p.Active() {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r memPrices) List(ctx context.Context) ([]*domain.Price, error) { return r.ListActive(ctx) }
func (r memPrices) Save(_ context.Context, p *domain.Price) error {
	p.PopEvents()
	r.m.prices[p.ID()] = p
	return nil
}

type memPlans struct{ m *memStore }

func (r memPlans) Get(_ context.Context, t domain.PlanTier, _ bool) (*domain.Plan, error) {
	return r.m.plans[t], nil
}

func (r memPlans) ListMap(context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	return r.m.plans, nil
}

func (r memPlans) Save(_ context.Context, p *domain.Plan) error { r.m.plans[p.Tier()] = p; return nil }

type memPolicy struct{ m *memStore }

func (r memPolicy) Get(context.Context, bool) (*domain.BillingPolicy, error) { return r.m.policy, nil }

func (r memPolicy) Save(_ context.Context, p *domain.BillingPolicy) error { r.m.policy = p; return nil }

type memEvents struct{ m *memStore }

func (r memEvents) MarkProcessed(_ context.Context, id, _ string) (bool, error) {
	if r.m.events[id] {
		return false, nil
	}
	r.m.events[id] = true
	return true, nil
}
func (r memEvents) PruneBefore(context.Context, time.Time) (int64, error) { return 0, nil }
