package stripe

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripego "github.com/stripe/stripe-go/v87"
)

func TestSnapshotFromSubscription(t *testing.T) {
	raw, err := os.ReadFile("testdata/subscription.json")
	require.NoError(t, err)
	var s stripego.Subscription
	require.NoError(t, json.Unmarshal(raw, &s))

	snap, err := snapshotFromSubscription(&s)
	require.NoError(t, err)
	assert.Equal(t, "price_plus_month", snap.ProviderPriceID)
	assert.Equal(t, domain.StatusActive, snap.Status)
	assert.True(t, snap.CurrentPeriodEnd.Equal(time.Unix(1761955200, 0)))
	assert.Equal(t, "00000000-0000-4000-8000-000000000110", snap.WorkspaceID.String())
}
