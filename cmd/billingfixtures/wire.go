//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billingfixtures"
)

var ProviderSet = wire.NewSet(
	billingfixtures.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeFixtures(ctx context.Context) (*billingfixtures.Fixtures, func(), error) {
	panic(wire.Build(ProviderSet))
}
