// Package slogtext provides slog.Handler implementation that output log messages in a textual human-readable form with colors.
package slogjson

import (
	"context"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/pamburus/slogx/slogtext/themes"
)

// NewHandler returns a new slog.Handler with the given Writer and
// optional custom configuration.
func NewHandler(writer io.Writer, options ...Option) *Handler {
	opts := defaultOptions().with(options)

	return &Handler{
		shared: &shared{
			opts,
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
	cache     cache
}

// Enabled returns true if the given level is enabled.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.replaceLevel(ctx, level) >= h.leveler.Level()
}

// Handle writes the log message to the Writer.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	hs := newHandleState(ctx, h)
	defer hs.release()

	hs.buf.AppendByte('{')

	record.Level = h.replaceLevel(ctx, record.Level)
	replace := h.replaceAttr

	if !record.Time.IsZero() {
		value := record.Time.Round(0)
		if replace != nil {
			attr := replace(nil, slog.Time(slog.TimeKey, value))
			if attr.Key != "" {
				attr.Value = attr.Value.Resolve()
				switch attr.Value.Kind() {
				case slog.KindTime:
					value = attr.Value.Time()
				case slog.KindString:
					h.appendString(hs, attr.Value.String())
				case slog.KindUint64:
					value = time.Unix(0, int64(attr.Value.Uint64()))
				case slog.KindInt64:
					value = time.Unix(0, attr.Value.Int64())
				default:
					value = time.Time{}
				}
			} else {
				value = time.Time{}
			}
		}
		if !value.IsZero() {
			value := h.encodeTimestamp(value)
			if value.Kind() != slog.KindTime {
				h.appendAttr(hs, slog.Attr{Key: slog.TimeKey, Value: value})
			}
		}
	}

	h.appendString(hs, slog.LevelKey)
	hs.buf.AppendByte(':')
	h.appendLevel(hs, record.Level)
	hs.buf.AppendByte(',')

	messageAttr := slog.String(slog.MessageKey, record.Message)

	if replace != nil {
		if attr := replace(nil, slog.String(slog.MessageKey, record.Message)); attr.Key != "" {
			attr.Value = attr.Value.Resolve()
			messageAttr = attr
		}
	}

	h.appendAttr(hs, messageAttr)

	h.appendHandlerAttrs(hs)

	record.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(hs, attr)

		return true
	})

	if h.includeSource {
		src := h.source(hs, record.PC)
		if src.File != "" {
			h.appendString(hs, slog.SourceKey)
			hs.buf.AppendByte(':')
			h.appendSource(hs, src)
		}
	}

	hs.buf.TrimBackByte(',')
	h.closeGroups(hs, len(hs.groups)+1)

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

	h.groups.append(group{len(h.attrs)})
	h.groupKeys.append(key)

	return h
}

func (h *Handler) fork() *Handler {
	return &Handler{
		h.shared,
		slices.Clip(h.attrs),
		h.groups.fork(),
		h.groupKeys.fork(),
		h.cache.fork(h),
	}
}
func (h *Handler) appendTime(hs *handleState, value time.Time) {
	resolvedValue := h.encodeTimestamp(value).Resolve()
	if resolvedValue.Kind() != slog.KindTime {
		h.appendValue(hs, resolvedValue)
	} else {
		hs.buf.AppendString("null")
	}
}

func (h *Handler) appendDuration(hs *handleState, value time.Duration) {
	resolvedValue := h.encodeDuration(value).Resolve()
	if resolvedValue.Kind() != slog.KindDuration {
		h.appendValue(hs, resolvedValue)
	} else {
		hs.buf.AppendString("null")
	}
}

func (h *Handler) appendHandlerAttrs(hs *handleState) {
	if h == nil || len(h.attrs) == 0 && h.groups.len() == 0 {
		return
	}

	appended := false

	h.cache.once.Do(func() {
		if h.hasUncachedAttrs() {
			appended = h.initAttrCache(hs)
		}
	})

	if !appended && len(h.cache.attrs) != 0 {
		hs.buf.AppendString(h.cache.attrs)
	}
}

// hasUncachedAttrs must be called under the cache.once lock.
func (h *Handler) hasUncachedAttrs() bool {
	return h.cache.numAttrs != len(h.attrs) || h.cache.numGroups != h.groups.len()
}

