// Package slogx provides extensions to the [slog] package.
// It focuses on performance and simplicity.
// Only functions working with [slog.Attr] are provided.
// Any slower alternatives are not supported.
package slogx

import (
	"cmp"
	"context"
	"log/slog"
	"runtime"
	"time"
)

// New returns a new [Logger] with the given handler.
func New(handler slog.Handler) *Logger {
	return &Logger{commonLogger{
		handler: handler,
		src:     true,
	}}
}

// NewContextLogger returns a new [ContextLogger] with the given handler.
func NewContextLogger(handler slog.Handler) *ContextLogger {
	return &ContextLogger{commonLogger{
		handler: handler,
		src:     true,
	}}
}

// Default returns a new [Logger] with the default handler from [slog.Default].
func Default() *Logger {
	return New(defaultHandler())
}

// With returns a new [Logger] based on [Default] with the given attributes.
func With(attrs ...slog.Attr) *Logger {
	return Default().With(attrs...)
}

// WithGroup returns a new [Logger] based on [Default] with the given group.
func WithGroup(group string) *Logger {
	return Default().WithGroup(group)
}

// Debug logs a message at the debug level.
func Debug(msg string, attrs ...slog.Attr) {
	logAttrs(context.Background(), defaultHandler(), slog.LevelDebug, msg, attrs)
}

// Info logs a message at the info level.
func Info(msg string, attrs ...slog.Attr) {
	logAttrs(context.Background(), defaultHandler(), slog.LevelInfo, msg, attrs)
}

// Warn logs a message at the warn level.
func Warn(msg string, attrs ...slog.Attr) {
	logAttrs(context.Background(), defaultHandler(), slog.LevelWarn, msg, attrs)
}

// Error logs a message at the error level.
func Error(msg string, attrs ...slog.Attr) {
	logAttrs(context.Background(), defaultHandler(), slog.LevelError, msg, attrs)
}

// Log logs a message at the given level.
func Log(level slog.Level, msg string, attrs ...slog.Attr) {
	logAttrs(context.Background(), defaultHandler(), level, msg, attrs)
}

// ---

// Logger is a simple logger that logs to a [slog.Handler].
// It is an alternative to [slog.Logger] focused on performance and simplicity.
// It forces to use [slog.Attr] for log attributes and does not support slow alternatives provided by [slog.Logger].
// It also takes [slog.Attr] in [Logger.With] because it is the only high performance way to add attributes.
type Logger struct {
	commonLogger
}

// Handler returns the logger's handler.
func (l *Logger) Handler() slog.Handler {
	return l.handlerForExport()
}

// SlogLogger returns a new [slog.Logger] that logs to the associated handler.
func (l *Logger) SlogLogger() *slog.Logger {
	return slog.New(l.handler)
}

// ContextLogger returns a new [ContextLogger] that takes context in logging methods.
func (l *Logger) ContextLogger() *ContextLogger {
	return &ContextLogger{l.commonLogger}
}

// Enabled returns true if the given level is enabled.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.handler.Enabled(ctx, level)
}

// With returns a new [Logger] with the given attributes optimized for short usage (one or two times).
func (l *Logger) With(attrs ...slog.Attr) *Logger {
	if len(attrs) != 0 {
		l = l.clone()
		l.setWithAttrs(attrs)
	}

	return l
}

// WithLongTerm returns a new [Logger] with the given attributes optimized for long usage.
func (l *Logger) WithLongTerm(attrs ...slog.Attr) *Logger {
	if len(attrs) != 0 || l.attrs.Len() != 0 {
		l = l.clone()
		l.setWithAttrs(attrs)
		l.setLongTerm()
	}

	return l
}

// WithGroup returns a new [Logger] with the given group.
func (l *Logger) WithGroup(group string) *Logger {
	if group != "" {
		l = l.clone()
		l.setWithGroup(group)
	}

	return l
}

// WithSource returns a new [Logger] that includes the source file and line in the log record if [enabled] is true.
func (l *Logger) WithSource(enabled bool) *Logger {
	if l.src != enabled {
		l = l.clone()
		l.src = enabled
	}

	return l
}

// Debug logs a message at the debug level.
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.log(context.Background(), slog.LevelDebug, msg, attrs, 0)
}

// DebugContext logs a message at the debug level with the given context.
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelDebug, msg, attrs, 0)
}

// Info logs a message at the info level.
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.log(context.Background(), slog.LevelInfo, msg, attrs, 0)
}

// InfoContext logs a message at the info level with the given context.
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelInfo, msg, attrs, 0)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.log(context.Background(), slog.LevelWarn, msg, attrs, 0)
}

// WarnContext logs a message at the warn level with the given context.
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelWarn, msg, attrs, 0)
}

// Error logs a message at the error level.
func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.log(context.Background(), slog.LevelError, msg, attrs, 0)
}

// ErrorContext logs a message at the error level with the given context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelError, msg, attrs, 0)
}

// Log logs a message at the given level.
func (l *Logger) Log(level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(context.Background(), level, msg, attrs, 0)
}

// LogContext logs a message at the given level.
func (l *Logger) LogContext(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, level, msg, attrs, 0)
}

// LogAttrs logs a message at the given level.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, level, msg, attrs, 0)
}

// LongTerm returns a new [Logger] with the attributes applied to the handler.
func (l *Logger) LongTerm() *Logger {
	if l.attrs.Len() != 0 {
		l = l.clone()
		l.setLongTerm()
	}

	return l
}

