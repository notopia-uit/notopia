// Made by AI
package casbin

import (
	"context"
	"log/slog"
	"os"

	"github.com/casbin/casbin/v3/log"
)

// SlogLogger is a Logger implementation that uses the standard library log/slog package.
// It provides structured logging with support for various slog.Handler types.
type SlogLogger struct {
	logger      *slog.Logger
	eventTypes  map[log.EventType]bool
	logCallback func(entry *log.LogEntry) error
}

var _ log.Logger = (*SlogLogger)(nil)

// NewSlogLogger creates a new SlogLogger instance.
// If logger is nil, a default JSON logger to os.Stdout is created.
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	if logger == nil {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger = slog.New(handler)
	}
	return &SlogLogger{
		logger:      logger,
		eventTypes:  make(map[log.EventType]bool),
		logCallback: nil,
	}
}

// NewSlogLoggerWithHandler creates a new SlogLogger with a custom slog.Handler.
func NewSlogLoggerWithHandler(handler slog.Handler) *SlogLogger {
	return NewSlogLogger(slog.New(handler))
}

// WithHandler sets a custom slog.Handler for this logger (builder pattern).
func (l *SlogLogger) WithHandler(handler slog.Handler) *SlogLogger {
	l.logger = slog.New(handler)
	return l
}

// WithSlogLogger sets the underlying slog.Logger directly (builder pattern).
func (l *SlogLogger) WithSlogLogger(logger *slog.Logger) *SlogLogger {
	if logger != nil {
		l.logger = logger
	}
	return l
}

// WithLogCallback sets a user-provided callback function (builder pattern).
func (l *SlogLogger) WithLogCallback(callback func(entry *log.LogEntry) error) *SlogLogger {
	l.logCallback = callback
	return l
}

// SetEventTypes sets the event types that should be logged.
// Only events matching these types will have IsActive set to true.
func (l *SlogLogger) SetEventTypes(eventTypes []log.EventType) error {
	l.eventTypes = make(map[log.EventType]bool)
	for _, et := range eventTypes {
		l.eventTypes[et] = true
	}
	return nil
}

// OnBeforeEvent is called before an event occurs.
// It sets the StartTime and determines if the event should be active based on configured event types.
func (l *SlogLogger) OnBeforeEvent(entry *log.LogEntry) error {
	if entry == nil {
		return nil
	}

	// Set IsActive based on whether this event type is enabled
	// If no event types are configured, all events are considered active
	if len(l.eventTypes) == 0 {
		entry.IsActive = true
	} else {
		entry.IsActive = l.eventTypes[entry.EventType]
	}

	return nil
}

// OnAfterEvent is called after an event completes.
// It logs the entry via slog if active, and calls the user callback if set.
func (l *SlogLogger) OnAfterEvent(entry *log.LogEntry) error {
	if entry == nil {
		return nil
	}

	// Only log if the event is active
	if entry.IsActive && l.logger != nil {
		l.logEntry(entry)
	}

	// Call user-provided callback if set
	if l.logCallback != nil {
		if err := l.logCallback(entry); err != nil {
			return err
		}
	}

	return nil
}

// SetLogCallback sets a user-provided callback function.
// The callback is called at the end of OnAfterEvent.
func (l *SlogLogger) SetLogCallback(callback func(entry *log.LogEntry) error) error {
	l.logCallback = callback
	return nil
}

// logEntry converts a LogEntry to slog structured logging output.
func (l *SlogLogger) logEntry(entry *log.LogEntry) {
	if l.logger == nil {
		return
	}

	// Determine log level based on event type and error status
	level := slog.LevelInfo
	if entry.Error != nil {
		level = slog.LevelError
	}

	// Build attributes based on event type
	attrs := []slog.Attr{
		slog.String("event_type", string(entry.EventType)),
		slog.Duration("duration", entry.Duration),
	}

	switch entry.EventType {
	case log.EventEnforce:
		attrs = append(
			attrs,
			slog.String("subject", entry.Subject),
			slog.String("object", entry.Object),
			slog.String("action", entry.Action),
			slog.String("domain", entry.Domain),
			slog.Bool("allowed", entry.Allowed),
		)

	case log.EventAddPolicy, log.EventRemovePolicy:
		attrs = append(
			attrs,
			slog.Int("rule_count", entry.RuleCount),
		)
		if len(entry.Rules) > 0 {
			attrs = append(attrs, slog.Any("rules", entry.Rules))
		}

	case log.EventLoadPolicy, log.EventSavePolicy:
		attrs = append(
			attrs,
			slog.Int("rule_count", entry.RuleCount),
		)
	}

	// Add error if present
	if entry.Error != nil {
		attrs = append(attrs, slog.String("error", entry.Error.Error()))
	}

	// Log with appropriate message and level
	msg := l.getEventMessage(entry.EventType)
	if level == slog.LevelError {
		//nolint:sloglint
		l.logger.LogAttrs(context.Background(), level, msg, attrs...)
	} else {
		//nolint:sloglint
		l.logger.LogAttrs(context.Background(), level, msg, attrs...)
	}
}

// getEventMessage returns a descriptive message for each event type.
func (l *SlogLogger) getEventMessage(eventType log.EventType) string {
	switch eventType {
	case log.EventEnforce:
		return "Enforce decision"
	case log.EventAddPolicy:
		return "Policy added"
	case log.EventRemovePolicy:
		return "Policy removed"
	case log.EventLoadPolicy:
		return "Policy loaded"
	case log.EventSavePolicy:
		return "Policy saved"
	default:
		return "Casbin event"
	}
}
