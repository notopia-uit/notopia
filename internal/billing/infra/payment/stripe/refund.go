package stripe

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
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
		Status:         toDomainRefundStatus(r.Status),
		FailureMessage: nil,
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
	if err != nil {
		return app.RefundResult{}, providerErr("create refund", err)
	}
	return toRefundResult(r), nil
}

func (p *Provider) GetRefund(ctx context.Context, providerRefundID string) (app.RefundResult, error) {
	r, err := p.sc.V1Refunds.Retrieve(ctx, providerRefundID, nil)
	if err != nil {
		return app.RefundResult{}, providerErr("retrieve refund", err)
	}
	return toRefundResult(r), nil
}
