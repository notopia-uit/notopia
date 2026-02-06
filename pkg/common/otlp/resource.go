package otlp

import (
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

func NewResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	}

	podName := os.Getenv("POD_NAME")
	if podName != "" && podName != "unknown" && podName != "unknown_pod" {
		attrs = append(attrs, semconv.K8SPodNameKey.String(podName))
	}

	podNamespace := os.Getenv("POD_NAMESPACE")
	if podNamespace != "" && podNamespace != "unknown" {
		attrs = append(attrs, semconv.K8SNamespaceNameKey.String(podNamespace))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}
