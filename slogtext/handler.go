// Package slogtext provides slog.Handler implementation that output log messages in a textual human-readable form with colors.
package slogtext

import (
	"bytes"
	"context"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/pamburus/slogx/internal/quoting"
	"github.com/pamburus/slogx/internal/tty"
)

// NewHandler returns a new slog.Handler with the given Writer and
// optional custom configuration.
func NewHandler(writer io.Writer, options ...Option) *Handler {
	opts := defaultOptions().with(options)

	if opts.color == ColorAuto {
		if f, ok := writer.(*os.File); ok {
			if tty.EnableSeqTTY(f, true) {
				opts.color = ColorAlways
			} else {
				opts.color = ColorNever
			}
		}
	}

	if opts.color == ColorNever {
		opts.theme = opts.theme.Plain()
	}

	return &Handler{
		shared: &shared{
			opts,
			newThemeCache(&opts.theme),
			sync.Mutex{},
			writer,
		},
	}
}

// ---

// Handler is a slog.Handler implementation that writes log messages in a
// textual human-readable form.
type Handler struct {
	*shared
	attrs     []slog.Attr
	groups    groups
	groupKeys groupKeys
	keyPrefix string
	cache     cache
}

// Enabled returns true if the given level is enabled.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.leveler.Level()
}

// Handle writes the log message to the Writer.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	hs := newHandleState(ctx, h)
	defer hs.release()

	replace := h.replaceAttr

	if !record.Time.IsZero() {
		val := record.Time.Round(0)
		h.tc.Timestamp.open(hs)
		if replace == nil {
			h.appendTimestamp(hs, record.Time)
		} else if attr := replace(nil, slog.Time(slog.TimeKey, val)); attr.Key != "" {
			if attr.Value.Kind() == slog.KindTime {
				h.appendTimestamp(hs, attr.Value.Time())
			} else {
				h.appendValue(hs, attr.Value, false)
			}
		}
		h.tc.Timestamp.close(hs)
	}

	h.appendLevel(hs, record.Level)

	h.tc.Message.open(hs)
	if replace == nil {
		hs.buf.AppendString(record.Message)
	} else if a := replace(nil, slog.String(slog.MessageKey, record.Message)); a.Key != "" {
		h.appendValue(hs, a.Value, false)
	}
	h.tc.Message.close(hs)

	h.appendHandlerAttrs(hs)

	record.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(hs, attr, len(h.keyPrefix))

		return true
	})

	if h.includeSource {
		src := h.source(record.PC)
		if src.File != "" {
			h.tc.Source.open(hs)
			if replace == nil {
				h.appendSource(hs, src)
			} else if attr := replace(nil, slog.Any(slog.SourceKey, src)); attr.Key != "" {
				h.appendValue(hs, attr.Value, false)
			}
			h.tc.Source.close(hs)
		}
	}

	if hs.buf.Len() == 0 {
		return nil
	}

	hs.buf.SetBack('\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.writer.Write(hs.buf)

	return err
}

// WithAttrs returns a new Handler with the given attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h = h.fork()
	if h.attrs == nil && !slices.ContainsFunc(attrs, slog.Attr{}.Equal) {
		h.attrs = attrs
	} else {
		h.attrs = slices.Grow(h.attrs, len(attrs))
		for _, a := range attrs {
			if !a.Equal(slog.Attr{}) {
				h.attrs = append(h.attrs, a)
			}
		}
	}

	return h
}

// WithGroup returns a new Handler with the given group.
func (h *Handler) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	h = h.fork()

	prefixLen := len(h.keyPrefix)
	h.keyPrefix += key + "."
	h.groups.append(group{len(h.attrs), prefixLen})
	h.groupKeys.append(key)

	return h
}

func (h *Handler) fork() *Handler {
	return &Handler{
		h.shared,
		slices.Clip(h.attrs),
		h.groups.fork(),
		h.groupKeys.fork(),
		h.keyPrefix,
		cache{
			attrs:     h.cache.attrs,
			numGroups: h.cache.numGroups,
			numAttrs:  h.cache.numAttrs,
		},
	}
}

