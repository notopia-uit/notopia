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
		"Authentik",
		"Database",
		"General",
		"Kafka",
		"Log",
		"Meilisearch",
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
