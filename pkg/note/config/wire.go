package config

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(
	ProvideConfig,
	wire.FieldsOf(
		new(*Config),
		"Server",
		"Database",
		"OTLP",
	),
	config.ProvideSet,
	viper.New,
)