func (h *Handler) appendTimestamp(hs *handleState, value time.Time) {
	hs.buf = h.encodeTimestamp(hs.buf, value)
}

func (h *Handler) appendTime(hs *handleState, value time.Time, quote bool) {
	hs.scratch.Reset()
	hs.scratch = h.encodeTimeValue(hs.scratch, value)
	if quote {
		h.appendAutoQuotedByteString(hs, hs.scratch.Bytes())
	} else {
		hs.buf.AppendBytes(hs.scratch.Bytes())
	}
}

func (h *Handler) appendDuration(hs *handleState, value time.Duration, quote bool) {
	hs.scratch.Reset()
	hs.scratch = h.encodeDuration(hs.scratch, value)
	if quote {
		h.appendAutoQuotedByteString(hs, hs.scratch.Bytes())
	} else {
		hs.buf.AppendBytes(hs.scratch.Bytes())
	}
}

func (h *Handler) appendHandlerAttrs(hs *handleState) {
	if h == nil || len(h.attrs) == 0 && h.groups.len() == 0 {
		return
	}

	appended := false

	h.cache.once.Do(func() {
		if h.cache.numAttrs == len(h.attrs) && h.cache.numGroups == h.groups.len() {
			return
		}

		pos := hs.buf.Len()
		hs.buf.AppendString(h.cache.attrs)

		begin := h.cache.numAttrs
		for i := h.cache.numGroups; i != h.groups.len(); i++ {
			group := h.groups.at(i)
			end := group.i
			for _, attr := range h.attrs[begin:end] {
				h.appendAttr(hs, attr, group.prefixLen)
			}
			begin = end
		}

		for _, attr := range h.attrs[begin:] {
			h.appendAttr(hs, attr, len(h.keyPrefix))
		}

		h.cache.attrs = hs.buf[pos:].String()
		h.cache.numGroups = h.groups.len()
		h.cache.numAttrs = len(h.attrs)
		appended = true
	})

	if !appended && len(h.cache.attrs) != 0 {
		hs.buf.AppendString(h.cache.attrs)
	}
}

func (h *Handler) appendAttr(hs *handleState, attr slog.Attr, basePrefixLen int) {
	attr.Value = attr.Value.Resolve()
	if rep := h.replaceAttr; rep != nil && attr.Value.Kind() != slog.KindGroup {
		attr = rep(hs.groups, attr)
		attr.Value = attr.Value.Resolve()
	}

	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			hs.keyPrefix.AppendString(attr.Key)
			hs.keyPrefix.AppendByte('.')
			hs.groups = append(hs.groups, attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(hs, groupAttr, basePrefixLen)
		}
		if attr.Key != "" {
			hs.keyPrefix = hs.keyPrefix[:hs.keyPrefix.Len()-len(attr.Key)-1]
			hs.groups = hs.groups[:len(hs.groups)-1]
		}
	} else {
		h.appendKey(hs, attr.Key, basePrefixLen)
		h.appendValue(hs, attr.Value, true)
		hs.buf.AppendByte(' ')
	}
}

func (h *Handler) appendKey(hs *handleState, key string, basePrefixLen int) {
	hs.buf.AppendString(h.tc.Key.prefix)
	hs.buf.AppendString(h.keyPrefix[:basePrefixLen])
	hs.buf.AppendBytes(hs.keyPrefix)
	hs.buf.AppendString(key)
	hs.buf.AppendString(h.tc.Key.suffix)
}

