package pgreadmodel

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type BillingReadModel struct{ q *pgsqlc.Queries }

var _ app.BillingReadModel = (*BillingReadModel)(nil)

func NewBillingReadModel(q *pgsqlc.Queries) *BillingReadModel { return &BillingReadModel{q: q} }

var ProvideBillingReadModel = NewBillingReadModel

func toErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return errs.NewInternal(msg, err)
}

func (m *BillingReadModel) subscriptions(ctx context.Context, rows []*pgsqlc.Subscription) ([]*domain.Subscription, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	refundRows, err := m.q.ReadListRefundsBySubscriptionIDs(ctx, ids)
	if err != nil {
		return nil, toErr("read refunds", err)
	}
	return pgmapper.ToSubscriptions(rows, refundRows), nil
}

func (m *BillingReadModel) CurrentSubscription(ctx context.Context, workspaceID uuid.UUID) (*domain.Subscription, error) {
	row, err := m.q.ReadGetCurrentSubscription(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, toErr("read current subscription", err)
	}
	subs, err := m.subscriptions(ctx, []*pgsqlc.Subscription{row})
	if err != nil {
		return nil, err
	}
	return subs[0], nil
}

func (m *BillingReadModel) CurrentSubscriptions(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error) {
	rows, err := m.q.ReadListCurrentSubscriptionsByWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, toErr("read current subscriptions", err)
	}
	return m.subscriptions(ctx, rows)
}

func (m *BillingReadModel) Account(ctx context.Context, workspaceID uuid.UUID) (*domain.BillingAccount, error) {
	row, err := m.q.ReadGetBillingAccount(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, toErr("read billing account", err)
	}
	return pgmapper.ToAccount(row), nil
}

func (m *BillingReadModel) Plans(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	rows, err := m.q.ReadListPlans(ctx)
	if err != nil {
		return nil, toErr("read plans", err)
	}
	return pgmapper.ToPlanMap(rows), nil
}

func (m *BillingReadModel) Policy(ctx context.Context) (*domain.BillingPolicy, error) {
	row, err := m.q.ReadGetPolicy(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound("billing policy not seeded")
	}
	if err != nil {
		return nil, toErr("read policy", err)
	}
	return pgmapper.ToPolicy(row), nil
}

func (m *BillingReadModel) ActivePrices(ctx context.Context) ([]*domain.Price, error) {
	rows, err := m.q.ReadListActivePrices(ctx)
	if err != nil {
		return nil, toErr("read active prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}

func (m *BillingReadModel) Prices(ctx context.Context) ([]*domain.Price, error) {
	rows, err := m.q.ReadListPrices(ctx)
	if err != nil {
		return nil, toErr("read prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}
