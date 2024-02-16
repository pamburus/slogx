package slogtext

import (
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

// WithColor sets the color setting for the Handler.
func WithColor(setting ColorSetting) Option {
	return func(o *options) {
		o.color = setting
	}
}

// WithReplaceAttrFunc sets the replace attribute function for the Handler.
func WithReplaceAttrFunc(f ReplaceAttrFunc) Option {
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

// WithTheme sets the theme for the Handler.
func WithTheme(theme Theme) Option {
	return func(o *options) {
		o.theme = theme
	}
}

// WithTimestampLayout sets the time layout for the Handler.
func WithTimestampLayout(layout string) Option {
	return WithTimestampEncodeFunc(timeLayout(layout))
}

// WithTimestampEncodeFunc sets the time encode function for the Handler.
func WithTimestampEncodeFunc(f TimeEncodeFunc) Option {
	return func(o *options) {
		o.encodeTimestamp = f
	}
}

// WithTimeValueLayout sets the time layout for the Handler.
func WithTimeValueLayout(layout string) Option {
	return WithTimeValueEncodeFunc(timeLayout(layout))
}

// WithTimeValueEncodeFunc sets the time encode function for the Handler.
func WithTimeValueEncodeFunc(f TimeEncodeFunc) Option {
	return func(o *options) {
		o.encodeTimeValue = f
	}
}

// ---

// ColorSetting is a setting for the color output.
type ColorSetting int

const (
	ColorAuto   ColorSetting = iota // ColorAuto enables color output if the output is a terminal.
	ColorNever                      // ColorNever disables color output.
	ColorAlways                     // ColorAlways enables color output.
)

// ---

// ReplaceAttrFunc is a function that replaces the attributes in the log message.
type ReplaceAttrFunc func([]string, slog.Attr) slog.Attr

// TimeEncodeFunc is a function that encodes the time in the log message.
type TimeEncodeFunc func([]byte, time.Time) []byte

// DurationEncodeFunc is a function that encodes the duration in the log message.
type DurationEncodeFunc func([]byte, time.Duration) []byte

// SourceEncodeFunc is a function that encodes the source in the log message.
type SourceEncodeFunc func([]byte, slog.Source) []byte

// ---

type options struct {
	leveler         slog.Leveler
	color           ColorSetting
	replaceAttr     ReplaceAttrFunc
	encodeTimestamp TimeEncodeFunc
	encodeTimeValue TimeEncodeFunc
	encodeDuration  DurationEncodeFunc
	encodeSource    SourceEncodeFunc
	includeSource   bool
	theme           Theme
}

func defaultOptions() options {
	return options{
		leveler:         slog.LevelInfo,
		color:           ColorAuto,
		encodeTimestamp: timeLayout(time.StampMilli),
		encodeTimeValue: timeLayout(time.StampMilli),
		encodeDuration:  DurationAsSeconds(),
		encodeSource:    SourceShort(),
		theme:           ThemeDefault(),
	}
}

func (o options) with(opts []Option) options {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// ---

func timeLayout(layout string) TimeEncodeFunc {
	return func(buf []byte, t time.Time) []byte {
		return t.AppendFormat(buf, layout)
	}
}
