package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v87/webhook"
)

const testWebhookSecret = "whsec_test"

type webhookFixture struct {
	handler          *StrictHandler
	completeCheckout *app.MockCompleteCheckoutCmd
	syncSubscription *app.MockSyncSubscriptionCmd
	syncRefund       *app.MockSyncRefundCmd
}

func newWebhookFixture(t *testing.T) *webhookFixture {
	t.Helper()
	f := &webhookFixture{
		handler:          nil,
		completeCheckout: app.NewMockCompleteCheckoutCmd(t),
		syncSubscription: app.NewMockSyncSubscriptionCmd(t),
		syncRefund:       app.NewMockSyncRefundCmd(t),
	}
	f.handler = &StrictHandler{
		App: &app.Server{
			Cmds: &app.Cmds{
				CompleteCheckout: f.completeCheckout,
				SyncSubscription: f.syncSubscription,
				SyncRefund:       f.syncRefund,
			},
			Queries: nil,
		},
		Identity:      nil,
		webhookSecret: testWebhookSecret,
	}
	return f
}

func webhookContext(payload []byte) context.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(rawBodyKey, payload)
	return c
}

const invoicePaidPayload = `{"id":"evt_1","object":"event","api_version":"2020-08-27","type":"invoice.paid",` +
	`"data":{"object":{"id":"in_1","object":"invoice","parent":{"subscription_details":{"subscription":"sub_123"}}}}}`

func TestStripeWebhookInvoicePaidSyncsSubscription(t *testing.T) {
	f := newWebhookFixture(t)
	payload := []byte(invoicePaidPayload)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: testWebhookSecret})
	f.syncSubscription.EXPECT().
		Handle(mock.Anything, &app.SyncSubscription{EventID: "evt_1", EventType: "invoice.paid", ProviderSubscriptionID: "sub_123"}).
		Return(nil).
		Once()

	res, err := f.handler.StripeWebhook(webhookContext(payload), billing.StripeWebhookRequestObject{
		Params: billing.StripeWebhookParams{StripeSignature: signed.Header},
	})

	require.NoError(t, err)
	assert.IsType(t, billing.StripeWebhook204Response{}, res)
}

func TestStripeWebhookBadSignature(t *testing.T) {
	f := newWebhookFixture(t)
	payload := []byte(invoicePaidPayload)

	_, err := f.handler.StripeWebhook(webhookContext(payload), billing.StripeWebhookRequestObject{
		Params: billing.StripeWebhookParams{StripeSignature: "t=1,v1=bad"},
	})

	var e errs.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, errs.CodeWebhookSignature, e.Code())
}
