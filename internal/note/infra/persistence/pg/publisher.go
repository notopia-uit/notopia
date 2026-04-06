package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type Publisher struct {
	publisher message.Publisher
}

func NewPublisher(
	pgxTx pgx.Tx,
	logger watermill.LoggerAdapter,
) (*Publisher, error) {
	sqlPublisher, err := sql.NewPublisher(
		sql.TxFromPgx(pgxTx),
		sql.PublisherConfig{
			SchemaAdapter:        sql.DefaultPostgreSQLSchema{},
			AutoInitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL publisher: %w", err)
	}
	publisher := forwarder.NewPublisher(sqlPublisher, forwarder.PublisherConfig{})
	return &Publisher{
		publisher: publisher,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, events ...domain.Event) error {
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		if err := p.publisher.Publish(string(domain.GetEventType(event)), msg); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}
	return nil
}

type PublisherFactory struct {
	logger watermill.LoggerAdapter
}

func NewPublisherFactory(logger watermill.LoggerAdapter) *PublisherFactory {
	return &PublisherFactory{
		logger: logger,
	}
}

var ProvidePublisherFactory = NewPublisherFactory

func (f *PublisherFactory) Create(pgxTx pgx.Tx) (*Publisher, error) {
	return NewPublisher(pgxTx, f.logger)
}
