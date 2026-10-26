package pgjsonb

import (
	"encoding/json"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type Features map[domain.FeatureKey]bool

func (f Features) MarshalJSON() ([]byte, error) { return marshalKeyed(f) }

func (f *Features) UnmarshalJSON(raw []byte) (err error) {
	*f, err = unmarshalKeyed[domain.FeatureKey, bool](raw, domain.ParseFeatureKey)
	return err
}

type Limits map[domain.LimitKey]int64

func (l Limits) MarshalJSON() ([]byte, error) { return marshalKeyed(l) }

func (l *Limits) UnmarshalJSON(raw []byte) (err error) {
	*l, err = unmarshalKeyed[domain.LimitKey, int64](raw, domain.ParseLimitKey)
	return err
}

type stringKey interface {
	comparable
	String() string
}

func marshalKeyed[K stringKey, V any](m map[K]V) ([]byte, error) {
	out := make(map[string]V, len(m))
	for k, v := range m {
		out[k.String()] = v
	}
	return json.Marshal(out)
}

func unmarshalKeyed[K comparable, V any](raw []byte, parse func(string) (K, error)) (map[K]V, error) {
	var in map[string]V
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	out := make(map[K]V, len(in))
	for s, v := range in {
		k, err := parse(s)
		if err != nil {
			slog.Warn("skipping unknown plan key", slog.String("key", s))
			continue
		}
		out[k] = v
	}
	return out, nil
}
