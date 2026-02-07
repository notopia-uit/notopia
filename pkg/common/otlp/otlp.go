package otlp

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// NOTE: I don't know, AI said me to do so
// Traces won't "link up" between your client and server unless they both know how to read/write the traceparent header.
// Even without global providers, you usually want to set the Global Propagator so the interceptor knows
// how to inject/extract the data from HTTP headers.
func init() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}
