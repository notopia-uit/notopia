//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorizationfixtures"
)

var ProviderSet = wire.NewSet(
	authorizationfixtures.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeFixtures(ctx context.Context) (*authorizationfixtures.Fixtures, func(), error) {
	panic(wire.Build(ProviderSet))
}
