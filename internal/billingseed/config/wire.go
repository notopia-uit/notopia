package config

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideValidate,
	ProvideViper,
	Provide,
	wire.FieldsOf(
		new(*Config),
		"Catalog",
		"Database",
		"General",
		"Identity",
		"Log",
		"Stripe",
	),
)
