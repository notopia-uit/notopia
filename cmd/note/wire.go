//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note"
	"github.com/notopia-uit/notopia/pkg/common/cli"
)

var ProviderSet = wire.NewSet(
	note.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (cli.Runner, func(), error) {
	panic(wire.Build(ProviderSet, wire.Bind(new(cli.Runner), new(*note.Server))))
}
func InitializeMigrator(ctx context.Context) (cli.Migrator, func(), error) {
	panic(wire.Build(note.MigrateProviderSet, wire.Value(ServiceName), wire.Value(ServiceVersion)))
}
