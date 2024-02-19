package slogjson

import (
	"context"
	"log/slog"
	"time"
)

// Option is a configuration option for the Handler.
type Option func(*options)

// WithLevel sets the leveler for the Handler.
func WithLevel(level slog.Leveler) Option {
	return func(o *options) {
		o.leveler = level
	}
}

// WithAttrReplaceFunc sets the replace attribute function for the Handler.
func WithAttrReplaceFunc(f AttrReplaceFunc) Option {
	return func(o *options) {
		o.replaceAttr = f
	}
}

// WithSource sets whether to include the source information in the log message.
func WithSource(include bool) Option {
	return func(o *options) {
		o.includeSource = include
	}
}

// WithTimeFormat sets the time format for the Handler.
func WithTimeFormat(format string) Option {
	return WithTimeEncodeFunc(timeFormat(format))
}

// WithTimeEncodeFunc sets the time encode function for the Handler.
func WithTimeEncodeFunc(f TimeEncodeFunc) Option {
	return func(o *options) {
		o.encodeTimestamp = f
	}
}

// WithTimeValueFormat sets the time format for the Handler.
func WithTimeValueFormat(format string) Option {
	return WithTimeValueEncodeFunc(timeFormat(format))
}

// WithTimeValueEncodeFunc sets the time encode function for the Handler.
func WithTimeValueEncodeFunc(f TimeEncodeFunc) Option {
	return func(o *options) {
		o.encodeTimeValue = f
	}
}

// WithBytesFormat sets the bytes format for the Handler.
func WithBytesFormat(f BytesFormat) Option {
	return func(o *options) {
		o.bytesFormat = f
	}
}

// WithLevelOffset sets whether to include the level offset in the log message.
func WithLevelOffset(enabled bool) Option {
	return func(o *options) {
		o.levelOffset = enabled
	}
}

// WithLevelReplaceFunc sets the replace level function for the Handler.
func WithLevelReplaceFunc(f LevelReplaceFunc) Option {
	return func(o *options) {
		o.replaceLevel = f
	}
}

// ---

// BytesFormat is a format for the bytes output.
type BytesFormat int

const (
	BytesFormatString = iota // BytesFormatString outputs the bytes as a string.
	BytesFormatHex           // BytesFormatHex outputs the bytes as a hexadecimal string.
	BytesFormatBase64        // BytesFormatBase64 outputs the bytes as a base64 string.
)

// ---

// AttrReplaceFunc is a function that replaces the attributes in the log message.
type AttrReplaceFunc func([]string, slog.Attr) slog.Attr

// TimeEncodeFunc is a function that encodes the time in the log message.
// TimeEncodeFunc can either encode time as a string into the given buffer,
// or return a slog.Value representing the time in one of the following kinds:
// KindFloat64, KindInt64, KindString, KindUint64.
// If the returned buffer is nil, the time is encoded as a slog.Value.
// If the returned buffer is nil or empty and slog.Value is not of the kinds
// listed above, the time attribute is dropped.
type TimeEncodeFunc func([]byte, time.Time) ([]byte, slog.Value)

// DurationEncodeFunc is a function that encodes the duration in the log message.
type DurationEncodeFunc func(time.Duration) slog.Value

// SourceEncodeFunc is a function that encodes the source in the log message.
type SourceEncodeFunc func(slog.Source) slog.Value

// LevelReplaceFunc is a function that replaces the level in the log message.
type LevelReplaceFunc func(context.Context, slog.Level) slog.Level

// ---

type options struct {
	leveler         slog.Leveler
	replaceAttr     AttrReplaceFunc
	encodeTimestamp TimeEncodeFunc
	encodeTimeValue TimeEncodeFunc
	encodeDuration  DurationEncodeFunc
	encodeSource    SourceEncodeFunc
	replaceLevel    LevelReplaceFunc
	includeSource   bool
	levelOffset     bool
	bytesFormat     BytesFormat
}

func defaultOptions() options {
	return options{
		leveler:         slog.LevelInfo,
		encodeTimestamp: timeFormat(time.RFC3339Nano),
		encodeTimeValue: timeFormat(time.RFC3339Nano),
		encodeDuration:  DurationAsSeconds(),
		encodeSource:    SourceLongObject(),
		replaceLevel:    doNotReplaceLevel,
	}
}

func (o options) with(opts []Option) options {
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return o
}

// ---

func timeFormat(layout string) TimeEncodeFunc {
	return func(buf []byte, value time.Time) ([]byte, slog.Value) {
		return value.AppendFormat(buf, layout), slog.Value{}
	}
}

func doNotReplaceLevel(_ context.Context, level slog.Level) slog.Level {
	return level
}
