// Package slogc provides context-centric logging facilities based on [slogx] and [slog] packages.
package slogc

import (
	"context"
	"log/slog"

	"github.com/pamburus/slogx"
)

// Logger is an alias for [slogx.ContextLogger].
type Logger = slogx.ContextLogger

// New returns a new context with the provided logger.
func New(ctx context.Context, logger *Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, &contextKeyLogger, logger)
}

// Get returns a [Logger] from the given context stored by [New].
// If the context is nil or the logger is not found, a [Default] logger is returned.
func Get(ctx context.Context) *Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(&contextKeyLogger).(*Logger); ok {
			return logger
		}
	}

	return Default()
}

// Default returns a [Logger] with the default handler from [slog.Default].
func Default() *Logger {
	return slogx.Default().ContextLogger()
}

// With returns a new context with a modified logger containing additionally the provided attributes.
func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	return New(ctx, Get(ctx).With(attrs...))
}

// WithGroup returns a new context with a modified logger containing the provided group.
func WithGroup(ctx context.Context, key string) context.Context {
	return New(ctx, Get(ctx).WithGroup(key))
}

// WithName returns a new context with a modified logger containing the provided name.
func WithName(ctx context.Context, name string) context.Context {
	return New(ctx, Get(ctx).WithName(name))
}

// WithSource returns a new context with a modified logger containing the source information enabled flag.
func WithSource(ctx context.Context, enabled bool) context.Context {
	return New(ctx, Get(ctx).WithSource(enabled))
}

// ---

// Debug logs a message at debug level.
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	Get(ctx).LogWithCallerSkip(ctx, 1, slog.LevelDebug, msg, attrs...)
}

// Info logs a message at info level.
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	Get(ctx).LogWithCallerSkip(ctx, 1, slog.LevelInfo, msg, attrs...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	Get(ctx).LogWithCallerSkip(ctx, 1, slog.LevelWarn, msg, attrs...)
}

// Error logs a message at error level.
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	Get(ctx).LogWithCallerSkip(ctx, 1, slog.LevelError, msg, attrs...)
}

// Log logs a message with attributes at the given level.
func Log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Get(ctx).LogWithCallerSkip(ctx, 1, level, msg, attrs...)
}

// ---

var (
	contextKeyLogger int
)
