package identity

import (
	"log/slog"
	"net/http"
)

type AuthentikLogRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

var _ http.RoundTripper = (*AuthentikLogRoundTripper)(nil)

func (l *AuthentikLogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := l.next.RoundTrip(req)
	l.logger.InfoContext(
		req.Context(),
		"Request to Authentik API completed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.Any("error", err),
	)

	return resp, err
}
