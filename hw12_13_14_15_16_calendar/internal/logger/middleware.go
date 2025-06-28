package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// HandlerMiddleware adds context fields to every log record.
type HandlerMiddleware struct {
	next slog.Handler
}

// NewHandlerMiddleware wraps the given handler with additional context.
func NewHandlerMiddleware(next slog.Handler) *HandlerMiddleware {
	return &HandlerMiddleware{
		next: next,
	}
}

// Enabled forwards the Enabled check to the next handler.
func (h *HandlerMiddleware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

// Handle enriches the record with context information before logging.
func (h *HandlerMiddleware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(key).(logCtx); ok {
		if c.EventID != uuid.Nil {
			rec.Add("eventID", c.EventID.String())
		}
		if c.Method != "" {
			rec.Add("method", c.Method)
		}
		if !c.Start.IsZero() {
			rec.Add("start", c.Start.Format(time.RFC3339))
		}
	}
	return h.next.Handle(ctx, rec)
}

// WithAttrs returns a new handler with additional attributes.
func (h *HandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddleware{
		next: h.next.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the given group name.
func (h *HandlerMiddleware) WithGroup(name string) slog.Handler {
	return &HandlerMiddleware{
		next: h.next.WithGroup(name),
	}
}

type logCtx struct {
	EventID uuid.UUID
	Method  string
	Start   time.Time
}

type keyType int

const key = keyType(0)

// WithLogEventID attaches an event ID to the logging context.
func WithLogEventID(ctx context.Context, eventID uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.EventID = eventID
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		EventID: eventID,
	})
}

// WithLogMethod attaches a method name to the logging context.
func WithLogMethod(ctx context.Context, method string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Method = method
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		Method: method,
	})
}

// WithLogStart adds a start time to the logging context.
func WithLogStart(ctx context.Context, start time.Time) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Start = start
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		Start: start,
	})
}

type errorWithCtx struct {
	next error
	ctx  logCtx
}

func (e *errorWithCtx) Unwrap() error {
	return e.next
}

func (e *errorWithCtx) Error() string {
	return e.next.Error()
}

// WrapError stores context fields in the returned error.
func WrapError(ctx context.Context, err error) error {
	var prefix string

	// Extract component from context
	if compVal := ctx.Value("component"); compVal != nil {
		if component, ok := compVal.(string); ok && component != "" {
			prefix = component
		}
	}

	c := logCtx{}
	// Extract method from context and append to prefix if present
	if x, ok := ctx.Value(key).(logCtx); ok {
		c = x
		if c.Method != "" {
			if prefix != "" {
				prefix += "." + c.Method
			} else {
				prefix = c.Method
			}
		}
	}

	// Wrap the error with prefix if available
	if prefix != "" {
		err = fmt.Errorf("%s: %w", prefix, err)
	}

	return &errorWithCtx{
		next: err,
		ctx:  c,
	}
}

// ErrorCtx extracts logging context from a wrapped error.
func ErrorCtx(ctx context.Context, err error) context.Context {
	var errWithCtx *errorWithCtx
	if errors.As(err, &errWithCtx) {
		return context.WithValue(ctx, key, errWithCtx.ctx)
	}
	return ctx
}
