package infra

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/infra/common"
	"github.com/notopia-uit/notopia/internal/note/infra/domainevent"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/pubsub"
	"github.com/notopia-uit/notopia/internal/note/infra/service"
)

var ProviderSet = wire.NewSet(
	wire.Bind(new(sql.Conn), new(*pgxpool.Pool)),

	common.ProviderSet,
	domainevent.ProviderSet,
	persistence.PostgresProviderSet,
	pubsub.ProviderSet,
	service.ProviderSet,
)
