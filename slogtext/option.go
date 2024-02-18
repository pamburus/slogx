package slogtext

import (
	"log/slog"
	"math"
	"time"

	"github.com/pamburus/slogx/slogtext/themes"
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

// WithMultilineExpansion sets whether to expand multiline strings in the log message.
func WithMultilineExpansion(setting ExpansionThreshold) Option {
	return func(o *options) {
		o.expansionThreshold = setting
	}
}

// WithTheme sets the theme for the Handler.
func WithTheme(theme Theme) Option {
	return func(o *options) {
		o.theme = theme
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

// ---

// ColorSetting is a setting for the color output.
type ColorSetting int

const (
	ColorAuto   ColorSetting = iota // ColorAuto enables color output if the output is a terminal.
	ColorNever                      // ColorNever disables color output.
	ColorAlways                     // ColorAlways enables color output.
)

// ---

// BytesFormat is a format for the bytes output.
type BytesFormat int

const (
	BytesFormatString = iota // BytesFormatString outputs the bytes as a string.
	BytesFormatHex           // BytesFormatHex outputs the bytes as a hex string.
	BytesFormatBase64        // BytesFormatBase64 outputs the bytes as a base64 string.
)

// ---

// ExpansionThreshold is a setting for the multiline string expansion.
type ExpansionThreshold int

const (
	ExpandAuto  ExpansionThreshold = 0           // ExpandAuto enables multiline string expansion if recommended.
	ExpandNever                    = math.MaxInt // ExpandNever disables multiline string expansion.
	ExpandAll                      = -1          // ExpandAll enables all multiline string expansion.
)

// ExpandIfOver returns an expansion threshold that expands multiline strings if length is over the given threshold.
func ExpandIfOver(threshold int) ExpansionThreshold {
	return ExpansionThreshold(threshold)
}

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
	leveler            slog.Leveler
	color              ColorSetting
	replaceAttr        ReplaceAttrFunc
	encodeTimestamp    TimeEncodeFunc
	encodeTimeValue    TimeEncodeFunc
	encodeDuration     DurationEncodeFunc
	encodeSource       SourceEncodeFunc
	includeSource      bool
	levelOffset        bool
	expansionThreshold ExpansionThreshold
	bytesFormat        BytesFormat
	theme              Theme
}

func defaultOptions() options {
	return options{
		leveler:            slog.LevelInfo,
		color:              ColorAuto,
		encodeTimestamp:    timeFormat(time.StampMilli),
		encodeTimeValue:    timeFormat(time.StampMilli),
		encodeDuration:     DurationAsSeconds(),
		encodeSource:       SourceShort(),
		expansionThreshold: 32,
		theme:              themes.Default(),
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
	return func(buf []byte, t time.Time) []byte {
		return t.AppendFormat(buf, layout)
	}
}
