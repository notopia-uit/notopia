package search

import (
	"log/slog"
	"net/http"
	"time"
)

type MeilisearchLogRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

var _ http.RoundTripper = (*MeilisearchLogRoundTripper)(nil)

func (l *MeilisearchLogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := l.next.RoundTrip(req)
	duration := time.Since(start)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	l.logger.InfoContext(
		req.Context(),
		"Request to Meilisearch API completed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", statusCode),
		slog.Duration("latency", duration),
		slog.Any("error", err),
	)

	return resp, err
}
