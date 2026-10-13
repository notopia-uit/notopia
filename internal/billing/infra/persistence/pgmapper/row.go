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

func ToPrice(row *pgsqlc.Price) *domain.Price {
	return domain.UnmarshalPrice(domain.UnmarshalPriceParams{
		ID: row.ID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
		Amount: domain.Money{Amount: row.Amount, Currency: row.Currency}, ProviderPriceID: row.ProviderPriceID,
		LookupKey: row.LookupKey, Active: row.Active, CreatedAt: row.CreatedAt, ArchivedAt: row.ArchivedAt,
	})
}

func ToPrices(rows []*pgsqlc.Price) []*domain.Price {
	out := make([]*domain.Price, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToPrice(row))
	}
	return out
}

func ToAccount(row *pgsqlc.BillingAccount) *domain.BillingAccount {
	return domain.UnmarshalBillingAccount(row.WorkspaceID, row.ProviderCustomerID,
		domain.Payer{ID: row.PayerID, Email: row.PayerEmail, Name: row.PayerName},
		row.PayerMissingSince, row.CreatedAt, row.UpdatedAt)
}

func ToSubscriptions(rows []*pgsqlc.Subscription, refundRows []*pgsqlc.Refund) []*domain.Subscription {
	refunds := map[uuid.UUID][]*domain.Refund{}
	for _, rr := range refundRows {
		refunds[rr.SubscriptionID] = append(refunds[rr.SubscriptionID], domain.UnmarshalRefund(domain.UnmarshalRefundParams{
			ID: rr.ID, ProviderRefundID: rr.ProviderRefundID, PaymentRef: rr.PaymentRef,
			Amount: domain.Money{Amount: rr.Amount, Currency: rr.Currency}, Reason: ToRefundReason(rr.Reason),
			Status: ToRefundStatus(rr.Status), FailureMessage: rr.FailureMessage, CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
		}))
	}
	out := make([]*domain.Subscription, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.UnmarshalSubscription(domain.UnmarshalSubscriptionParams{
			ID: row.ID, WorkspaceID: row.WorkspaceID, Plan: ToPlanTier(row.Plan), Interval: ToInterval(row.BillingInterval),
			PriceID: row.PriceID, Status: ToStatus(row.Status), ProviderSubscriptionID: row.ProviderSubscriptionID,
			ProviderPaymentID: row.ProviderPaymentID, CurrentPeriodStart: row.CurrentPeriodStart, CurrentPeriodEnd: row.CurrentPeriodEnd,
			CancelAtPeriodEnd: row.CancelAtPeriodEnd, CanceledAt: row.CanceledAt, PaidAt: row.PaidAt,
			ProviderSyncedAt: row.ProviderSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Refunds: refunds[row.ID],
		}))
	}
	return out
}