func (l Logger) clone() *Logger {
	return &l
}

// ---

// ContextLogger is an alternative to [Logger] having only methods with context for logging messages.
type ContextLogger struct {
	commonLogger
}

// Logger returns a new [Logger] with the associated handler.
func (l *ContextLogger) Logger() *Logger {
	return &Logger{l.commonLogger}
}

// Handler returns the associated handler.
func (l *ContextLogger) Handler() slog.Handler {
	return l.handlerForExport()
}

// SlogLogger returns a new [slog.Logger] that logs to the associated handler.
func (l *ContextLogger) SlogLogger() *slog.Logger {
	return slog.New(l.handler)
}

// Enabled returns true if the given level is enabled.
func (l *ContextLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.handler.Enabled(ctx, level)
}

// With returns a new [ContextLogger] with the given attributes.
func (l *ContextLogger) With(attrs ...slog.Attr) *ContextLogger {
	if len(attrs) != 0 {
		l = l.clone()
		l.setWithAttrs(attrs)
	}

	return l
}

// WithLongTerm returns a new [ContextLogger] with the given attributes optimized for multiple usage.
func (l *ContextLogger) WithLongTerm(attrs ...slog.Attr) *ContextLogger {
	if len(attrs) != 0 {
		l = l.clone()
		l.setWithAttrs(attrs)
		l.setLongTerm()
	}

	return l
}

// WithGroup returns a new [ContextLogger] with the given group.
func (l *ContextLogger) WithGroup(group string) *ContextLogger {
	if group != "" {
		l = l.clone()
		l.setWithGroup(group)
	}

	return l
}

// WithSource returns a new [ContextLogger] that includes the source file and line in the log record if [enabled] is true.
func (l *ContextLogger) WithSource(enabled bool) *ContextLogger {
	if l.src != enabled {
		l = l.clone()
		l.src = enabled
	}

	return l
}

// Debug logs a message at the debug level.
func (l *ContextLogger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelDebug, msg, attrs, 0)
}

// Info logs a message at the info level.
func (l *ContextLogger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelInfo, msg, attrs, 0)
}

// Warn logs a message at the warn level.
func (l *ContextLogger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelWarn, msg, attrs, 0)
}

// Error logs a message at the error level.
func (l *ContextLogger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelError, msg, attrs, 0)
}

// Log logs a message at the given level.
func (l *ContextLogger) Log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, level, msg, attrs, 0)
}

// LogAttrs logs a message at the given level.
func (l *ContextLogger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, level, msg, attrs, 0)
}

// LogWithCallerSkip logs a message at the given level with additional skipping of the specified amount of call stack frames.
func (l *ContextLogger) LogWithCallerSkip(ctx context.Context, skip int, level slog.Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, level, msg, attrs, skip)
}

// LongTerm returns a new [Logger] with the attributes applied to the handler.
func (l *ContextLogger) LongTerm() *ContextLogger {
	if l.attrs.Len() != 0 {
		l = l.clone()
		l.setLongTerm()
	}

	return l
}

func (l ContextLogger) clone() *ContextLogger {
	return &l
}

// ---

func defaultHandler() slog.Handler {
	return slog.Default().Handler()
}

// ---

type commonLogger struct {
	handler slog.Handler
	src     bool
	attrs   AttrPack
}

func (l *commonLogger) handlerForExport() slog.Handler {
	handler := l.handler
	if l.attrs.Len() != 0 {
		handler = handler.WithAttrs(l.attrs.Collect())
	}

	return handler
}

func (l *commonLogger) setWithAttrs(attrs []slog.Attr) {
	if len(attrs) != 0 {
		l.attrs = l.attrs.Clone()
		l.attrs.Add(attrs...)
	}
}

func (l *commonLogger) setWithGroup(group string) {
	if group != "" {
		l.setLongTerm()
		l.handler = l.handler.WithGroup(group)
	}
}

func (l *commonLogger) setLongTerm() {
	if l.attrs.Len() != 0 {
		l.handler = l.handlerForExport()
		l.attrs = AttrPack{}
	}
}

func (l *commonLogger) log(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr, skip int) {
	ctx = cmp.Or(ctx, context.Background())

	if !l.handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	if l.src {
		runtime.Callers(skip+3, pcs[:])
	}

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	if l.attrs.Len() != 0 {
		l.attrs.Enumerate(func(attr slog.Attr) bool {
			r.AddAttrs(attr)

			return true
		})
	}

	r.AddAttrs(attrs...)

	_ = l.handler.Handle(ctx, r)
}

// ---

func logAttrs(ctx context.Context, handler slog.Handler, level slog.Level, msg string, attrs []slog.Attr) {
	l := commonLogger{
		handler: handler,
		src:     true,
	}
	l.log(ctx, level, msg, attrs, 1)
}

// ---

type commonLoggerInterface[T any] interface {
	Handler() slog.Handler
	SlogLogger() *slog.Logger
	Enabled(context.Context, slog.Level) bool
	LogAttrs(context.Context, slog.Level, string, ...slog.Attr)
	With(...slog.Attr) T
	WithLongTerm(...slog.Attr) T
	WithGroup(string) T
	WithSource(bool) T
	LongTerm() T
}

var (
	_ commonLoggerInterface[*Logger]        = (*Logger)(nil)
	_ commonLoggerInterface[*ContextLogger] = (*ContextLogger)(nil)
)
