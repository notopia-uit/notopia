package stripe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func TestPeriodPaymentsAfterMonthToYearUpgradeIncludesTheOriginalMonthlyInvoice(t *testing.T) {
	day := 24 * time.Hour
	periodStart := time.Unix(1_800_000_000, 0)
	periodEnd := periodStart.Add(365 * day)
	invoice := func(id string, created time.Time, covStart, covEnd time.Time) string {
		return fmt.Sprintf(
			`{"id":%q,"object":"invoice","created":%d,"lines":{"object":"list","has_more":false,"data":[{"id":"il_%s","object":"line_item","period":{"start":%d,"end":%d}}]}}`,
			id, created.Unix(), id, covStart.Unix(), covEnd.Unix(),
		)
	}
	invoices := strings.Join([]string{
		invoice("in_y", periodStart, periodStart, periodEnd),
		invoice("in_m", periodStart.Add(-15*day), periodStart.Add(-15*day), periodStart.Add(15*day)),
		invoice("in_prev", periodStart.Add(-45*day), periodStart.Add(-45*day), periodStart.Add(-15*day)),
		invoice("in_old", periodStart.Add(-30*day), periodStart.Add(-30*day), periodStart),
	}, ",")
	var createdGTE int64
	p := providerWithHandler(t, func(r *http.Request) (int, string) {
		switch r.URL.Path {
		case "/v1/invoices":
			createdGTE, _ = strconv.ParseInt(r.URL.Query().Get("created[gte]"), 10, 64)
			return http.StatusOK, `{"object":"list","has_more":false,"url":"/v1/invoices","data":[` + invoices + `]}`
		case "/v1/invoice_payments":
			pi := "pi_" + strings.TrimPrefix(r.URL.Query().Get("invoice"), "in_")
			return http.StatusOK, `{"object":"list","has_more":false,"url":"/v1/invoice_payments","data":[{"id":"inpay","object":"invoice_payment","status":"paid","amount_paid":1000,"currency":"vnd","payment":{"type":"payment_intent","payment_intent":"` + pi + `"}}]}`
		}
		t.Errorf("unexpected stripe request %s", r.URL.Path)
		return http.StatusOK, `{}`
	})

	got, err := p.PeriodPayments(context.Background(), "sub_1", periodStart, periodEnd)

	require.NoError(t, err)
	refs := make([]string, 0, len(got))
	for _, pay := range got {
		refs = append(refs, pay.PaymentRef)
	}
	assert.ElementsMatch(t, []string{"pi_y", "pi_m"}, refs)
	assert.LessOrEqual(t, createdGTE, periodStart.Add(-365*day).Unix(), "the created window reaches the longest billing interval back")
}
