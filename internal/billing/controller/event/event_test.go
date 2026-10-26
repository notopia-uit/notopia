package event

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
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

func TestDecodeAcksPermanentErrors(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"not found", errs.NewNotFound("x"), false},
		{"invalid", errs.NewInvalid("x"), false},
		{"wrapped not found", fmt.Errorf("wrap: %w", errs.NewNotFound("x")), false},
		{"provider", errs.New(errs.CodePaymentProvider, "x", nil), true},
		{"plain", errors.New("boom"), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			handler := decode(func(_ context.Context, _ *share.WorkspaceDeletedIntegrationEvent) error { return c.err })
			err := handler(message.NewMessage("1", []byte(`{}`)))
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
