package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type subscriptionRepo struct{ q *pgsqlc.Queries }

func (r *subscriptionRepo) hydrate(ctx context.Context, rows []*pgsqlc.Subscription) ([]*domain.Subscription, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	refundRows, err := r.q.ListRefundsBySubscriptionIDs(ctx, ids)
	if err != nil {
		return nil, wrap("list refunds", err)
	}
	return pgmapper.ToSubscriptions(rows, refundRows), nil
}

func (r *subscriptionRepo) one(ctx context.Context, row *pgsqlc.Subscription, err error, missingIsNil bool, what string) (*domain.Subscription, error) {
	if notFound(err) {
		if missingIsNil {
			return nil, nil
		}
		return nil, errs.NewNotFound(fmt.Sprintf("subscription %s not found", what))
	}
	if err != nil {
		return nil, wrap("get subscription", err)
	}
	subs, err := r.hydrate(ctx, []*pgsqlc.Subscription{row})
	if err != nil {
		return nil, err
	}
	return subs[0], nil
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscription(ctx, pgsqlc.GetSubscriptionParams{ID: id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, false, id.String())
}

func (r *subscriptionRepo) GetCurrentByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetCurrentSubscription(ctx, pgsqlc.GetCurrentSubscriptionParams{WorkspaceID: workspaceID, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByProviderSubscriptionID(ctx context.Context, id string, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscriptionByProviderSubscriptionID(ctx, pgsqlc.GetSubscriptionByProviderSubscriptionIDParams{ProviderSubscriptionID: &id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByProviderPaymentID(ctx context.Context, id string, forUpdate bool) (*domain.Subscription, error) {
	row, err := r.q.GetSubscriptionByProviderPaymentID(ctx, pgsqlc.GetSubscriptionByProviderPaymentIDParams{ProviderPaymentID: &id, ForUpdate: forUpdate})
	return r.one(ctx, row, err, true, "")
}

func (r *subscriptionRepo) GetByRefundID(ctx context.Context, refundID uuid.UUID, forUpdate bool) (*domain.Subscription, error) {
	subID, err := r.q.GetSubscriptionIDByRefundID(ctx, refundID)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, wrap("get subscription id by refund", err)
	}
	return r.GetByID(ctx, subID, forUpdate)
}

func (r *subscriptionRepo) ListCurrentByWorkspaces(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error) {
	rows, err := r.q.ListCurrentSubscriptionsByWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, wrap("list current subscriptions", err)
	}
	return r.hydrate(ctx, rows)
}

func (r *subscriptionRepo) ListIDsWithUnsubmittedRefunds(ctx context.Context, limit int) ([]uuid.UUID, error) {
	ids, err := r.q.ListSubscriptionIDsWithUnsubmittedRefunds(ctx, int32(limit))
	if err != nil {
		return nil, wrap("list subscriptions with unsubmitted refunds", err)
	}
	return ids, nil
}

func (r *subscriptionRepo) ListNeedingReconcile(ctx context.Context, now time.Time, limit int) ([]*domain.Subscription, error) {
	rows, err := r.q.ListSubscriptionsNeedingReconcile(ctx, &pgsqlc.ListSubscriptionsNeedingReconcileParams{Now: now, RowLimit: int32(limit)})
	if err != nil {
		return nil, wrap("list subscriptions needing reconcile", err)
	}
	return r.hydrate(ctx, rows)
}

func (r *subscriptionRepo) ListCanceledAwaitingProviderCancel(ctx context.Context, limit int) ([]*domain.Subscription, error) {
	rows, err := r.q.ListCanceledSubscriptionsAwaitingProviderCancel(ctx, int32(limit))
	if err != nil {
		return nil, wrap("list canceled subscriptions awaiting provider cancel", err)
	}
	return r.hydrate(ctx, rows)
}

func (r *subscriptionRepo) Save(ctx context.Context, s *domain.Subscription) error {
	err := r.q.UpsertSubscription(ctx, &pgsqlc.UpsertSubscriptionParams{
		ID: s.ID(), WorkspaceID: s.WorkspaceID(), Plan: pgmapper.FromPlanTier(s.Plan()), BillingInterval: pgmapper.FromInterval(s.Interval()),
		PriceID: s.PriceID(), Status: pgmapper.FromStatus(s.Status()), ProviderSubscriptionID: s.ProviderSubscriptionID(),
		ProviderPaymentID: s.ProviderPaymentID(), CurrentPeriodStart: s.CurrentPeriodStart(), CurrentPeriodEnd: s.CurrentPeriodEnd(),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd(), CanceledAt: s.CanceledAt(), PastDueSince: s.PastDueSince(), PaidAt: s.PaidAt(),
		ProviderSyncedAt: s.ProviderSyncedAt(), CreatedAt: s.CreatedAt(), UpdatedAt: s.UpdatedAt(),
	})
	if err != nil {
		return wrap("save subscription", err)
	}
	for _, rf := range s.Refunds() {
		if err := r.q.UpsertRefund(ctx, &pgsqlc.UpsertRefundParams{
			ID: rf.ID(), SubscriptionID: s.ID(), ProviderRefundID: rf.ProviderRefundID(), PaymentRef: rf.PaymentRef(),
			Amount: rf.Amount().Amount, Currency: rf.Amount().Currency, Reason: pgmapper.FromRefundReason(rf.Reason()),
			Status: pgmapper.FromRefundStatus(rf.Status()), FailureMessage: rf.FailureMessage(), CreatedAt: rf.CreatedAt(), UpdatedAt: rf.UpdatedAt(),
		}); err != nil {
			return wrap("save refund", err)
		}
	}
	drain(ctx, s.PopEvents())
	return nil
}