func (h *Handler) appendValue(hs *handleState, v slog.Value, quote bool) {
	switch v.Kind() {
	case slog.KindString:
		if h.tc.String.empty {
			h.appendString(hs, v.String(), quote)
		} else {
			hs.buf.AppendString(h.tc.String.prefix)
			h.appendString(hs, v.String(), quote)
			hs.buf.AppendString(h.tc.String.suffix)
		}
	case slog.KindInt64:
		h.tc.Number.apply(hs, func() {
			hs.buf.AppendInt(v.Int64())
		})
	case slog.KindUint64:
		h.tc.Number.apply(hs, func() {
			hs.buf.AppendUint(v.Uint64())
		})
	case slog.KindFloat64:
		h.tc.Number.apply(hs, func() {
			hs.buf.AppendFloat64(v.Float64())
		})
	case slog.KindBool:
		h.tc.Bool.apply(hs, func() {
			hs.buf.AppendBool(v.Bool())
		})
	case slog.KindDuration:
		h.tc.Duration.apply(hs, func() {
			h.appendDuration(hs, v.Duration(), quote)
		})
	case slog.KindGroup:
		attrs := v.Group()
		if len(attrs) == 0 {
			hs.buf.AppendString(h.tc.EmptyMap)
		} else {
			hs.buf.AppendString(h.tc.Map.prefix)
			for i, attr := range attrs {
				if i != 0 {
					hs.buf.AppendString(h.tc.MapSep1)
				}
				h.appendString(hs, attr.Key, quote)
				hs.buf.AppendString(h.tc.MapSep2)
				h.appendValue(hs, attr.Value, true)
			}
			hs.buf.AppendString(h.tc.Map.suffix)
		}
	case slog.KindTime:
		h.tc.Time.apply(hs, func() {
			h.appendTime(hs, v.Time(), quote)
		})
	case slog.KindAny:
		switch v := v.Any().(type) {
		case nil:
			hs.buf.AppendString(h.tc.Null)
		case slog.Level:
			h.appendLevelValue(hs, v)
		case error:
			h.tc.Error.apply(hs, func() {
				h.appendString(hs, v.Error(), quote)
			})
		case fmt.Stringer:
			if v, ok := safeResolveValue(h, hs, v.String); ok {
				h.appendString(hs, v, quote)
			}
		case encoding.TextMarshaler:
			if data, ok := safeResolveValueErr(h, hs, v.MarshalText); ok {
				h.appendByteString(hs, data, quote)
			}
		case *slog.Source:
			h.appendSource(hs, *v)
		case slog.Source:
			h.appendSource(hs, v)
		case []byte:
			h.appendBytesValue(hs, v, quote)
		default:
			h.appendAnyValue(hs, v, quote)
		}
	}
}

func (h *Handler) appendBytesValue(hs *handleState, v []byte, quote bool) {
	switch h.bytesFormat {
	default:
		fallthrough
	case BytesFormatString:
		hs.buf.AppendString(h.tc.String.prefix)
		h.appendByteString(hs, v, quote)
		hs.buf.AppendString(h.tc.String.suffix)
	case BytesFormatHex:
		hs.buf.AppendString(h.tc.QuotedString.prefix)
		hex.Encode(hs.buf.Extend(hex.EncodedLen(len(v))), v)
		hs.buf.AppendString(h.tc.QuotedString.suffix)
	case BytesFormatBase64:
		hs.buf.AppendString(h.tc.QuotedString.prefix)
		base64.StdEncoding.Encode(hs.buf.Extend(base64.StdEncoding.EncodedLen(len(v))), v)
		hs.buf.AppendString(h.tc.QuotedString.suffix)
	}
}

