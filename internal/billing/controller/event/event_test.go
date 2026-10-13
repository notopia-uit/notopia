package event

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/share"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	id := uuid.New()
	var got *share.WorkspaceDeletedIntegrationEvent
	handler := decode(func(_ context.Context, ev *share.WorkspaceDeletedIntegrationEvent) error {
		got = ev
		return nil
	})

	require.NoError(t, handler(message.NewMessage("1", []byte(`{"id":"`+id.String()+`"}`))))
	require.NotNil(t, got)
	assert.Equal(t, id, got.Id)

	got = nil
	require.NoError(t, handler(message.NewMessage("2", []byte(`not json`))), "undecodable payload is acked")
	assert.Nil(t, got)
}