// initAttrCache must be called under the cache.once lock.
func (h *Handler) initAttrCache(hs *handleState) bool {
	pos := hs.buf.Len()
	hs.buf.AppendString(h.cache.attrs)

	begin := h.cache.numAttrs
	for i := h.cache.numGroups; i != h.groups.len(); i++ {
		group := h.groups.at(i)
		end := group.i
		h.openGroup(hs, h.groupKeys.at(i))
		for _, attr := range h.attrs[begin:end] {
			h.appendAttr(hs, attr)
		}
		begin = end
	}

	for _, attr := range h.attrs[begin:] {
		h.appendAttr(hs, attr)
	}

	h.cache.attrs = hs.buf[pos:].String()
	h.cache.numGroups = h.groups.len()
	h.cache.numAttrs = len(h.attrs)

	return true
}

func (h *Handler) openGroup(hs *handleState, key string) {
	h.appendString(hs, key)
	hs.buf.AppendByte(':')
	hs.buf.AppendByte('{')
}

func (h *Handler) closeGroups(hs *handleState, n int) {
	for n > 0 {
		k := min(n, len(groupCloseChain))
		hs.buf.AppendString(groupCloseChain[:k])
		n -= k
	}
}

func (h *Handler) appendAttr(hs *handleState, attr slog.Attr) {
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
			hs.groups = append(hs.groups, attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(hs, groupAttr)
		}
		if attr.Key != "" {
			hs.groups = hs.groups[:len(hs.groups)-1]
		}
	} else {
		h.appendKey(hs, attr.Key)
		hs.buf.AppendByte(':')
		h.appendValue(hs, attr.Value)
		hs.buf.AppendByte(',')
	}
}

func (h *Handler) appendKey(hs *handleState, key string) {
	h.appendString(hs, key)
}

func (h *Handler) appendValue(hs *handleState, v slog.Value) bool {
	switch v.Kind() {
	case slog.KindString:
		return h.appendString(hs, v.String())
	case slog.KindInt64:
		hs.buf.AppendInt(v.Int64())
	case slog.KindUint64:
		hs.buf.AppendUint(v.Uint64())
	case slog.KindFloat64:
		hs.buf.AppendFloat64(v.Float64())
	case slog.KindBool:
		hs.buf.AppendBool(v.Bool())
	case slog.KindDuration:
		h.appendDuration(hs, v.Duration())
	case slog.KindGroup:
		attrs := v.Group()
		hs.buf.AppendByte('{')
		for i, attr := range attrs {
			if i != 0 {
				hs.buf.AppendByte(',')
			}
			h.appendString(hs, attr.Key)
			hs.buf.AppendByte(':')
			h.appendValue(hs, attr.Value)
		}
		hs.buf.AppendByte('}')
	case slog.KindTime:
		h.appendTime(hs, v.Time())
	case slog.KindAny:
		switch v := v.Any().(type) {
		case nil:
			hs.buf.AppendString("null")
		case slog.Level:
			h.appendLevel(hs, v)
		case error:
			return h.appendString(hs, v.Error())
		case fmt.Stringer:
			if v, ok := safeResolveValue(h, hs, v.String); ok {
				return h.appendString(hs, v)
			}
		case encoding.TextMarshaler:
			if data, ok := safeResolveValueErr(h, hs, v.MarshalText); ok {
				h.appendString(hs, unsafe.String(&data[0], len(data)))
				runtime.KeepAlive(&data)
			}
		case *slog.Source:
			h.appendSource(hs, *v)
		case slog.Source:
			h.appendSource(hs, v)
		case []byte:
			return h.appendBytesValue(hs, v)
		default:
			h.appendAnyValue(hs, v)
		}
	}

	return true
}

func (h *Handler) appendBytesValue(hs *handleState, v []byte) bool {
	switch h.bytesFormat {
	default:
		fallthrough
	case BytesFormatString:
		return h.appendString(hs, unsafe.String(&v[0], len(v)))
	case BytesFormatHex:
		hex.Encode(hs.buf.Extend(hex.EncodedLen(len(v))), v)
	case BytesFormatBase64:
		base64.StdEncoding.Encode(hs.buf.Extend(base64.StdEncoding.EncodedLen(len(v))), v)
	}

	return true
}

func (h *Handler) appendAnyValue(hs *handleState, v any) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		switch {
		case rv.IsNil():
			hs.buf.AppendString("null")
		default:
			hs.buf.AppendByte('[')
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					hs.buf.AppendByte(',')
				}
				h.appendValue(hs, slog.AnyValue(rv.Index(i).Interface()))
			}
			hs.buf.AppendByte(']')
		}
	case reflect.Map:
		switch {
		case rv.IsNil():
			hs.buf.AppendString("null")
		default:
			hs.buf.AppendByte('{')
			for i, k := range rv.MapKeys() {
				if i != 0 {
					hs.buf.AppendByte(',')
				}
				if k.Kind() == reflect.String {
					h.appendString(hs, k.String())
				} else {
					hs.scratch.Reset()
					_, _ = fmt.Fprintf(&hs.scratch, "%v", k.Interface())
					h.appendString(hs, unsafe.String(&hs.scratch[0], hs.scratch.Len()))
				}
				hs.buf.AppendByte(':')
				h.appendValue(hs, slog.AnyValue(rv.MapIndex(k).Interface()))
			}
			hs.buf.AppendByte('}')
		}
	default:
		hs.scratch.Reset()
		_, _ = fmt.Fprintf(&hs.scratch, "%v", v)
		h.appendString(hs, unsafe.String(&hs.scratch[0], len(hs.scratch)))
	}
}

