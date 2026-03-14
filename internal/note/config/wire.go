package config

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(
		new(*Config),
		"Database",
		"General",
		"Log",
		"Server",
		"Kafka",
	),
)
