package stripe

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) GetSubscriptionSnapshot(ctx context.Context, providerSubscriptionID string) (domain.ProviderSnapshot, error) {
	s, err := p.sc.V1Subscriptions.Retrieve(ctx, providerSubscriptionID, nil)
	if err != nil {
		return domain.ProviderSnapshot{}, providerErr("retrieve subscription", err)
	}
	return snapshotFromSubscription(s)
}

func (p *Provider) SetCancelAtPeriodEnd(ctx context.Context, providerSubscriptionID string, cancel bool) error {
	_, err := p.sc.V1Subscriptions.Update(ctx, providerSubscriptionID, &stripego.SubscriptionUpdateParams{
		CancelAtPeriodEnd: stripego.Bool(cancel),
	})
	return providerErr("update subscription cancel_at_period_end", err)
}

func (p *Provider) CancelSubscriptionNow(ctx context.Context, providerSubscriptionID string) error {
	_, err := p.sc.V1Subscriptions.Cancel(ctx, providerSubscriptionID, &stripego.SubscriptionCancelParams{
		Prorate:    stripego.Bool(false),
		InvoiceNow: stripego.Bool(false),
	})
	if isStripeCode(err, stripego.ErrorCodeResourceMissing) {
		return nil
	}
	return providerErr("cancel subscription", err)
}

// PeriodPayments: a period can hold the renewal invoice plus proration invoices from upgrades
// (always_invoice). Coverage = the invoice's line periods clipped to the current period.
func (p *Provider) PeriodPayments(
	ctx context.Context,
	providerSubscriptionID string,
	periodStart, periodEnd time.Time,
) ([]domain.PaymentInfo, error) {
	params := &stripego.InvoiceListParams{
		Subscription: stripego.String(providerSubscriptionID),
		Status:       stripego.String(string(stripego.InvoiceStatusPaid)),
		CreatedRange: &stripego.RangeQueryParams{
			GreaterThanOrEqual: periodStart.Add(-24 * time.Hour).Unix(),
		},
	}
	var out []domain.PaymentInfo
	for inv, err := range p.sc.V1Invoices.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, providerErr("list invoices", err)
		}
		covStart, covEnd := invoiceCoverage(inv)
		if !covEnd.After(periodStart) || !covStart.Before(periodEnd) {
			continue
		}
		payments, err := p.invoicePayments(ctx, inv, covStart, covEnd)
		if err != nil {
			return nil, err
		}
		out = append(out, payments...)
	}
	if len(out) == 0 {
		return nil, errs.NewNotFound("no paid payment in the current period")
	}
	return out, nil
}

func invoiceCoverage(inv *stripego.Invoice) (time.Time, time.Time) {
	var covStart, covEnd time.Time
	if inv.Lines == nil {
		return covStart, covEnd
	}
	for _, l := range inv.Lines.Data {
		if l.Period == nil {
			continue
		}
		ls, le := time.Unix(l.Period.Start, 0), time.Unix(l.Period.End, 0)
		if covStart.IsZero() || ls.Before(covStart) {
			covStart = ls
		}
		if le.After(covEnd) {
			covEnd = le
		}
	}
	return covStart, covEnd
}

func (p *Provider) invoicePayments(
	ctx context.Context,
	inv *stripego.Invoice,
	covStart, covEnd time.Time,
) ([]domain.PaymentInfo, error) {
	params := &stripego.InvoicePaymentListParams{
		Invoice: stripego.String(inv.ID),
		Status:  stripego.String("paid"),
	}
	var out []domain.PaymentInfo
	for ip, err := range p.sc.V1InvoicePayments.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, providerErr("list invoice payments", err)
		}
		if ip.Payment == nil || ip.Payment.PaymentIntent == nil {
			continue
		}
		paidAt := time.Unix(inv.Created, 0)
		if ip.StatusTransitions != nil && ip.StatusTransitions.PaidAt > 0 {
			paidAt = time.Unix(ip.StatusTransitions.PaidAt, 0)
		}
		out = append(out, domain.PaymentInfo{
			Amount: domain.Money{
				Amount:   ip.AmountPaid,
				Currency: string(ip.Currency),
			},
			PaymentRef:    ip.Payment.PaymentIntent.ID,
			PaidAt:        paidAt,
			CoverageStart: covStart,
			CoverageEnd:   covEnd,
		})
	}
	return out, nil
}

func (p *Provider) PaymentForIntent(ctx context.Context, paymentIntentID string) (domain.PaymentInfo, error) {
	pi, err := p.sc.V1PaymentIntents.Retrieve(ctx, paymentIntentID, nil)
	if err != nil {
		return domain.PaymentInfo{}, providerErr("retrieve payment intent", err)
	}
	return paymentFromIntent(pi), nil
}
