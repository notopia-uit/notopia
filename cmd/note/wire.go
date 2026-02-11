//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note"
)

var ProviderSet = wire.NewSet(
	note.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (*note.Server, func(), error) {
	panic(wire.Build(ProviderSet))
}
