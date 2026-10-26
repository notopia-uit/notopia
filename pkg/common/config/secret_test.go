package commonconfig_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/stretchr/testify/assert"
)

const rawSecret = "sk_test_super_secret"

type secretHolder struct {
	Name string               `json:"name"`
	S    commonconfig.Secret  `json:"s"`
	Ptr  *commonconfig.Secret `json:"ptr"`
	Nest struct{ T commonconfig.Secret }
}

func newHolder() secretHolder {
	s := commonconfig.Secret(rawSecret)
	h := secretHolder{Name: "n", S: s, Ptr: &s}
	h.Nest.T = s
	return h
}

func TestSecretNeverLeaks(t *testing.T) {
	h := newHolder()

	t.Run("slog text", func(t *testing.T) {
		var buf bytes.Buffer
		slog.New(slog.NewTextHandler(&buf, nil)).Info("cfg", slog.Any("config", h), slog.Any("direct", h.S))
		assert.NotContains(t, buf.String(), rawSecret)
		assert.Contains(t, buf.String(), "[REDACTED]")
	})

	t.Run("slog json", func(t *testing.T) {
		var buf bytes.Buffer
		slog.New(slog.NewJSONHandler(&buf, nil)).Info("cfg", slog.Any("config", h), slog.Any("direct", h.S))
		assert.NotContains(t, buf.String(), rawSecret)
		assert.Contains(t, buf.String(), "[REDACTED]")
	})

	t.Run("fmt verbs", func(t *testing.T) {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
			out := fmt.Sprintf(verb, h)
			assert.NotContains(t, out, rawSecret, verb)
			assert.NotContains(t, fmt.Sprintf(verb, h.S), rawSecret, verb)
		}
		assert.NotContains(t, fmt.Sprintf("%+v", struct{ S commonconfig.Secret }{h.S}), rawSecret)
	})

	t.Run("json", func(t *testing.T) {
		b, err := json.Marshal(h)
		assert.NoError(t, err)
		assert.NotContains(t, string(b), rawSecret)
	})
}

func TestSecretEmptyAndReveal(t *testing.T) {
	assert.Empty(t, commonconfig.Secret("").String())
	assert.Equal(t, rawSecret, commonconfig.Secret(rawSecret).Reveal())
}
