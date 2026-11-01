package pgmapper

import (
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

func ToPlan(row *pgsqlc.Plan) *domain.Plan {
	return domain.UnmarshalPlan(ToPlanTier(row.Tier), row.Features, row.Limits, row.UpdatedAt)
}

func ToPlanMap(rows []*pgsqlc.Plan) map[domain.PlanTier]*domain.Plan {
	out := make(map[domain.PlanTier]*domain.Plan, len(rows))
	for _, row := range rows {
		p := ToPlan(row)
		out[p.Tier()] = p
	}
	return out
}

func ToPolicy(row *pgsqlc.BillingPolicy) *domain.BillingPolicy {
	return domain.UnmarshalBillingPolicy(int(row.PastDueGraceDays), int(row.LifetimeRefundWindowDays), row.PortalConfigurationID, row.UpdatedAt)
}

func ToPrice(row *pgsqlc.Price) (*domain.Price, error) {
	amount, err := ToMoney(row.Amount, row.Currency)
	if err != nil {
		return nil, err
	}
	return domain.UnmarshalPrice(domain.UnmarshalPriceParams{
		ID: row.ID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
		Amount: amount, ProviderPriceID: row.ProviderPriceID,
		ProviderLookupKey: row.LookupKey, Active: row.Active, CreatedAt: row.CreatedAt, ArchivedAt: row.ArchivedAt,
	}), nil
}

func ToPrices(rows []*pgsqlc.Price) ([]*domain.Price, error) {
	out := make([]*domain.Price, 0, len(rows))
	for _, row := range rows {
		p, err := ToPrice(row)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func ToAccount(row *pgsqlc.BillingAccount) (*domain.BillingAccount, error) {
	payer, err := domain.NewPayer(row.PayerID, row.PayerEmail, row.PayerName)
	if err != nil {
		return nil, err
	}
	return domain.UnmarshalBillingAccount(row.WorkspaceID, row.ProviderCustomerID, payer,
		row.PayerMissingSince, row.CreatedAt, row.UpdatedAt), nil
}

func ToSubscriptions(rows []*pgsqlc.Subscription, refundRows []*pgsqlc.Refund) ([]*domain.Subscription, error) {
	refunds := map[uuid.UUID][]*domain.Refund{}
	for _, rr := range refundRows {
		amount, err := ToMoney(rr.Amount, rr.Currency)
		if err != nil {
			return nil, err
		}
		refunds[rr.SubscriptionID] = append(refunds[rr.SubscriptionID], domain.UnmarshalRefund(domain.UnmarshalRefundParams{
			ID: rr.ID, ProviderRefundID: rr.ProviderRefundID, PaymentRef: rr.PaymentRef,
			Amount: amount, Reason: ToRefundReason(rr.Reason),
			Status: ToRefundStatus(rr.Status), FailureMessage: rr.FailureMessage, CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
		}))
	}
	out := make([]*domain.Subscription, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.UnmarshalSubscription(domain.UnmarshalSubscriptionParams{
			ID: row.ID, WorkspaceID: row.WorkspaceID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
			PriceID: row.PriceID, Status: ToStatus(row.Status), ProviderSubscriptionID: row.ProviderSubscriptionID,
			ProviderPaymentID: row.ProviderPaymentID, CurrentPeriodStart: row.CurrentPeriodStart, CurrentPeriodEnd: row.CurrentPeriodEnd,
			CancelAtPeriodEnd: row.CancelAtPeriodEnd, CanceledAt: row.CanceledAt, PastDueSince: row.PastDueSince, PaidAt: row.PaidAt,
			ProviderSyncedAt: row.ProviderSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Refunds: refunds[row.ID],
		}))
	}
	return out, nil
}
