package stripe

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
)

func toRefundResult(r *stripego.Refund) app.RefundResult {
	res := app.RefundResult{
		ProviderRefundID: r.ID,
		RefundID:         nil,
		PaymentRef:       "",
		Amount: domain.Money{
			Amount:   r.Amount,
			Currency: string(r.Currency),
		},
		Status:                 toDomainRefundStatus(r.Status),
		FailureMessage:         nil,
		ProviderSubscriptionID: "",
	}
	if r.PaymentIntent != nil {
		res.PaymentRef = r.PaymentIntent.ID
	}
	if id, err := uuid.Parse(r.Metadata["refund_id"]); err == nil {
		res.RefundID = &id
	}
	if r.FailureReason != "" {
		res.FailureMessage = new(string(r.FailureReason))
	}
	return res
}

func (p *Provider) CreateRefund(ctx context.Context, refundID uuid.UUID, paymentRef string, amount domain.Money) (app.RefundResult, error) {
	params := &stripego.RefundCreateParams{
		PaymentIntent: stripego.String(paymentRef),
		Amount:        stripego.Int64(amount.Amount),
		Metadata:      map[string]string{"refund_id": refundID.String()},
	}
	params.SetIdempotencyKey("refund:" + refundID.String())
	r, err := p.sc.V1Refunds.Create(ctx, params)
	if isInvalidRequest(err) {
		return app.RefundResult{}, errs.New(errs.CodeInvalid, "stripe create refund rejected", err)
	}
	if err != nil {
		return app.RefundResult{}, providerErr("create refund", err)
	}
	return toRefundResult(r), nil
}

// subscriptionOfPayment resolves the subscription that owns a payment intent through its invoice payment.
func (p *Provider) subscriptionOfPayment(ctx context.Context, paymentIntentID string) (string, error) {
	params := &stripego.InvoicePaymentListParams{
		Payment: &stripego.InvoicePaymentListPaymentParams{
			Type:          stripego.String("payment_intent"),
			PaymentIntent: stripego.String(paymentIntentID),
		},
	}
	params.Limit = stripego.Int64(1)
	params.AddExpand("data.invoice")
	for ip, err := range p.sc.V1InvoicePayments.List(ctx, params).All(ctx) {
		if err != nil {
			return "", providerErr("list invoice payments", err)
		}
		if ip.Invoice == nil || ip.Invoice.Parent == nil || ip.Invoice.Parent.SubscriptionDetails == nil ||
			ip.Invoice.Parent.SubscriptionDetails.Subscription == nil {
			return "", nil
		}
		return ip.Invoice.Parent.SubscriptionDetails.Subscription.ID, nil
	}
	return "", nil
}

func (p *Provider) GetRefund(ctx context.Context, providerRefundID string) (app.RefundResult, error) {
	r, err := p.sc.V1Refunds.Retrieve(ctx, providerRefundID, nil)
	if err != nil {
		return app.RefundResult{}, providerErr("retrieve refund", err)
	}
	res := toRefundResult(r)
	if res.RefundID == nil && res.PaymentRef != "" {
		if res.ProviderSubscriptionID, err = p.subscriptionOfPayment(ctx, res.PaymentRef); err != nil {
			return app.RefundResult{}, err
		}
	}
	return res, nil
}
