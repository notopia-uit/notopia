package commonconfig

import (
	"fmt"
	"log/slog"
)

const redacted = "[REDACTED]"

// Secret never prints its value through slog, fmt or encoding/json.
// Use Reveal when the raw value is really needed.
type Secret string

var (
	_ slog.LogValuer = Secret("")
	_ fmt.Stringer   = Secret("")
	_ fmt.GoStringer = Secret("")
	_ fmt.Formatter  = Secret("")
)

func (s Secret) Reveal() string {
	return string(s)
}

func (s Secret) String() string {
	if s == "" {
		return ""
	}
	return redacted
}

func (s Secret) GoString() string {
	return fmt.Sprintf("%q", s.String())
}

func (s Secret) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

func (s Secret) Format(f fmt.State, _ rune) {
	_, _ = fmt.Fprint(f, s.String())
}

func (s Secret) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
