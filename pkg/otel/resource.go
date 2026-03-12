package otel

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/metadata"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

func NewResource(
	ctx context.Context,
	serviceName metadata.ServiceName,
	serviceVersion metadata.ServiceVersion,
) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName.String()),
		semconv.ServiceVersion(serviceVersion.String()),
	}

	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithContainer(),
	)
}

var ProvideResource = NewResource
