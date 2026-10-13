package stripe

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) ensureProduct(ctx context.Context, tier string) error {
	params := &stripego.ProductCreateParams{
		ID:       stripego.String(productID(tier)),
		Name:     stripego.String("Notopia " + tier),
		Metadata: map[string]string{"tier": tier},
	}
	params.SetIdempotencyKey("product:" + tier)
	_, err := p.sc.V1Products.Create(ctx, params)
	if err != nil && !isStripeCode(err, stripego.ErrorCodeResourceAlreadyExists) {
		return providerErr("create product", err)
	}
	return nil
}

func (p *Provider) CreatePrice(ctx context.Context, in app.PriceInput) (string, error) {
	tier, err := toStripeTier(in.Plan)
	if err != nil {
		return "", err
	}
	interval, recurring, err := toStripeRecurringInterval(in.Interval)
	if err != nil {
		return "", err
	}
	if err := p.ensureProduct(ctx, tier); err != nil {
		return "", err
	}
	params := &stripego.PriceCreateParams{
		Product:           stripego.String(productID(tier)),
		Currency:          stripego.String(in.Amount.Currency),
		UnitAmount:        stripego.Int64(in.Amount.Amount),
		LookupKey:         stripego.String(domain.LookupKey(in.Plan, in.Interval)),
		TransferLookupKey: stripego.Bool(true),
		Metadata:          map[string]string{"price_id": in.PriceID.String()},
	}
	if recurring {
		params.Recurring = &stripego.PriceCreateRecurringParams{
			Interval: stripego.String(interval),
		}
	}
	params.SetIdempotencyKey("price:" + in.PriceID.String())
	pr, err := p.sc.V1Prices.Create(ctx, params)
	if err != nil {
		return "", providerErr("create price", err)
	}
	return pr.ID, nil
}

func (p *Provider) ArchivePrice(ctx context.Context, providerPriceID string) error {
	_, err := p.sc.V1Prices.Update(ctx, providerPriceID, &stripego.PriceUpdateParams{
		Active: stripego.Bool(false),
	})
	return providerErr("archive price", err)
}

func (p *Provider) CreatePortalConfiguration(ctx context.Context, activePrices []*domain.Price, returnURL string) (string, error) {
	products, err := portalProducts(activePrices)
	if err != nil {
		return "", err
	}
	features := &stripego.BillingPortalConfigurationCreateFeaturesParams{
		InvoiceHistory: &stripego.BillingPortalConfigurationCreateFeaturesInvoiceHistoryParams{
			Enabled: stripego.Bool(true),
		},
		PaymentMethodUpdate: &stripego.BillingPortalConfigurationCreateFeaturesPaymentMethodUpdateParams{
			Enabled: stripego.Bool(true),
		},
		SubscriptionCancel: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionCancelParams{
			Enabled: stripego.Bool(true),
			Mode:    stripego.String(string(stripego.BillingPortalConfigurationFeaturesSubscriptionCancelModeAtPeriodEnd)),
		},
		SubscriptionUpdate: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateParams{
			Enabled:               stripego.Bool(len(products) > 0),
			DefaultAllowedUpdates: []*string{stripego.String("price")},
			Products:              products,
			ProrationBehavior:     stripego.String("always_invoice"),
			ScheduleAtPeriodEnd: &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateScheduleAtPeriodEndParams{
				Conditions: []*stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateScheduleAtPeriodEndConditionParams{
					{Type: stripego.String("decreasing_item_amount")},
					{Type: stripego.String("shortening_interval")},
				},
			},
		},
	}
	cfg, err := p.sc.V1BillingPortalConfigurations.Create(ctx, &stripego.BillingPortalConfigurationCreateParams{
		Features:         features,
		DefaultReturnURL: stripego.String(returnURL),
	})
	if err != nil {
		return "", providerErr("create portal configuration", err)
	}
	return cfg.ID, nil
}

// portalProducts groups recurring prices by product; lifetime is never switchable in the portal.
func portalProducts(
	activePrices []*domain.Price,
) ([]*stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateProductParams, error) {
	byProduct := map[string][]*string{}
	for _, pr := range activePrices {
		if pr.Interval() == domain.IntervalLifetime {
			continue
		}
		tier, err := toStripeTier(pr.Plan())
		if err != nil {
			return nil, err
		}
		pid := productID(tier)
		byProduct[pid] = append(byProduct[pid], stripego.String(pr.ProviderPriceID()))
	}
	products := make([]*stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateProductParams, 0, len(byProduct))
	for pid, prices := range byProduct {
		products = append(products, &stripego.BillingPortalConfigurationCreateFeaturesSubscriptionUpdateProductParams{
			Product: stripego.String(pid),
			Prices:  prices,
		})
	}
	return products, nil
}
