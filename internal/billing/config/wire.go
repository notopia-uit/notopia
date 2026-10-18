package config

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(new(*Config), "Billing", "Cron", "Database", "General", "Identity", "Kafka", "Log", "Redis", "Server", "Services", "Stripe"),
)
