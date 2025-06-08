package logger

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type HandlerMiddleware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) *HandlerMiddleware {
	return &HandlerMiddleware{
		next: next,
	}
}

func (h *HandlerMiddleware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

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

func (h *HandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddleware{
		next: h.next.WithAttrs(attrs),
	}
}

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

func WithLogEventID(ctx context.Context, eventID uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.EventID = eventID
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		EventID: eventID,
	})
}

func WithLogMethod(ctx context.Context, method string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Method = method
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{
		Method: method,
	})
}

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

func WrapError(ctx context.Context, err error) error {
	c := logCtx{}
	if x, ok := ctx.Value(key).(logCtx); ok {
		c = x
	}
	return &errorWithCtx{
		next: err,
		ctx:  c,
	}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	var errWithCtx *errorWithCtx
	if errors.As(err, &errWithCtx) {
		return context.WithValue(ctx, key, errWithCtx.ctx)
	}
	return ctx
}
