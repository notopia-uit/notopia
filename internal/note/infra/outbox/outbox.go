package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
)

type Publisher message.Publisher

type Outbox struct {
	fwd *forwarder.Forwarder
}

func NewOutbox(
	publisher Publisher,
	logger watermill.LoggerAdapter,
	schemaAdapter sql.SchemaAdapter,
	pgxConn sql.Conn,
) (*Outbox, error) {
	subcriber, err := sql.NewSubscriber(
		sql.BeginnerFromPgx(pgxConn),
		sql.SubscriberConfig{
			PollInterval:     time.Second,
			InitializeSchema: true,
			SchemaAdapter:    schemaAdapter,
			ConsumerGroup:    "", // NOTE: If scale, we should care about this
			OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL subcriber router: %w", err)
	}
	router.AddMiddleware(middleware.Recoverer, wotel.Trace())
	fwd, err := forwarder.NewForwarder(
		subcriber,
		publisher,
		logger,
		forwarder.Config{
			Router: router,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create forwarder: %w", err)
	}
	return &Outbox{
		fwd: fwd,
	}, nil
}

var ProvideOutbox = NewOutbox

func (o *Outbox) Run(ctx context.Context) error {
	if err := o.fwd.Run(ctx); err != nil {
		return fmt.Errorf("failed to run forwarder: %w", err)
	}
	return nil
}

func (o *Outbox) Stop() error {
	if err := o.fwd.Close(); err != nil {
		return fmt.Errorf("failed to close forwarder: %w", err)
	}
	return nil
}
