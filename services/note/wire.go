//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/services/note/app"
	components "github.com/notopia-uit/notopia/services/note/component"
	"github.com/notopia-uit/notopia/services/note/config"
	"github.com/notopia-uit/notopia/services/note/infra"
	controller "github.com/notopia-uit/notopia/services/note/transport"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	app.ProviderSet,
	components.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)

func InitializeServer(ctx context.Context) (*Server, func(), error) {
	panic(wire.Build(ProviderSet))
}
