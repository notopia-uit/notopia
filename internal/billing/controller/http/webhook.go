package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/webhook"
)

const stripeWebhookPath = "/billing/webhooks/stripe"

type rawBodyKey struct{}

func handlerOptions() billing.GinServerOptions {
	return billing.GinServerOptions{
		BaseURL:      "",
		Middlewares:  []billing.MiddlewareFunc{captureWebhookRawBody},
		ErrorHandler: nil,
	}
}

func captureWebhookRawBody(c *gin.Context) {
	if c.FullPath() != stripeWebhookPath {
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.Set(rawBodyKey{}, body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
}

func rawBody(ctx context.Context) []byte {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil
	}
	v, _ := c.Get(rawBodyKey{})
	body, _ := v.([]byte)
	return body
}

func (h *StrictHandler) StripeWebhook(ctx context.Context, req billing.StripeWebhookRequestObject) (billing.StripeWebhookResponseObject, error) {
	body := rawBody(ctx)
	if body == nil {
		return nil, errs.NewInvalid("missing body")
	}
	ev, err := webhook.ConstructEventWithOptions(
		body,
		req.Params.StripeSignature,
		h.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		return nil, errs.New(errs.CodeWebhookSignature, "invalid webhook signature", err)
	}

	eventType := string(ev.Type)
	switch ev.Type { //nolint:exhaustive // unhandled event types are acknowledged and ignored
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := decodeEventObject(ev, &session); err != nil {
			return nil, err
		}
		err = h.App.Cmds.CompleteCheckout.Handle(ctx, &app.CompleteCheckout{EventID: ev.ID, EventType: eventType, SessionID: session.ID})
	case stripe.EventTypeCustomerSubscriptionCreated,
		stripe.EventTypeCustomerSubscriptionUpdated,
		stripe.EventTypeCustomerSubscriptionDeleted:
		var sub stripe.Subscription
		if err := decodeEventObject(ev, &sub); err != nil {
			return nil, err
		}
		err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{EventID: ev.ID, EventType: eventType, ProviderSubscriptionID: sub.ID})
	case stripe.EventTypeInvoicePaid, stripe.EventTypeInvoicePaymentFailed:
		var inv stripe.Invoice
		if err := decodeEventObject(ev, &inv); err != nil {
			return nil, err
		}
		if inv.Parent == nil || inv.Parent.SubscriptionDetails == nil || inv.Parent.SubscriptionDetails.Subscription == nil {
			break
		}
		err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{
			EventID:                ev.ID,
			EventType:              eventType,
			ProviderSubscriptionID: inv.Parent.SubscriptionDetails.Subscription.ID,
		})
	case stripe.EventTypeRefundCreated, stripe.EventTypeRefundUpdated, stripe.EventTypeRefundFailed:
		var refund stripe.Refund
		if err := decodeEventObject(ev, &refund); err != nil {
			return nil, err
		}
		err = h.App.Cmds.SyncRefund.Handle(ctx, &app.SyncRefund{EventID: ev.ID, EventType: eventType, ProviderRefundID: refund.ID})
	}
	if err != nil {
		return nil, err
	}
	return billing.StripeWebhook204Response{}, nil
}

func decodeEventObject(ev stripe.Event, into any) error {
	if err := json.Unmarshal(ev.Data.Raw, into); err != nil {
		return errs.NewInvalid("invalid webhook object")
	}
	return nil
}
