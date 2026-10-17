package stripe

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripego "github.com/stripe/stripe-go/v87"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func providerWithResponses(t *testing.T, responses map[string]string) *Provider {
	t.Helper()
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, ok := responses[r.URL.Path]
		if !ok {
			t.Errorf("unexpected stripe request %s", r.URL.Path)
			body = `{}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
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

func TestPeriodPaymentsWithoutPaidInvoiceIsEmpty(t *testing.T) {
	p := providerWithResponses(t, map[string]string{
		"/v1/invoices": `{"object":"list","data":[],"has_more":false,"url":"/v1/invoices"}`,
	})

	got, err := p.PeriodPayments(context.Background(), "sub_1", time.Now(), time.Now().Add(time.Hour))

	require.NoError(t, err)
	assert.Empty(t, got)
}
