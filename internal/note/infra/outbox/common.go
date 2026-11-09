package outbox

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/notopia-uit/notopia/internal/note/config"
)

const ForwarderTopic = "note_outbox_forwarder"

func NewSchemaAdapter(
	cfg *config.DomainEvent,
) *sql.PostgreSQLQueueSchema {
	return &sql.PostgreSQLQueueSchema{
		GenerateMessagesTableName: func(_ string) string {
			return cfg.OutboxTableName
		},
	}
}

var ProvideSchemaAdapter = NewSchemaAdapter

func NewOffsetsAdapter(
	cfg *config.DomainEvent,
) *sql.PostgreSQLQueueOffsetsAdapter {
	return &sql.PostgreSQLQueueOffsetsAdapter{
		DeleteOnAck: true,
		GenerateMessagesTableName: func(_ string) string {
			return cfg.OutboxTableName
		},
	}
}

var ProvideOffsetsAdapter = NewOffsetsAdapter
