//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing"
)

var ProviderSet = wire.NewSet(
	billing.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (*billing.Server, func(), error) {
	panic(wire.Build(ProviderSet))
}
