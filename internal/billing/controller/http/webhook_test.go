package http

import (
	"bytes"
	"net/http"
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

func (f *webhookFixture) post(t *testing.T, payload []byte, signature string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	validate, err := ValidateHandler()
	require.NoError(t, err)
	engine := gin.New()
	engine.Use(validate)
	billing.RegisterHandlersWithOptions(engine, NewHandler(f.handler), handlerOptions())

	req := httptest.NewRequest(http.MethodPost, "/billing/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", signature)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func sign(payload []byte) string {
	return webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: testWebhookSecret}).Header
}

const invoicePaidPayload = `{"id":"evt_1","object":"event","api_version":"2020-08-27","type":"invoice.paid",` +
	`"data":{"object":{"id":"in_1","object":"invoice","parent":{"subscription_details":{"subscription":"sub_123"}}}}}`

const invoicePaidPrettyPayload = `{
  "id": "evt_1",
  "object": "event",
  "api_version": "2020-08-27",
  "type": "invoice.paid",
  "data": {
    "object": {
      "id": "in_1",
      "object": "invoice",
      "parent": {
        "subscription_details": {
          "subscription": "sub_123"
        }
      }
    }
  }
}`

func (f *webhookFixture) expectInvoicePaid() {
	f.syncSubscription.EXPECT().
		Handle(mock.Anything, &app.SyncSubscription{EventID: "evt_1", EventType: "invoice.paid", ProviderSubscriptionID: "sub_123"}).
		Return(nil).
		Once()
}

func TestStripeWebhookInvoicePaidSyncsSubscription(t *testing.T) {
	f := newWebhookFixture(t)
	payload := []byte(invoicePaidPayload)
	f.expectInvoicePaid()

	rec := f.post(t, payload, sign(payload))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestStripeWebhookPrettyPrintedBodyVerifies(t *testing.T) {
	f := newWebhookFixture(t)
	payload := []byte(invoicePaidPrettyPayload)
	f.expectInvoicePaid()

	rec := f.post(t, payload, sign(payload))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestStripeWebhookTrailingNewlineBodyVerifies(t *testing.T) {
	f := newWebhookFixture(t)
	payload := []byte(invoicePaidPrettyPayload + "\n")
	f.expectInvoicePaid()

	rec := f.post(t, payload, sign(payload))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestStripeWebhookBadSignature(t *testing.T) {
	f := newWebhookFixture(t)

	rec := f.post(t, []byte(invoicePaidPayload), "t=1,v1=bad")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), errs.CodeWebhookSignature.String())
}
