// Package slogtext provides slog.Handler implementation that output log messages in a textual human-readable form with colors.
package slogtext

import (
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
	"unsafe"

	"github.com/pamburus/slogx/internal/quoting"
	"github.com/pamburus/slogx/internal/stylecache"
	"github.com/pamburus/slogx/internal/syntax"
	"github.com/pamburus/slogx/internal/tty"
	"github.com/pamburus/slogx/slogtext/themes"
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
			*stylecache.New(&opts.theme, stylecache.DefaultConfig()),
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
		hs.buf.AppendString(h.stc.Time.Prefix)
		if replace == nil {
			h.appendTimestamp(hs, record.Time)
		} else if attr := replace(nil, slog.Time(slog.TimeKey, val)); attr.Key != "" {
			if attr.Value.Kind() == slog.KindTime {
				h.appendTimestamp(hs, attr.Value.Time())
			} else {
				h.appendValue(hs, attr.Value, false, false)
			}
		}
		hs.buf.AppendString(h.stc.Time.Suffix)
	}

	h.appendLevel(hs, record.Level)

	if replace != nil {
		if attr := replace(nil, slog.String(slog.MessageKey, record.Message)); attr.Key != "" {
			switch {
			case attr.Value.Kind() == slog.KindString:
				record.Message = attr.Value.String()
			default:
				record.Message = ""
				h.appendValue(hs, attr.Value, false, false)
			}
		}
	}

	hs.messageBegin = hs.buf.Len()

	if record.Message != "" {
		if quoting.MessageContext().IsNeeded(record.Message) {
			if !h.appendQuotedString(hs, &h.stc.Message, record.Message, true) {
				hs.addAttrToExpand(slog.String(slog.MessageKey, record.Message))
			}
		} else {
			h.appendUnquotedString(hs, &h.stc.Message, record.Message)
		}
	}

	h.appendHandlerAttrs(hs)

	record.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(hs, attr, len(h.keyPrefix))

		return true
	})

	if h.includeSource {
		src := h.source(record.PC)
		if src.File != "" {
			hs.buf.AppendString(h.stc.Source.Prefix)
			h.appendSource(hs, src)
			hs.buf.AppendString(h.stc.Source.Suffix)
		}
	}

	if len(hs.attrsToExpand) != 0 {
		hs.buf.TrimBackByte(' ')
		hs.buf.AppendString(h.stc.ExpandedMessageSign)
		hs.expandingAttrs = true
		for _, attr := range hs.attrsToExpand {
			hs.buf.AppendByte('\n')
			hs.buf.AppendBytes(hs.buf[:hs.messageBegin])
			hs.buf.AppendString(h.stc.ExpandedKey.Prefix)
			hs.buf.AppendString(h.keyPrefix)
			hs.buf.AppendString(attr.KeyPrefix)
			hs.buf.AppendString(attr.Key)
			hs.buf.AppendString(h.stc.ExpandedKey.Suffix)
			hs.buf.AppendByte('\n')
			h.appendValue(hs, attr.Value, false, false)
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
		h.appendAutoQuotedString(hs, &h.stc.StringValue, unsafe.String(&hs.scratch[0], hs.scratch.Len()), false)
	} else {
		hs.buf.AppendBytes(hs.scratch.Bytes())
	}
}

func (h *Handler) appendDuration(hs *handleState, value time.Duration, quote bool) {
	hs.scratch.Reset()
	hs.scratch = h.encodeDuration(hs.scratch, value)
	if quote {
		h.appendAutoQuotedString(hs, &h.stc.StringValue, unsafe.String(&hs.scratch[0], hs.scratch.Len()), false)
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
		if !h.appendValue(hs, attr.Value, true, !hs.expandingAttrs) {
			hs.addAttrToExpand(attr)
		}
		hs.buf.AppendByte(' ')
	}
}

func (h *Handler) appendKey(hs *handleState, key string, basePrefixLen int) {
	hs.buf.AppendString(h.stc.Key.Prefix)
	hs.buf.AppendString(h.keyPrefix[:basePrefixLen])
	hs.buf.AppendBytes(hs.keyPrefix)
	hs.buf.AppendString(key)
	hs.buf.AppendString(h.stc.Key.Suffix)
}

