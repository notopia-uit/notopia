package stripe

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	stripego "github.com/stripe/stripe-go/v87"
)

func (p *Provider) CreateCustomer(ctx context.Context, workspaceID uuid.UUID, payer domain.Payer) (string, error) {
	params := &stripego.CustomerCreateParams{
		Email: stripego.String(payer.Email),
		Name:  stripego.String(payer.Name),
		Metadata: map[string]string{
			"workspace_id": workspaceID.String(),
			"payer_id":     payer.ID,
		},
	}
	params.SetIdempotencyKey("customer:" + workspaceID.String())
	c, err := p.sc.V1Customers.Create(ctx, params)
	if err != nil {
		return "", providerErr("create customer", err)
	}
	return c.ID, nil
}

func (p *Provider) UpdateCustomerContact(ctx context.Context, customerID string, payer domain.Payer) error {
	_, err := p.sc.V1Customers.Update(ctx, customerID, &stripego.CustomerUpdateParams{
		Email:    stripego.String(payer.Email),
		Name:     stripego.String(payer.Name),
		Metadata: map[string]string{"payer_id": payer.ID},
	})
	return providerErr("update customer", err)
}

func (p *Provider) CreateCheckoutSession(ctx context.Context, in app.CheckoutInput) (string, error) {
	purpose, err := toStripePurpose(in.Purpose)
	if err != nil {
		return "", err
	}
	meta := map[string]string{
		"workspace_id": in.WorkspaceID.String(),
		"payer_id":     in.Payer.ID,
		"payer_email":  in.Payer.Email,
		"payer_name":   in.Payer.Name,
		"purpose":      purpose,
	}
	if in.PriceID != uuid.Nil {
		meta["price_id"] = in.PriceID.String()
	}
	params := &stripego.CheckoutSessionCreateParams{
		Customer:          stripego.String(in.CustomerID),
		ClientReferenceID: stripego.String(in.WorkspaceID.String()),
		SuccessURL:        stripego.String(in.SuccessURL),
		CancelURL:         stripego.String(in.CancelURL),
		Metadata:          meta,
	}
	lineItems := []*stripego.CheckoutSessionCreateLineItemParams{
		{
			Price:    stripego.String(in.ProviderPriceID),
			Quantity: stripego.Int64(1),
		},
	}
	switch in.Purpose {
	case app.CheckoutPurposeSubscribe:
		params.Mode = stripego.String(string(stripego.CheckoutSessionModeSubscription))
		params.LineItems = lineItems
		params.SubscriptionData = &stripego.CheckoutSessionCreateSubscriptionDataParams{Metadata: meta}
	case app.CheckoutPurposeLifetime:
		params.Mode = stripego.String(string(stripego.CheckoutSessionModePayment))
		params.LineItems = lineItems
		params.PaymentIntentData = &stripego.CheckoutSessionCreatePaymentIntentDataParams{Metadata: meta}
	case app.CheckoutPurposePayerChange:
		params.Mode = stripego.String(string(stripego.CheckoutSessionModeSetup))
		params.Currency = stripego.String(p.currency)
		params.SetupIntentData = &stripego.CheckoutSessionCreateSetupIntentDataParams{Metadata: meta}
	case app.CheckoutPurposeUnspecified:
	}
	s, err := p.sc.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return "", providerErr("create checkout session", err)
	}
	return s.URL, nil
}

func (p *Provider) GetCheckoutSession(ctx context.Context, id string) (app.CheckoutSessionResult, error) {
	params := &stripego.CheckoutSessionRetrieveParams{}
	params.AddExpand("setup_intent")
	params.AddExpand("payment_intent")
	s, err := p.sc.V1CheckoutSessions.Retrieve(ctx, id, params)
	if err != nil {
		return app.CheckoutSessionResult{}, providerErr("retrieve checkout session", err)
	}

	// Foreign sessions lack metadata; CompleteCheckout drops them by purpose.
	ws, _ := uuid.Parse(s.Metadata["workspace_id"])
	priceID, _ := uuid.Parse(s.Metadata["price_id"])
	paid := s.PaymentStatus == stripego.CheckoutSessionPaymentStatusPaid ||
		s.PaymentStatus == stripego.CheckoutSessionPaymentStatusNoPaymentRequired
	var payment domain.PaymentInfo
	if s.PaymentIntent != nil {
		payment = paymentFromIntent(s.PaymentIntent)
	}

	res := app.CheckoutSessionResult{
		ID:          s.ID,
		Purpose:     toAppCheckoutPurpose(s.Metadata["purpose"]),
		Complete:    s.Status == stripego.CheckoutSessionStatusComplete && paid,
		WorkspaceID: ws,
		Payer: domain.Payer{
			ID:    s.Metadata["payer_id"],
			Email: s.Metadata["payer_email"],
			Name:  s.Metadata["payer_name"],
		},
		PriceID:                priceID,
		CustomerID:             "",
		ProviderSubscriptionID: "",
		PaymentMethodID:        "",
		Payment:                payment,
	}
	if s.Customer != nil {
		res.CustomerID = s.Customer.ID
	}
	if s.Subscription != nil {
		res.ProviderSubscriptionID = s.Subscription.ID
	}
	if s.SetupIntent != nil && s.SetupIntent.PaymentMethod != nil {
		res.PaymentMethodID = s.SetupIntent.PaymentMethod.ID
	}
	return res, nil
}

func (p *Provider) CreatePortalSession(ctx context.Context, customerID, configurationID, returnURL string) (string, error) {
	s, err := p.sc.V1BillingPortalSessions.Create(ctx, &stripego.BillingPortalSessionCreateParams{
		Customer:      stripego.String(customerID),
		Configuration: stripego.String(configurationID),
		ReturnURL:     stripego.String(returnURL),
	})
	if err != nil {
		return "", providerErr("create portal session", err)
	}
	return s.URL, nil
}

func (p *Provider) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string, providerSubscriptionID *string) error {
	_, err := p.sc.V1Customers.Update(ctx, customerID, &stripego.CustomerUpdateParams{
		InvoiceSettings: &stripego.CustomerUpdateInvoiceSettingsParams{
			DefaultPaymentMethod: stripego.String(paymentMethodID),
		},
	})
	if err != nil {
		return providerErr("set customer default payment method", err)
	}
	if providerSubscriptionID == nil {
		return nil
	}
	_, err = p.sc.V1Subscriptions.Update(ctx, *providerSubscriptionID, &stripego.SubscriptionUpdateParams{
		DefaultPaymentMethod: stripego.String(paymentMethodID),
	})
	return providerErr("set subscription default payment method", err)
}

func (p *Provider) DetachPaymentMethods(ctx context.Context, customerID, keepPaymentMethodID string) error {
	list := p.sc.V1Customers.ListPaymentMethods(ctx, &stripego.CustomerListPaymentMethodsParams{
		Customer: stripego.String(customerID),
	})
	for pm, err := range list.All(ctx) {
		if err != nil {
			return providerErr("list payment methods", err)
		}
		if pm.ID == keepPaymentMethodID {
			continue
		}
		if _, err := p.sc.V1PaymentMethods.Detach(ctx, pm.ID, nil); err != nil {
			return providerErr("detach payment method", err)
		}
	}
	return nil
}
