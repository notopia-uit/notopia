package components

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/otlp"
)

const (
	// TODO: Let something track this version?
	ServiceName    otlp.ServiceName    = "notopia-note"
	ServiceVersion otlp.ServiceVersion = "v0.0.0"
)

var ProviderSet = wire.NewSet(
	ProvideValidate,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)