func (h *Handler) appendAnyValue(hs *handleState, v any, quote bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		switch {
		case rv.IsNil():
			hs.buf.AppendString(h.tc.Null)
		case rv.Len() == 0:
			hs.buf.AppendString(h.tc.EmptyArray)
		default:
			hs.buf.AppendString(h.tc.Array.prefix)
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					hs.buf.AppendString(h.tc.ArraySep)
				}
				h.appendValue(hs, slog.AnyValue(rv.Index(i).Interface()), quote)
			}
			hs.buf.AppendString(h.tc.Array.suffix)
		}
	case reflect.Map:
		switch {
		case rv.IsNil():
			hs.buf.AppendString(h.tc.Null)
		case rv.Len() == 0:
			hs.buf.AppendString(h.tc.EmptyMap)
		default:
			hs.buf.AppendString(h.tc.Map.prefix)
			for i, k := range rv.MapKeys() {
				if i != 0 {
					hs.buf.AppendString(h.tc.MapSep1)
				}
				h.appendValue(hs, slog.AnyValue(k.Interface()), true)
				hs.buf.AppendString(h.tc.MapSep2)
				h.appendValue(hs, slog.AnyValue(rv.MapIndex(k).Interface()), true)
			}
			hs.buf.AppendString(h.tc.Map.suffix)
		}
	default:
		hs.scratch.Reset()
		_, _ = fmt.Fprintf(&hs.scratch, "%+v", v)
		h.appendByteString(hs, hs.scratch.Bytes(), quote)
	}
}

func (h *Handler) appendString(hs *handleState, s string, quote bool) {
	if quote {
		if s == "null" {
			hs.buf.AppendString(h.tc.QuotedNull)
		} else {
			h.appendAutoQuotedString(hs, s)
		}
	} else {
		hs.buf.AppendString(s)
	}
}

func (h *Handler) appendByteString(hs *handleState, s []byte, quote bool) {
	if quote {
		if bytes.Equal(s, []byte("null")) {
			hs.buf.AppendString(h.tc.QuotedNull)
		} else {
			h.appendAutoQuotedByteString(hs, s)
		}
	} else {
		hs.buf.AppendBytes(s)
	}
}

func (h *Handler) appendAutoQuotedString(hs *handleState, v string) {
	switch {
	case len(v) == 0:
		hs.buf.AppendString(h.tc.QuadQuote)
	case quoting.IsNeeded(v):
		h.appendQuotedString(hs, v)
	default:
		hs.buf.AppendString(v)
	}
}

func (h *Handler) appendQuotedString(hs *handleState, v string) {
	hs.buf.AppendString(h.tc.DoubleQuote)
	h.appendEscapedString(hs, v)
	hs.buf.AppendString(h.tc.DoubleQuote)
}

func (h *Handler) appendAutoQuotedByteString(hs *handleState, v []byte) {
	switch {
	case len(v) == 0:
		hs.buf.AppendString(h.tc.QuadQuote)
	case quoting.IsNeededForBytes(v):
		h.appendQuotedByteString(hs, v)
	default:
		hs.buf.AppendBytes(v)
	}
}

func (h *Handler) appendQuotedByteString(hs *handleState, v []byte) {
	hs.buf.AppendString(h.tc.DoubleQuote)
	h.appendEscapedByteString(hs, v)
	hs.buf.AppendString(h.tc.DoubleQuote)
}

func (h *Handler) appendEscapedString(hs *handleState, s string) {
	p := 0

	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c < utf8.RuneSelf && c >= 0x20 && c != '\\' && c != '"':
			i++

		case c < utf8.RuneSelf:
			hs.buf.AppendString(s[p:i])
			switch c {
			case '\t':
				hs.buf.AppendString(h.tc.EscTab)
			case '\r':
				hs.buf.AppendString(h.tc.EscCR)
			case '\n':
				hs.buf.AppendString(h.tc.EscLF)
			case '\\':
				hs.buf.AppendString(h.tc.EscBackslash)
			case '"':
				hs.buf.AppendString(h.tc.EscQuote)
			default:
				hs.buf.AppendString(h.theme.StringEscape.Prefix)
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hexDigits[c>>4])
				hs.buf.AppendByte(hexDigits[c&0xf])
				hs.buf.AppendString(h.theme.StringEscape.Suffix)
			}
			i++
			p = i

		default:
			v, wd := utf8.DecodeRuneInString(s[i:])
			if v == utf8.RuneError && wd == 1 {
				hs.buf.AppendString(s[p:i])
				hs.buf.AppendString(h.theme.StringEscape.Prefix)
				hs.buf.AppendString(`\ufffd`)
				hs.buf.AppendString(h.theme.StringEscape.Suffix)
				i++
				p = i
			} else {
				i += wd
			}
		}
	}

	hs.buf.AppendString(s[p:])
}

