package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type CheckoutPurpose uint8

const (
	CheckoutPurposeUnspecified CheckoutPurpose = iota
	CheckoutPurposeSubscribe
	CheckoutPurposeLifetime
	CheckoutPurposePayerChange
)

func (p CheckoutPurpose) IsValid() bool {
	switch p {
	case CheckoutPurposeSubscribe, CheckoutPurposeLifetime, CheckoutPurposePayerChange:
		return true
	case CheckoutPurposeUnspecified:
		return false
	}
	return false
}

func (p CheckoutPurpose) String() string {
	switch p {
	case CheckoutPurposeUnspecified:
		return "unspecified"
	case CheckoutPurposeSubscribe:
		return "subscribe"
	case CheckoutPurposeLifetime:
		return "lifetime"
	case CheckoutPurposePayerChange:
		return "payer_change"
	default:
		return "unknown"
	}
}

type CheckoutInput struct {
	Purpose         CheckoutPurpose
	WorkspaceID     uuid.UUID
	CustomerID      string
	Payer           domain.Payer
	PriceID         uuid.UUID // zero for payer_change
	ProviderPriceID string    // empty for payer_change
	SuccessURL      string
	CancelURL       string
}

type CheckoutSessionResult struct {
	ID                     string
	Purpose                CheckoutPurpose
	Complete               bool
	WorkspaceID            uuid.UUID
	Payer                  domain.Payer
	PriceID                uuid.UUID
	CustomerID             string
	ProviderSubscriptionID string
	PaymentMethodID        string
	Payment                domain.PaymentInfo // lifetime only
}

type RefundResult struct {
	ProviderRefundID string
	RefundID         *uuid.UUID // from metadata, nil for dashboard refunds
	PaymentRef       string
	Amount           domain.Money
	Status           domain.RefundStatus
	FailureMessage   *string
	// ProviderSubscriptionID owns the refunded payment; set by GetRefund for dashboard refunds, empty when unknown.
	ProviderSubscriptionID string
}

type PriceInput struct {
	PriceID  uuid.UUID
	Plan     domain.PlanTier
	Interval domain.Interval
	Amount   domain.Money
}

type PaymentProvider interface {
	CreateCustomer(ctx context.Context, workspaceID uuid.UUID, payer domain.Payer) (string, error)
	UpdateCustomerContact(ctx context.Context, customerID string, payer domain.Payer) error
	CreateCheckoutSession(ctx context.Context, in CheckoutInput) (string, error)
	GetCheckoutSession(ctx context.Context, id string) (CheckoutSessionResult, error)
	CreatePortalSession(ctx context.Context, customerID, configurationID, returnURL string) (string, error)
	GetSubscriptionSnapshot(ctx context.Context, providerSubscriptionID string) (domain.ProviderSnapshot, error)
	SetCancelAtPeriodEnd(ctx context.Context, providerSubscriptionID string, cancel bool) error
	CancelSubscriptionNow(ctx context.Context, providerSubscriptionID string) error
	// PeriodPayments lists every paid invoice payment whose invoice lines overlap [periodStart, periodEnd).
	PeriodPayments(ctx context.Context, providerSubscriptionID string, periodStart, periodEnd time.Time) ([]domain.PaymentInfo, error)
	PaymentForIntent(ctx context.Context, paymentIntentID string) (domain.PaymentInfo, error)
	CreateRefund(ctx context.Context, refundID uuid.UUID, paymentRef string, amount domain.Money) (RefundResult, error)
	GetRefund(ctx context.Context, providerRefundID string) (RefundResult, error)
	SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string, providerSubscriptionID *string) error
	// DetachPaymentMethods detaches every payment method except keepPaymentMethodID ("" detaches all).
	DetachPaymentMethods(ctx context.Context, customerID, keepPaymentMethodID string) error
	CreatePrice(ctx context.Context, in PriceInput) (string, error)
	ArchivePrice(ctx context.Context, providerPriceID string) error
	CreatePortalConfiguration(ctx context.Context, activePrices []*domain.Price, returnURL string) (string, error)
}
