package otel

import (
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/propagation"
)

func NewTextMapPropagator() propagation.TextMapPropagator {
	return autoprop.NewTextMapPropagator()
}

var ProvideTextMapPropagator = NewTextMapPropagator
