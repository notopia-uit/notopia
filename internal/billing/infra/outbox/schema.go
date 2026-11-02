package outbox

import "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"

// ForwarderTopic must match in the per-transaction publisher and the forwarder.
const ForwarderTopic = "billing_outbox_forwarder"

// messagesTable is created by pgmigration/00002_outbox.sql.
const messagesTable = "outbox_messages"

func messagesTableName(string) string { return messagesTable }

func NewSchemaAdapter() *sql.PostgreSQLQueueSchema {
	return &sql.PostgreSQLQueueSchema{
		GenerateWhereClause:       nil,
		GeneratePayloadType:       nil,
		GenerateMessagesTableName: messagesTableName,
		SubscribeBatchSize:        0,
	}
}

var ProvideSchemaAdapter = NewSchemaAdapter

func NewOffsetsAdapter() *sql.PostgreSQLQueueOffsetsAdapter {
	return &sql.PostgreSQLQueueOffsetsAdapter{
		DeleteOnAck:               true,
		GenerateMessagesTableName: messagesTableName,
	}
}

var ProvideOffsetsAdapter = NewOffsetsAdapter
