package stripe

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
)

const (
	purposeSubscribe   = "subscribe"
	purposeLifetime    = "lifetime"
	purposePayerChange = "payer_change"
)

func toStripeTier(t domain.PlanTier) (string, error) {
	switch t {
	case domain.PlanTierPlus:
		return "plus", nil
	case domain.PlanTierPro:
		return "pro", nil
	case domain.PlanTierUnspecified, domain.PlanTierFree:
	}
	return "", errs.NewInvalid(fmt.Sprintf("plan %v cannot be sold", t))
}

// toStripeRecurringInterval returns recurring=false for lifetime.
func toStripeRecurringInterval(i domain.Interval) (string, bool, error) {
	switch i {
	case domain.IntervalMonth:
		return string(stripego.PriceRecurringIntervalMonth), true, nil
	case domain.IntervalYear:
		return string(stripego.PriceRecurringIntervalYear), true, nil
	case domain.IntervalLifetime:
		return "", false, nil
	case domain.IntervalUnspecified:
	}
	return "", false, errs.NewInvalid(fmt.Sprintf("invalid interval %v", i))
}

func toStripePurpose(p app.CheckoutPurpose) (string, error) {
	switch p {
	case app.CheckoutPurposeSubscribe:
		return purposeSubscribe, nil
	case app.CheckoutPurposeLifetime:
		return purposeLifetime, nil
	case app.CheckoutPurposePayerChange:
		return purposePayerChange, nil
	case app.CheckoutPurposeUnspecified:
	}
	return "", errs.NewInternal(fmt.Sprintf("invalid checkout purpose %v", p), nil)
}

// Sessions not created by billing have no purpose; CompleteCheckout ignores CheckoutPurposeUnspecified.
func toAppCheckoutPurpose(s string) app.CheckoutPurpose {
	switch s {
	case purposeSubscribe:
		return app.CheckoutPurposeSubscribe
	case purposeLifetime:
		return app.CheckoutPurposeLifetime
	case purposePayerChange:
		return app.CheckoutPurposePayerChange
	default:
		return app.CheckoutPurposeUnspecified
	}
}

func toDomainStatus(s stripego.SubscriptionStatus) domain.Status {
	switch s {
	case stripego.SubscriptionStatusActive, stripego.SubscriptionStatusTrialing:
		return domain.StatusActive
	case stripego.SubscriptionStatusPastDue:
		return domain.StatusPastDue
	case stripego.SubscriptionStatusIncomplete:
		return domain.StatusIncomplete
	case stripego.SubscriptionStatusIncompleteExpired:
		return domain.StatusIncompleteExpired
	case stripego.SubscriptionStatusCanceled:
		return domain.StatusCanceled
	case stripego.SubscriptionStatusUnpaid, stripego.SubscriptionStatusPaused:
		return domain.StatusUnpaid
	default:
		return domain.StatusUnpaid
	}
}

func toDomainRefundStatus(s stripego.RefundStatus) domain.RefundStatus {
	switch s {
	case stripego.RefundStatusSucceeded:
		return domain.RefundStatusSucceeded
	case stripego.RefundStatusFailed:
		return domain.RefundStatusFailed
	case stripego.RefundStatusCanceled:
		return domain.RefundStatusCanceled
	case stripego.RefundStatusPending, stripego.RefundStatusRequiresAction:
		return domain.RefundStatusPending
	default:
		return domain.RefundStatusPending
	}
}

func snapshotFromSubscription(s *stripego.Subscription) (domain.ProviderSnapshot, error) {
	if s.Items == nil || len(s.Items.Data) == 0 || s.Items.Data[0].Price == nil {
		return domain.ProviderSnapshot{}, errs.NewInternal("stripe subscription "+s.ID+" has no items", nil)
	}
	item := s.Items.Data[0]
	ws, err := uuid.Parse(s.Metadata["workspace_id"])
	if err != nil {
		return domain.ProviderSnapshot{}, errs.NewInternal("stripe subscription "+s.ID+" has no workspace_id metadata", err)
	}
	var canceledAt *time.Time
	if s.CanceledAt > 0 {
		canceledAt = new(time.Unix(s.CanceledAt, 0))
	}
	return domain.ProviderSnapshot{
		ProviderSubscriptionID: s.ID,
		ProviderPriceID:        item.Price.ID,
		WorkspaceID:            ws,
		Status:                 toDomainStatus(s.Status),
		CurrentPeriodStart:     time.Unix(item.CurrentPeriodStart, 0),
		CurrentPeriodEnd:       time.Unix(item.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:      s.CancelAtPeriodEnd || s.CancelAt > 0,
		CanceledAt:             canceledAt,
	}, nil
}

func paymentFromIntent(pi *stripego.PaymentIntent) domain.PaymentInfo {
	return domain.PaymentInfo{
		Amount: domain.Money{
			Amount:   pi.AmountReceived,
			Currency: string(pi.Currency),
		},
		PaymentRef:    pi.ID,
		PaidAt:        time.Unix(pi.Created, 0),
		CoverageStart: time.Time{},
		CoverageEnd:   time.Time{},
	}
}
