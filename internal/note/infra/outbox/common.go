package outbox

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/notopia-uit/notopia/internal/note/config"
)

func NewSchemaAdapter(
	cfg *config.DomainEvent,
) *sql.DefaultPostgreSQLSchema {
	return &sql.DefaultPostgreSQLSchema{
		GenerateMessagesTableName: func(_ string) string {
			return cfg.OutboxTableName
		},
	}
}

var ProvideSchemaAdapter = NewSchemaAdapter
