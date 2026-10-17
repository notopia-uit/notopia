package stripe

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripego "github.com/stripe/stripe-go/v87"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func providerWithHandler(t *testing.T, handle func(r *http.Request) (int, string)) *Provider {
	t.Helper()
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		status, body := handle(r)
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})}
	return &Provider{
		sc:       stripego.NewClient("sk_test_x", stripego.WithBackends(stripego.NewBackends(client))),
		currency: "vnd",
	}
}

func providerWithResponses(t *testing.T, responses map[string]string) *Provider {
	t.Helper()
	return providerWithHandler(t, func(r *http.Request) (int, string) {
		body, ok := responses[r.URL.Path]
		if !ok {
			t.Errorf("unexpected stripe request %s", r.URL.Path)
			body = `{}`
		}
		return http.StatusOK, body
	})
}

func TestPeriodPaymentsWithoutPaidInvoiceIsEmpty(t *testing.T) {
	p := providerWithResponses(t, map[string]string{
		"/v1/invoices": `{"object":"list","data":[],"has_more":false,"url":"/v1/invoices"}`,
	})

	got, err := p.PeriodPayments(context.Background(), "sub_1", time.Now(), time.Now().Add(time.Hour))

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetRefundResolvesSubscriptionOfDashboardRefund(t *testing.T) {
	p := providerWithResponses(t, map[string]string{
		"/v1/refunds/re_1": `{"id":"re_1","object":"refund","amount":10000,"currency":"vnd","payment_intent":"pi_1","status":"succeeded","metadata":{}}`,
		"/v1/invoice_payments": `{"object":"list","has_more":false,"url":"/v1/invoice_payments","data":[{
			"id":"inpay_1","object":"invoice_payment","status":"paid",
			"invoice":{"id":"in_1","object":"invoice","parent":{"type":"subscription_details","subscription_details":{"subscription":"sub_9"}}},
			"payment":{"type":"payment_intent","payment_intent":"pi_1"}}]}`,
	})

	res, err := p.GetRefund(context.Background(), "re_1")

	require.NoError(t, err)
	assert.Equal(t, "pi_1", res.PaymentRef)
	assert.Equal(t, "sub_9", res.ProviderSubscriptionID)
}

func TestGetRefundOfOurRefundSkipsSubscriptionLookup(t *testing.T) {
	p := providerWithResponses(t, map[string]string{
		"/v1/refunds/re_1": `{"id":"re_1","object":"refund","amount":10000,"currency":"vnd","payment_intent":"pi_1","status":"succeeded","metadata":{"refund_id":"00000000-0000-4000-8000-000000000001"}}`,
	})

	res, err := p.GetRefund(context.Background(), "re_1")

	require.NoError(t, err)
	require.NotNil(t, res.RefundID)
	assert.Empty(t, res.ProviderSubscriptionID)
}

func TestCreateRefundMapsInvalidRequestToInvalid(t *testing.T) {
	p := providerWithHandler(t, func(*http.Request) (int, string) {
		return http.StatusBadRequest, `{"error":{"type":"invalid_request_error","message":"Refund amount is greater than unrefunded amount on charge"}}`
	})

	_, err := p.CreateRefund(context.Background(), uuid.New(), "pi_1", domain.Money{Amount: 10000, Currency: "vnd"})

	var e errs.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, errs.CodeInvalid, e.Code())
}

func TestCreateRefundKeepsTransientErrorsAsProviderErrors(t *testing.T) {
	p := providerWithHandler(t, func(*http.Request) (int, string) {
		return http.StatusTooManyRequests, `{"error":{"type":"rate_limit_error","message":"slow down"}}`
	})

	_, err := p.CreateRefund(context.Background(), uuid.New(), "pi_1", domain.Money{Amount: 10000, Currency: "vnd"})

	var e errs.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, errs.CodePaymentProvider, e.Code())
}
