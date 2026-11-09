package config

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideViper,
	wire.FieldsOf(
		new(*Config),
		"General",
		"Log",
		"Server",
		"Database",
		"Kafka",
	),
)

var SeedProviderSet = wire.NewSet(ProvideViper, ProvideSeedConfig, wire.FieldsOf(new(*SeedConfig), "General", "Log", "Database"))
