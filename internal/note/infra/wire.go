package infra

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/infra/common"
	"github.com/notopia-uit/notopia/internal/note/infra/identity"
	"github.com/notopia-uit/notopia/internal/note/infra/integrationpublisher"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/service"
	"github.com/notopia-uit/notopia/internal/note/infra/workspaceevent"
)

var ProviderSet = wire.NewSet(
	wire.Bind(new(sql.Conn), new(*pgxpool.Pool)),

	common.ProviderSet,
	identity.ProviderSet,
	integrationpublisher.ProviderSet,
	outbox.ProviderSet,
	persistence.PostgresProviderSet,
	service.ProviderSet,
	workspaceevent.ProviderSet,
)
