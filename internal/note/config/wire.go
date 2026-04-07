package config

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(
		new(*Config),
		"Advanced",
		"Database",
		"General",
		"Kafka",
		"Log",
		"Redis",
		"Server",
		"Services",
	),
	wire.FieldsOf(
		new(*Advanced),
		"DomainEvent",
		"WorkspaceEvent",
	),
)
