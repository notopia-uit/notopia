//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billingseed"
)

var ProviderSet = wire.NewSet(
	billingseed.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeSeeder(ctx context.Context) (*billingseed.Seeder, func(), error) {
	panic(wire.Build(ProviderSet))
}
