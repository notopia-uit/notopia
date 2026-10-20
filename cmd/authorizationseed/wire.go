//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorizationseed"
)

var ProviderSet = wire.NewSet(
	authorizationseed.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeSeeder(ctx context.Context) (*authorizationseed.Seeder, func(), error) {
	panic(wire.Build(ProviderSet))
}
