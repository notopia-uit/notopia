//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization"
)

var ProviderSet = wire.NewSet(
	authorization.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (*authorization.Server, func(), error) {
	panic(wire.Build(ProviderSet))
}
