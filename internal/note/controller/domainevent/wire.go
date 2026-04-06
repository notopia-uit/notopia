package domainevent

import "github.com/goforj/wire"

// Currently we let infra provide the pgx connection, which is not really... clear
var ProviderSet = wire.NewSet(
	ProvideDomainEvent,
)