func (h *Handler) appendValue(hs *handleState, v slog.Value, quote bool, breakOnNewLine bool) bool {
	switch v.Kind() {
	case slog.KindString:
		return h.appendString(hs, &h.stc.StringValue, v.String(), quote, breakOnNewLine)
	case slog.KindInt64:
		hs.buf.AppendString(h.stc.NumberValue.Prefix)
		hs.buf.AppendInt(v.Int64())
		hs.buf.AppendString(h.stc.NumberValue.Suffix)
	case slog.KindUint64:
		hs.buf.AppendString(h.stc.NumberValue.Prefix)
		hs.buf.AppendUint(v.Uint64())
		hs.buf.AppendString(h.stc.NumberValue.Suffix)
	case slog.KindFloat64:
		hs.buf.AppendString(h.stc.NumberValue.Prefix)
		hs.buf.AppendFloat64(v.Float64())
		hs.buf.AppendString(h.stc.NumberValue.Suffix)
	case slog.KindBool:
		hs.buf.AppendString(h.stc.BoolValue.Prefix)
		hs.buf.AppendBool(v.Bool())
		hs.buf.AppendString(h.stc.BoolValue.Suffix)
	case slog.KindDuration:
		hs.buf.AppendString(h.stc.DurationValue.Prefix)
		h.appendDuration(hs, v.Duration(), quote)
		hs.buf.AppendString(h.stc.DurationValue.Suffix)
	case slog.KindGroup:
		attrs := v.Group()
		if len(attrs) == 0 {
			hs.buf.AppendString(h.stc.EmptyMap)
		} else {
			hs.buf.AppendString(h.stc.Map.Prefix)
			for i, attr := range attrs {
				if i != 0 {
					hs.buf.AppendString(h.stc.MapPairSep)
				}
				h.appendString(hs, &h.stc.StringValue, attr.Key, quote, false)
				hs.buf.AppendString(h.stc.MapKeyValueSep)
				h.appendValue(hs, attr.Value, true, false)
			}
			hs.buf.AppendString(h.stc.Map.Suffix)
		}
	case slog.KindTime:
		hs.buf.AppendString(h.stc.TimeValue.Prefix)
		h.appendTime(hs, v.Time(), quote)
		hs.buf.AppendString(h.stc.TimeValue.Suffix)
	case slog.KindAny:
		switch v := v.Any().(type) {
		case nil:
			hs.buf.AppendString(h.stc.Null)
		case slog.Level:
			h.appendLevelValue(hs, v)
		case error:
			return h.appendString(hs, &h.stc.ErrorValue, v.Error(), quote, breakOnNewLine)
		case fmt.Stringer:
			if v, ok, errorAppended := safeResolveValue(h, hs, v.String); ok {
				return h.appendString(hs, &h.stc.StringValue, v, quote, breakOnNewLine)
			} else {
				return errorAppended
			}
		case encoding.TextMarshaler:
			if data, ok, errorAppended := safeResolveValueErr(h, hs, v.MarshalText); ok {
				h.appendString(hs, &h.stc.StringValue, unsafe.String(&data[0], len(data)), quote, breakOnNewLine)
			} else {
				return errorAppended
			}
		case *slog.Source:
			h.appendSource(hs, *v)
		case slog.Source:
			h.appendSource(hs, v)
		case []byte:
			return h.appendBytesValue(hs, v, quote, breakOnNewLine)
		default:
			h.appendAnyValue(hs, v, quote)
		}
	}

	return true
}

func (h *Handler) appendBytesValue(hs *handleState, v []byte, quote bool, breakOnNewLine bool) bool {
	switch h.bytesFormat {
	default:
		fallthrough
	case BytesFormatString:
		return h.appendString(hs, &h.stc.StringValue, unsafe.String(&v[0], len(v)), quote, breakOnNewLine)
	case BytesFormatHex:
		hs.buf.AppendString(h.stc.StringValue.Quoted.Prefix)
		hex.Encode(hs.buf.Extend(hex.EncodedLen(len(v))), v)
		hs.buf.AppendString(h.stc.StringValue.Quoted.Suffix)
	case BytesFormatBase64:
		hs.buf.AppendString(h.stc.StringValue.Quoted.Prefix)
		base64.StdEncoding.Encode(hs.buf.Extend(base64.StdEncoding.EncodedLen(len(v))), v)
		hs.buf.AppendString(h.stc.StringValue.Quoted.Suffix)
	}

	return true
}

