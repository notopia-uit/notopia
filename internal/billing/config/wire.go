package config

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(new(*Config), "Billing", "Cron", "Database", "DomainEvent", "General", "Identity", "Kafka", "Log", "Redis", "Server", "Services", "Stripe"),
)

var (
	MigrateProviderSet = wire.NewSet(ProvideViper, ProvideMigrateConfig, wire.FieldsOf(new(*MigrateConfig), "General", "Log", "Database"))
	SeedProviderSet    = wire.NewSet(ProvideViper, ProvideSeedConfig, wire.FieldsOf(new(*SeedConfig), "General", "Log", "Database", "Identity", "Stripe", "Billing"))
)
