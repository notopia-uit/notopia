//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization"
	"github.com/notopia-uit/notopia/pkg/common/cli"
)

var ProviderSet = wire.NewSet(
	authorization.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (cli.Runner, func(), error) {
	panic(wire.Build(ProviderSet, wire.Bind(new(cli.Runner), new(*authorization.Server))))
}
func InitializeSeeder(ctx context.Context) (cli.Seeder, func(), error) {
	panic(wire.Build(authorization.SeedProviderSet, wire.Value(ServiceName), wire.Value(ServiceVersion)))
}