func (h *Handler) appendEscapedByteString(hs *handleState, s []byte) {
	p := 0

	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c < utf8.RuneSelf && c >= 0x20 && c != '\\' && c != '"':
			i++

		case c < utf8.RuneSelf:
			hs.buf.AppendBytes(s[p:i])
			switch c {
			case '\t':
				hs.buf.AppendString(h.tc.EscTab)
			case '\r':
				hs.buf.AppendString(h.tc.EscCR)
			case '\n':
				hs.buf.AppendString(h.tc.EscLF)
			case '\\':
				hs.buf.AppendString(h.tc.EscBackslash)
			case '"':
				hs.buf.AppendString(h.tc.EscQuote)
			default:
				hs.buf.AppendString(h.theme.StringEscape.Prefix)
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hexDigits[c>>4])
				hs.buf.AppendByte(hexDigits[c&0xf])
				hs.buf.AppendString(h.theme.StringEscape.Suffix)
			}
			i++
			p = i

		default:
			v, wd := utf8.DecodeRune(s[i:])
			if v == utf8.RuneError && wd == 1 {
				hs.buf.AppendBytes(s[p:i])
				hs.buf.AppendString(h.theme.StringEscape.Prefix)
				hs.buf.AppendString(`\ufffd`)
				hs.buf.AppendString(h.theme.StringEscape.Suffix)
				i++
				p = i
			} else {
				i += wd
			}
		}
	}

	hs.buf.AppendBytes(s[p:])
}

func (h *Handler) appendSource(hs *handleState, source slog.Source) {
	hs.buf = h.encodeSource(hs.buf, source)
}

func (h *Handler) appendLevel(hs *handleState, level slog.Level) {
	if h.levelOffset {
		h.appendLevelWithOffset(hs, level)
	} else {
		hs.buf.AppendString(h.tc.LevelLabel[levelIndex(level)])
	}
}

func (h *Handler) appendLevelWithOffset(hs *handleState, level slog.Level) {
	appendOffset := func(offset int64) {
		switch {
		case offset > 9:
			hs.buf.AppendByte('+')
			hs.buf.AppendByte('~')
		case offset < -9:
			hs.buf.AppendByte('-')
			hs.buf.AppendByte('~')
		case offset >= 0:
			hs.buf.AppendByte('+')
			fallthrough
		default:
			hs.buf.AppendInt(offset)
		}
	}

	i := levelIndex(level)
	hs.buf.AppendString(h.tc.Level[i].prefix)
	offset := int64(level - levels[i])
	if offset != 0 {
		hs.buf.AppendString(levelLabels[i][:1])
		appendOffset(offset)
	} else {
		hs.buf.AppendString(levelLabels[i])
	}
	hs.buf.AppendString(h.tc.Level[i].suffix)
}

func (h *Handler) appendLevelValue(hs *handleState, level slog.Level) {
	appendOffset := func(offset int64) {
		if offset != 0 {
			if offset > 0 {
				hs.buf.AppendByte('+')
			}
			hs.buf.AppendInt(offset)
		}
	}

	i := levelIndex(level)
	hs.buf.AppendString(h.theme.LevelValue[i].Prefix)
	hs.buf.AppendString(levelNames[i])
	appendOffset(int64(level - levels[i]))
	hs.buf.AppendString(h.theme.LevelValue[i].Suffix)
}