func (h *Handler) appendString(hs *handleState, s string) bool {
	hs.buf.AppendByte('"')
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
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('t')
			case '\r':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('r')
			case '\n':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('n')
			case '\\':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('\\')
			case '"':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('"')
			default:
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hexDigits[c>>4])
				hs.buf.AppendByte(hexDigits[c&0xf])
			}
			i++
			p = i

		default:
			v, wd := utf8.DecodeRuneInString(s[i:])
			if v == utf8.RuneError && wd == 1 {
				hs.buf.AppendString(s[p:i])
				hs.buf.AppendString(`\ufffd`)
				i++
				p = i
			} else {
				i += wd
			}
		}
	}

	hs.buf.AppendString(s[p:])
	hs.buf.AppendByte('"')

	return true
}

func (h *Handler) appendSource(hs *handleState, source slog.Source) {
	h.appendValue(hs, h.encodeSource(source))
}

func (h *Handler) appendLevel(hs *handleState, level slog.Level) {
	appendOffset := func(offset int64) {
		if offset != 0 {
			if offset > 0 {
				hs.buf.AppendByte('+')
			}
			hs.buf.AppendInt(offset)
		}
	}

	i := levelIndex(level)
	hs.buf.AppendByte('"')
	hs.buf.AppendString(levelNames[i])
	appendOffset(int64(level - levels[i]))
	hs.buf.AppendByte('"')
}

func (h *Handler) source(hs *handleState, pc uintptr) slog.Source {
	frame, _ := runtime.CallersFrames(hs.pcs[:]).Next()

	return slog.Source{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}
}

func (h *Handler) appendEvalError(hs *handleState, err error) {
	h.appendString(hs, err.Error())
}

func (h *Handler) appendEvalPanic(hs *handleState, p any) {
	hs.scratch.Reset()
	_, _ = fmt.Fprintf(&hs.scratch, "%v", p)

	h.appendString(hs, unsafe.String(&hs.scratch[0], len(hs.scratch)))
}

// ---

type shared struct {
	options
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

func (g *groupKeys) at(i int) string {
	if i < g.headLen {
		return g.head[i]
	}

	return g.tail[i-g.headLen]
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

func safeResolveValue[T any](h *Handler, hs *handleState, resolve func() T) (_ T, resolved bool) {
	defer func() {
		if p := recover(); p != nil {
			h.appendEvalPanic(hs, p)
			resolved = false
		}
	}()

	return resolve(), true
}

func safeResolveValueErr[T any](h *Handler, hs *handleState, resolve func() (T, error)) (value T, resolved bool) {
	var err error

	defer func() {
		if p := recover(); p != nil {
			h.appendEvalPanic(hs, p)
		} else if err != nil {
			h.appendEvalError(hs, err)
		}
	}()

	value, err = resolve()

	return value, err == nil
}

// ---

type group struct {
	i int
}

// cache must be modified only under the once lock.
// cache must be read either under the once lock or after the once lock is released.
type cache struct {
	attrs     string
	numGroups int
	numAttrs  int
	once      sync.Once
}

func (c *cache) fork(h *Handler) cache {
	c.once.Do(func() {
		if h.hasUncachedAttrs() {
			hs := newHandleState(context.Background(), h)
			h.initAttrCache(hs)
			hs.release()
		}
	})

	return cache{
		attrs:     c.attrs,
		numGroups: c.numGroups,
		numAttrs:  c.numAttrs,
	}
}

// ---

var levels = [numLevels]slog.Level{
	slog.LevelDebug,
	slog.LevelInfo,
	slog.LevelWarn,
	slog.LevelError,
}

var levelNames = [numLevels]string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
}

func levelIndex(level slog.Level) int {
	return min(max(0, (int(level)+4)/4), numLevels-1)
}

// ---

const hexDigits = "0123456789abcdef"
const numLevels = themes.NumLevels
const numEmbeddedGroups = 4
const groupCloseChain = "}}}}}}}}}}}}}}}}"

var _ slog.Handler = (*Handler)(nil)
