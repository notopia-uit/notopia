package search

import (
	"log/slog"
	"net/http"
)

type MeilisearchLogRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

var _ http.RoundTripper = (*MeilisearchLogRoundTripper)(nil)

func (l *MeilisearchLogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := l.next.RoundTrip(req)
	l.logger.InfoContext(
		req.Context(),
		"Request to Meilisearch API completed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.Any("error", err),
	)

	return resp, err
}
