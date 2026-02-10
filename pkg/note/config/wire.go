package config

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/config"
)

var ProviderSet = wire.NewSet(
	ProvideViper,
	Provide,
	wire.FieldsOf(
		new(*Config),
		"Server",
		"Database",
		"OTLP",
	),
	config.ProvideSet,
)