func (h *Handler) appendAnyValue(hs *handleState, v any, quote bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		switch {
		case rv.IsNil():
			hs.buf.AppendString(h.stc.Null)
		case rv.Len() == 0:
			hs.buf.AppendString(h.stc.EmptyArray)
		default:
			hs.buf.AppendString(h.stc.Array.Prefix)
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					hs.buf.AppendString(h.stc.ArraySep)
				}
				h.appendValue(hs, slog.AnyValue(rv.Index(i).Interface()), true, false)
			}
			hs.buf.AppendString(h.stc.Array.Suffix)
		}
	case reflect.Map:
		switch {
		case rv.IsNil():
			hs.buf.AppendString(h.stc.Null)
		case rv.Len() == 0:
			hs.buf.AppendString(h.stc.EmptyMap)
		default:
			hs.buf.AppendString(h.stc.Map.Prefix)
			for i, k := range rv.MapKeys() {
				if i != 0 {
					hs.buf.AppendString(h.stc.MapPairSep)
				}
				h.appendValue(hs, slog.AnyValue(k.Interface()), true, false)
				hs.buf.AppendString(h.stc.MapKeyValueSep)
				h.appendValue(hs, slog.AnyValue(rv.MapIndex(k).Interface()), true, false)
			}
			hs.buf.AppendString(h.stc.Map.Suffix)
		}
	default:
		hs.scratch.Reset()
		_, _ = fmt.Fprintf(&hs.scratch, "%+v", v)
		h.appendString(hs, &h.stc.StringValue, unsafe.String(&hs.scratch[0], len(hs.scratch)), quote, false)
	}
}

func (h *Handler) appendString(hs *handleState, ss *stylecache.StringStyle, s string, quote bool, breakOnNewLine bool) bool {
	switch {
	case hs.expandingAttrs:
		s = strings.TrimSpace(s)
		for {
			i := strings.IndexByte(s, '\n')
			if i == -1 {
				i = len(s)
			}
			hs.buf.AppendBytes(hs.buf[:hs.messageBegin])
			hs.buf.AppendString(syntax.ExpandedValuePrefix)
			hs.buf.AppendString(ss.Unquoted.Prefix)
			hs.buf.AppendString(s[:i])
			hs.buf.AppendString(ss.Unquoted.Suffix)
			hs.buf.AppendByte('\n')
			if i < len(s) {
				s = s[i+1:]
			} else {
				break
			}
		}
	case quote:
		if s == "null" {
			hs.buf.AppendString(ss.Null)
		} else {
			return h.appendAutoQuotedString(hs, ss, s, breakOnNewLine)
		}
	default:
		hs.buf.AppendString(s)
	}

	return true
}

func (h *Handler) appendAutoQuotedString(hs *handleState, ss *stylecache.StringStyle, v string, breakOnNewLine bool) bool {
	switch {
	case len(v) == 0:
		hs.buf.AppendString(ss.Empty)
	case quoting.StringValueContext().IsNeeded(v):
		return h.appendQuotedString(hs, ss, v, breakOnNewLine)
	default:
		hs.buf.AppendString(v)
	}

	return true
}

func (h *Handler) appendUnquotedString(hs *handleState, ss *stylecache.StringStyle, v string) {
	hs.buf.AppendString(ss.Unquoted.Prefix)
	hs.buf.AppendString(v)
	hs.buf.AppendString(ss.Unquoted.Suffix)
}

func (h *Handler) appendQuotedString(hs *handleState, ss *stylecache.StringStyle, v string, breakOnNewLine bool) bool {
	hs.buf.AppendString(ss.Quoted.Prefix)
	done := h.appendEscapedString(hs, ss, v, breakOnNewLine)
	hs.buf.AppendString(ss.Quoted.Suffix)
	if !done {
		hs.buf.TrimBackByte(' ')
		hs.buf.AppendString(ss.Elipsis)
	}

	return done
}

