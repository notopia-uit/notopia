package http

import (
	"context"
	"encoding/json"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/stripe/stripe-go/v87/webhook"
)

type stripeObject struct {
	ID     string `json:"id"`
	Parent *struct {
		SubscriptionDetails *struct {
			Subscription json.RawMessage `json:"subscription"`
		} `json:"subscription_details"`
	} `json:"parent"`
}

// stripeID reads an expandable field: either "id" or {"id": ...}.
func stripeID(raw json.RawMessage) (string, error) {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, nil
	}
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", errs.NewInvalid("invalid expandable id in webhook object")
	}
	return obj.ID, nil
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
	var obj stripeObject
	if err := json.Unmarshal(ev.Data.Raw, &obj); err != nil {
		return nil, errs.NewInvalid("invalid webhook object")
	}

	switch string(ev.Type) {
	case "checkout.session.completed":
		err = h.App.Cmds.CompleteCheckout.Handle(ctx, &app.CompleteCheckout{EventID: ev.ID, SessionID: obj.ID})
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{EventID: ev.ID, EventType: string(ev.Type), ProviderSubscriptionID: obj.ID})
	case "invoice.paid", "invoice.payment_failed":
		if obj.Parent == nil || obj.Parent.SubscriptionDetails == nil {
			break
		}
		var sub string
		if sub, err = stripeID(obj.Parent.SubscriptionDetails.Subscription); err == nil && sub != "" {
			err = h.App.Cmds.SyncSubscription.Handle(ctx, &app.SyncSubscription{EventID: ev.ID, EventType: string(ev.Type), ProviderSubscriptionID: sub})
		}
	case "refund.created", "refund.updated", "refund.failed":
		err = h.App.Cmds.SyncRefund.Handle(ctx, &app.SyncRefund{EventID: ev.ID, EventType: string(ev.Type), ProviderRefundID: obj.ID})
	}
	if err != nil {
		return nil, err
	}
	return billing.StripeWebhook204Response{}, nil
}
