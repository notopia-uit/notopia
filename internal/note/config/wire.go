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
		"Cron",
		"Database",
		"General",
		"Identity",
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

var MigrateProviderSet = wire.NewSet(ProvideViper, ProvideMigrateConfig, wire.FieldsOf(new(*MigrateConfig), "General", "Log", "Database"))
