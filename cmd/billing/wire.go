//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing"
	"github.com/notopia-uit/notopia/pkg/common/cli"
)

var ProviderSet = wire.NewSet(
	billing.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (cli.Runner, func(), error) {
	panic(wire.Build(ProviderSet, wire.Bind(new(cli.Runner), new(*billing.Server))))
}
func InitializeSeeder(ctx context.Context) (cli.Seeder, func(), error) {
	panic(wire.Build(billing.SeedProviderSet, wire.Value(ServiceName), wire.Value(ServiceVersion)))
}
func InitializeMigrator(ctx context.Context) (cli.Migrator, func(), error) {
	panic(wire.Build(billing.MigrateProviderSet, wire.Value(ServiceName), wire.Value(ServiceVersion)))
}
