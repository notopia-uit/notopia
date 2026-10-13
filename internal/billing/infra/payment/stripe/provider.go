package stripe

import (
	"errors"
	"net/http"

	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	stripego "github.com/stripe/stripe-go/v87"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

type Provider struct {
	sc       *stripego.Client
	currency string
}

var _ app.PaymentProvider = (*Provider)(nil)

// stripe-go forwards ctx to every request, so the otelhttp spans nest under the use case span.
func New(stripeCfg *config.Stripe, billingCfg *config.Billing) *Provider {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanOptions(trace.WithAttributes(
				semconv.ServicePeerName("stripe"),
			)),
		),
	}
	return &Provider{
		sc: stripego.NewClient(
			stripeCfg.SecretKey.Reveal(),
			stripego.WithBackends(stripego.NewBackends(httpClient)),
		),
		currency: billingCfg.Currency,
	}
}

var Provide = New

func providerErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return errs.New(errs.CodePaymentProvider, "stripe "+op+" failed", err)
}

func isStripeCode(err error, code stripego.ErrorCode) bool {
	var se *stripego.Error
	return errors.As(err, &se) && se.Code == code
}

func productID(tier string) string {
	return "notopia_" + tier
}
