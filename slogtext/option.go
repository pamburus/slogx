package slogtext

import (
	"context"
	"io"
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
		o.enableColor = setting
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

// WithSourceKey sets the source key for the Handler.
func WithSourceKey(key string) Option {
	return func(o *options) {
		o.sourceKey = key
	}
}

// WithExpansion sets attribute expansion setting for the Handler.
func WithExpansion(setting ExpansionSetting) Option {
	return func(o *options) {
		if setting < 0 || setting >= numExpansionSettings {
			setting = ExpansionAuto
		}
		o.expansion = expansionProfiles[setting]
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

// WithLevelReplaceFunc sets the replace level function for the Handler.
func WithLevelReplaceFunc(f LevelReplaceFunc) Option {
	return func(o *options) {
		o.replaceLevel = f
	}
}

// WithLoggerFromContext sets the logger name from context function for the Handler.
func WithLoggerFromContext(fn func(context.Context) string) Option {
	return func(o *options) {
		o.loggerFromContext = fn
	}
}

// WithLoggerKey sets the logger name key for the Handler.
func WithLoggerKey(key string) Option {
	return func(o *options) {
		o.loggerKey = key
	}
}

// ---

// ColorSetting is a setting for the color output.
type ColorSetting func(io.Writer) bool

// ColorAlways enables color output.
func ColorAlways(io.Writer) bool {
	return true
}

// ColorNever disables color output.
func ColorNever(io.Writer) bool {
	return false
}

// ---

// BytesFormat is a format for the bytes output.
type BytesFormat int

const (
	BytesFormatString = iota // BytesFormatString outputs the bytes as a string.
	BytesFormatHex           // BytesFormatHex outputs the bytes as a hex string.
	BytesFormatBase64        // BytesFormatBase64 outputs the bytes as a base64 string.
)

// ---

type ExpansionSetting int

const (
	ExpansionAuto   ExpansionSetting = iota // ExpansionAuto enables attribute expansion when needed.
	ExpansionNever                          // ExpansionNever disables attribute expansion completely.
	ExpansionLow                            // ExpansionLow enables attribute expansion for multiline strings only that are 32+ characters long.
	ExpansionMedium                         // ExpansionMedium enables attribute expansion for multiline strings, long strings and moderately escaped strings as well as very long messages.
	ExpansionHigh                           // ExpansionHigh enables attribute expansion for most attributes if they are long or multiline.
	ExpansionAlways                         // ExpansionAlways enables attribute expansion for all attributes.
	numExpansionSettings
)

// ---

// AttrReplaceFunc is a function that replaces the attributes in the log message.
type AttrReplaceFunc func([]string, slog.Attr) slog.Attr

// TimeEncodeFunc is a function that encodes the time in the log message.
type TimeEncodeFunc func([]byte, time.Time) []byte

// DurationEncodeFunc is a function that encodes the duration in the log message.
type DurationEncodeFunc func([]byte, time.Duration) []byte

// SourceEncodeFunc is a function that encodes the source in the log message.
type SourceEncodeFunc func([]byte, slog.Source) []byte

// LevelReplaceFunc is a function that replaces the level in the log message.
type LevelReplaceFunc func(context.Context, slog.Level) slog.Level

// ---

type options struct {
	leveler           slog.Leveler
	enableColor       ColorSetting
	replaceAttr       AttrReplaceFunc
	encodeTimestamp   TimeEncodeFunc
	encodeTimeValue   TimeEncodeFunc
	encodeDuration    DurationEncodeFunc
	encodeSource      SourceEncodeFunc
	replaceLevel      LevelReplaceFunc
	includeSource     bool
	sourceKey         string
	levelOffset       bool
	expansion         expansionProfile
	bytesFormat       BytesFormat
	loggerFromContext func(context.Context) string
	loggerKey         string
	theme             Theme
}

func defaultOptions() options {
	return options{
		leveler:         slog.LevelInfo,
		enableColor:     ColorNever,
		encodeTimestamp: timeFormat(time.StampMilli),
		encodeTimeValue: timeFormat(time.StampMilli),
		encodeDuration:  DurationAsSeconds(),
		encodeSource:    SourceShort(),
		sourceKey:       slog.SourceKey,
		replaceLevel:    doNotReplaceLevel,
		expansion:       expansionProfiles[ExpansionAuto],
		theme:           themes.Default(),
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

func doNotReplaceLevel(_ context.Context, level slog.Level) slog.Level {
	return level
}

// ---

var expansionProfiles = [numExpansionSettings]expansionProfile{
	ExpansionAuto:   {32, 8, 128, 256, 16, 24},
	ExpansionNever:  {math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt},
	ExpansionLow:    {128, 16, 256, 512, 32, 48},
	ExpansionMedium: {32, 8, 128, 256, 16, 24},
	ExpansionHigh:   {32, 4, 64, 192, 8, 16},
	ExpansionAlways: {0, 0, 0, 0, 0, 0},
}

type expansionProfile struct {
	multilineLengthThreshold int
	escapeThreshold          int
	attrLengthThreshold      int
	totalLengthThreshold     int
	attrCountThreshold       int
	keyLengthThreshold       int
}
