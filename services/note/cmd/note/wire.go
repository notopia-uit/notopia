//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/services/note/internal"
)

var ProviderSet = wire.NewSet(
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
}

func InitializeServer(ctx context.Context) (*internal.Server, func(), error) {
	panic(wire.Build(internal.ProviderSet))
}