func (h *Handler) source(pc uintptr) slog.Source {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	return slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

func (h *Handler) appendEncodeError(hs *handleState, err error) {
	h.tc.EvalError.apply(hs, func() {
		h.appendQuotedString(hs, err.Error())
	})
}

func (h *Handler) appendEncodePanic(hs *handleState, p any) {
	hs.scratch.Reset()
	_, _ = fmt.Fprintf(&hs.scratch, "%v", p)

	h.tc.EvalPanic.apply(hs, func() {
		h.appendQuotedByteString(hs, hs.scratch)
	})
}

// ---

type shared struct {
	options
	tc     themeCache
	mu     sync.Mutex
	writer io.Writer
}

// ---

func newStyleCache(theme *Theme) styleCache {
	sc := styleCache{
		Source:    styleFromTheme(theme.Source).withTrailingSpace(),
		Timestamp: styleFromTheme(theme.Time).withTrailingSpace(),
		Key:       styleFromTheme(theme.Key).withExtraSuffix(theme.KeyValueSep.Prefix + "=" + theme.KeyValueSep.Suffix),
		Message:   styleFromTheme(theme.Message).withTrailingSpace(),
		String:    styleFromTheme(theme.StringValue),
		Quote:     styleFromTheme(theme.StringQuote),
		Escape:    styleFromTheme(theme.StringEscape),
		Number:    styleFromTheme(theme.NumberValue),
		Bool:      styleFromTheme(theme.BoolValue),
		Error:     styleFromTheme(theme.ErrorValue),
		Duration:  styleFromTheme(theme.DurationValue),
		Time:      styleFromTheme(theme.TimeValue),
		EvalError: styleFromTheme(theme.EvalError),
		EvalPanic: styleFromTheme(theme.EvalPanic),
		Array:     newStyle(styleFromTheme(theme.Array.Begin).render("["), styleFromTheme(theme.Array.End).render("]")),
		Map:       newStyle(styleFromTheme(theme.Map.Begin).render("{"), styleFromTheme(theme.Map.End).render("}")),
	}

	quote := sc.Quote.render(`"`)
	sc.QuotedString.set(sc.String.prefix+quote, quote+sc.String.suffix)

	for i := 0; i < NumLevels; i++ {
		sc.Level[i] = styleFromTheme(theme.Level[i]).withTrailingSpace()
	}

	return sc
}

type styleCache struct {
	Source       style
	Timestamp    style
	Key          style
	Message      style
	String       style
	QuotedString style
	Quote        style
	Escape       style
	Number       style
	Bool         style
	Error        style
	Duration     style
	Time         style
	EvalError    style
	EvalPanic    style
	Array        style
	Map          style
	Level        [4]style
}

// ---

func newThemeCache(theme *Theme) themeCache {
	sc := newStyleCache(theme)
	tc := themeCache{
		styleCache: sc,
	}

	tc.DoubleQuote = tc.Quote.render(`"`)
	tc.QuadQuote = tc.Quote.render(`""`)
	tc.QuotedNull = tc.QuotedString.render("null")
	tc.EscTab = tc.Escape.render(`\t`)
	tc.EscCR = tc.Escape.render(`\r`)
	tc.EscLF = tc.Escape.render(`\n`)
	tc.EscBackslash = tc.Escape.render(`\`) + `\`
	tc.EscQuote = tc.Escape.render(`\`) + `"`
	tc.ArraySep = styleFromTheme(theme.Array.Sep1).render(",")
	tc.MapSep1 = styleFromTheme(theme.Map.Sep1).render(",")
	tc.MapSep2 = styleFromTheme(theme.Map.Sep2).render(":")
	tc.EmptyArray = strings.TrimSpace(sc.Array.prefix) + strings.TrimSpace(sc.Array.suffix)
	tc.EmptyMap = strings.TrimSpace(sc.Map.prefix) + strings.TrimSpace(sc.Map.suffix)
	tc.Null = styleFromTheme(theme.NullValue).render("null")
	for i := 0; i < 4; i++ {
		tc.LevelLabel[i] = styleFromTheme(theme.Level[i]).withTrailingSpace().render(levelLabels[i])
	}

	return tc
}

type themeCache struct {
	styleCache
	DoubleQuote  string
	QuadQuote    string
	QuotedNull   string
	EscTab       string
	EscCR        string
	EscLF        string
	EscBackslash string
	EscQuote     string
	LevelLabel   [NumLevels]string
	EmptyArray   string
	ArraySep     string
	EmptyMap     string
	MapSep1      string
	MapSep2      string
	Null         string
}

// ---

type groupKeys struct {
	head    [4]string
	tail    []string
	headLen int
}

func (g *groupKeys) append(key string) {
	if g.headLen < len(g.head) {
		g.head[g.headLen] = key
		g.headLen++
	} else {
		g.tail = append(g.tail, key)
	}
}

func (g *groupKeys) collect(buf []string) []string {
	buf = append(buf, g.head[:g.headLen]...)
	buf = append(buf, g.tail...)

	return buf
}

func (g groupKeys) fork() groupKeys {
	g.tail = slices.Clip(g.tail)

	return g
}

// ---

type groups struct {
	head    [4]group
	tail    []group
	headLen int
}

func (g *groups) len() int {
	return g.headLen + len(g.tail)
}

func (g *groups) append(group group) {
	if g.headLen < len(g.head) {
		g.head[g.headLen] = group
		g.headLen++
	} else {
		g.tail = append(g.tail, group)
	}
}

func (g *groups) at(i int) group {
	if i < g.headLen {
		return g.head[i]
	}

	return g.tail[i-g.headLen]
}

func (g groups) fork() groups {
	g.tail = slices.Clip(g.tail)

	return g
}

// ---

func newStyle(prefix, suffix string) style {
	return style{prefix, suffix, prefix == "" && suffix == ""}
}

func styleFromTheme(item ThemeItem) style {
	return style{item.Prefix, item.Suffix, item.IsEmpty()}
}

type style struct {
	prefix string
	suffix string
	empty  bool
}

func (s *style) set(prefix, suffix string) {
	s.prefix = prefix
	s.suffix = suffix
	s.empty = prefix == "" && suffix == ""
}

func (s style) withTrailingSpace() style {
	return s.withExtraSuffix(" ")
}

func (s style) withExtraSuffix(suffix string) style {
	s.set(s.prefix, s.suffix+suffix)

	return s
}

func (s *style) open(hs *handleState) {
	hs.buf.AppendString(s.prefix)
}

func (s *style) close(hs *handleState) {
	hs.buf.AppendString(s.suffix)
}

func (s *style) apply(hs *handleState, appendValue func()) {
	if s.empty {
		appendValue()
	} else {
		s.open(hs)
		appendValue()
		s.close(hs)
	}
}

func (s style) render(inner string) string {
	return s.prefix + inner + s.suffix
}

// ---

func safeResolveValue[T any](h *Handler, hs *handleState, resolve func() T) (_ T, ok bool) {
	defer func() {
		if p := recover(); p != nil {
			h.appendEncodePanic(hs, p)
			ok = false
		}
	}()

	return resolve(), true
}

func safeResolveValueErr[T any](h *Handler, hs *handleState, resolve func() (T, error)) (value T, ok bool) {
	var err error

	defer func() {
		if p := recover(); p != nil {
			h.appendEncodePanic(hs, p)
		} else if err != nil {
			h.appendEncodeError(hs, err)
		}
	}()

	value, err = resolve()

	return value, err == nil
}

// ---

type group struct {
	i         int
	prefixLen int
}

type cache struct {
	attrs     string
	numGroups int
	numAttrs  int
	once      sync.Once
}

// ---

var levelLabels = [4]string{"DBG", "INF", "WRN", "ERR"}
var levelNames = [4]string{"DEBUG", "INFO", "WARN", "ERROR"}
var levels = [4]slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

func levelIndex(level slog.Level) int {
	return min(max(0, (int(level)+4)/4), 3)
}

// ---

const hexDigits = "0123456789abcdef"

var _ slog.Handler = (*Handler)(nil)
