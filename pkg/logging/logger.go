package logging

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type TickLoggerHandler struct {
	h     slog.Handler
	b     *bytes.Buffer
	m     *sync.Mutex
	clock GlobalTimer
}

func NewTickLoggerHandler(h slog.Handler, clock GlobalTimer) *TickLoggerHandler {
	return &TickLoggerHandler{h: h, b: &bytes.Buffer{}, m: &sync.Mutex{}, clock: clock}
}

func (h *TickLoggerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *TickLoggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TickLoggerHandler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *TickLoggerHandler) WithGroup(name string) slog.Handler {
	return &TickLoggerHandler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *TickLoggerHandler) Handle(ctx context.Context, r slog.Record) error {
	currentTick := fmt.Sprintf("tick %d: ", h.clock.GetCurrentTick())

	fmt.Println(
		currentTick,
		r.Message,
	)

	return nil
}

type GlobalTimer interface {
	GetCurrentTick() int
}

func ParseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic(fmt.Sprintf("Unknown log level %s", level))
	}
}