func (h *Handler) appendEscapedString(hs *handleState, ss *stylecache.StringStyle, s string, breakOnNewLine bool) bool {
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
				hs.buf.AppendString(ss.Escape.Tab)
			case '\r':
				hs.buf.AppendString(ss.Escape.CR)
			case '\n':
				hs.buf.AppendString(ss.Escape.LF)
				if breakOnNewLine && i > int(h.expansionThreshold) {
					return false
				}
			case '\\':
				hs.buf.AppendString(ss.Escape.Backslash)
			case '"':
				hs.buf.AppendString(ss.Escape.Quote)
			default:
				hs.buf.AppendString(ss.Escape.Style.Prefix)
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hexDigits[c>>4])
				hs.buf.AppendByte(hexDigits[c&0xf])
				hs.buf.AppendString(ss.Escape.Style.Suffix)
			}
			i++
			p = i

		default:
			v, wd := utf8.DecodeRuneInString(s[i:])
			if v == utf8.RuneError && wd == 1 {
				hs.buf.AppendString(s[p:i])
				hs.buf.AppendString(ss.Escape.Style.Prefix)
				hs.buf.AppendString(`\ufffd`)
				hs.buf.AppendString(ss.Escape.Style.Suffix)
				i++
				p = i
			} else {
				i += wd
			}
		}
	}

	hs.buf.AppendString(s[p:])

	return true
}

func (h *Handler) appendSource(hs *handleState, source slog.Source) {
	hs.buf = h.encodeSource(hs.buf, source)
}

func (h *Handler) appendLevel(hs *handleState, level slog.Level) {
	if h.levelOffset {
		h.appendLevelWithOffset(hs, level)
	} else {
		hs.buf.AppendString(h.stc.LevelLabel[levelIndex(level)])
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
	hs.buf.AppendString(h.stc.LevelValue[i].Prefix)
	offset := int64(level - levels[i])
	if offset != 0 {
		hs.buf.AppendString(h.stc.Config.LevelLabels[i][:1])
		appendOffset(offset)
	} else {
		hs.buf.AppendString(h.stc.Config.LevelLabels[i])
	}
	hs.buf.AppendString(h.stc.LevelValue[i].Suffix)
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
	hs.buf.AppendString(h.stc.LevelValue[i].Prefix)
	hs.buf.AppendString(h.stc.Config.LevelNames[i])
	appendOffset(int64(level - levels[i]))
	hs.buf.AppendString(h.stc.LevelValue[i].Suffix)
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

func (h *Handler) appendEvalError(hs *handleState, err error) bool {
	if hs.expandingAttrs {
		return h.appendString(hs, &h.stc.ErrorValue, err.Error(), false, false)
	}

	hs.buf.AppendString(h.stc.EvalError.Prefix)
	done := h.appendQuotedString(hs, &h.stc.ErrorValue, err.Error(), true)
	hs.buf.AppendString(h.stc.EvalError.Suffix)
	return done
}

func (h *Handler) appendEvalPanic(hs *handleState, p any) bool {
	hs.scratch.Reset()
	_, _ = fmt.Fprintf(&hs.scratch, "%v", p)

	if hs.expandingAttrs {
		return h.appendString(hs, &h.stc.ErrorValue, unsafe.String(&hs.scratch[0], len(hs.scratch)), false, false)
	}

	hs.buf.AppendString(h.stc.EvalError.Prefix)
	done := h.appendQuotedString(hs, &h.stc.ErrorValue, unsafe.String(&hs.scratch[0], len(hs.scratch)), true)
	hs.buf.AppendString(h.stc.EvalError.Suffix)

	return done
}

// ---

type shared struct {
	options
	stc    stylecache.StyleCache
	mu     sync.Mutex
	writer io.Writer
}

// ---

type groupKeys struct {
	head    [numEmbeddedGroups]string
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
	head    [numEmbeddedGroups]group
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

func safeResolveValue[T any](h *Handler, hs *handleState, resolve func() T) (_ T, resolved, errorAppended bool) {
	defer func() {
		if p := recover(); p != nil {
			errorAppended = h.appendEvalPanic(hs, p)
			resolved = false
		}
	}()

	return resolve(), true, true
}

func safeResolveValueErr[T any](h *Handler, hs *handleState, resolve func() (T, error)) (value T, resolved, errorAppended bool) {
	var err error

	defer func() {
		if p := recover(); p != nil {
			errorAppended = h.appendEvalPanic(hs, p)
		} else if err != nil {
			errorAppended = h.appendEvalError(hs, err)
		}
	}()

	value, err = resolve()

	return value, err == nil, true
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

var levels = [numLevels]slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

func levelIndex(level slog.Level) int {
	return min(max(0, (int(level)+4)/4), numLevels-1)
}

// ---

const hexDigits = "0123456789abcdef"
const numLevels = themes.NumLevels
const numEmbeddedGroups = 4

var _ slog.Handler = (*Handler)(nil)
