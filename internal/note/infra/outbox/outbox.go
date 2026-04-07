package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher message.Publisher

type Outbox struct {
	fwd *forwarder.Forwarder
}

func NewOutbox(
	publisher Publisher,
	logger watermill.LoggerAdapter,
	pgxConn sql.Conn,
) (*Outbox, error) {
	subcriber, err := sql.NewSubscriber(
		sql.BeginnerFromPgx(pgxConn),
		sql.SubscriberConfig{
			PollInterval:     time.Second,
			InitializeSchema: true,
			SchemaAdapter:    sql.DefaultPostgreSQLSchema{},
			ConsumerGroup:    "", // NOTE: If scale, we should care about this
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL subscriber: %w", err)
	}
	fwd, err := forwarder.NewForwarder(subcriber, publisher, logger, forwarder.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create forwarder: %w", err)
	}
	return &Outbox{
		fwd: fwd,
	}, nil
}

func (o *Outbox) Start(ctx context.Context) error {
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
