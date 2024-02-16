// Package slogtext provides slog.Handler implementation that output log messages in a textual human-readable form with colors.
package slogtext

import (
	"bytes"
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"slices"
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
		opts.theme = ThemePlain()
	}

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

	rep := h.replaceAttr

	if !record.Time.IsZero() {
		val := record.Time.Round(0)
		if rep == nil {
			h.appendTimestamp(hs, record.Time)
		} else if attr := rep(nil, slog.Time(slog.TimeKey, val)); attr.Key != "" {
			if attr.Value.Kind() == slog.KindTime {
				h.appendTimestamp(hs, attr.Value.Time())
			} else {
				h.appendValue(hs, attr.Value, false)
			}
		}
		hs.buf.AppendByte(' ')
	}

	if rep == nil {
		h.appendLevel(hs, record.Level)
	} else if attr := rep(nil, slog.Any(slog.LevelKey, record.Level)); attr.Key != "" {
		h.appendValue(hs, attr.Value, false)
	}
	hs.buf.AppendByte(' ')

	if rep == nil {
		hs.buf.AppendString(record.Message)
	} else if a := rep(nil, slog.String(slog.MessageKey, record.Message)); a.Key != "" {
		h.appendValue(hs, a.Value, false)
	}
	hs.buf.AppendByte(' ')

	h.appendHandlerAttrs(hs)

	record.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(hs, attr, len(hs.h.keyPrefix))
		return true
	})

	if h.includeSource {
		src := h.source(record.PC)
		if src.File != "" {
			hs.buf.AppendString(h.theme.Source.Prefix)
			if rep == nil {
				h.appendSource(hs, src)
			} else if attr := rep(nil, slog.Any(slog.SourceKey, src)); attr.Key != "" {
				h.appendValue(hs, attr.Value, false)
			}
			hs.buf.AppendString(h.theme.Source.Suffix)
			hs.buf.AppendByte(' ')
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
	hs.buf.AppendString(hs.h.theme.Timestamp.Prefix)
	hs.buf = h.encodeTimestamp(hs.buf, value)
	hs.buf.AppendString(hs.h.theme.Timestamp.Suffix)
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

	if !appended && len(hs.h.cache.attrs) != 0 {
		hs.buf.AppendString(hs.h.cache.attrs)
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
	hs.buf.AppendString(hs.h.theme.Key.Prefix)
	hs.buf.AppendString(hs.h.keyPrefix[:basePrefixLen])
	hs.buf.AppendBytes(hs.keyPrefix)
	hs.buf.AppendString(key)
	hs.buf.AppendString(hs.h.theme.Key.Suffix)
}

func (h *Handler) appendValue(hs *handleState, v slog.Value, quote bool) {
	switch v.Kind() {
	case slog.KindString:
		h.appendString(hs, v.String(), quote)
	case slog.KindInt64:
		hs.buf.AppendInt(v.Int64())
	case slog.KindUint64:
		hs.buf.AppendUint(v.Uint64())
	case slog.KindFloat64:
		hs.buf.AppendFloat64(v.Float64())
	case slog.KindBool:
		hs.buf.AppendBool(v.Bool())
	case slog.KindDuration:
		h.appendDuration(hs, v.Duration(), quote)
	case slog.KindTime:
		h.appendTime(hs, v.Time(), quote)
	case slog.KindAny:
		switch v := v.Any().(type) {
		case slog.Level:
			h.appendLevel(hs, v)
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err != nil {
				break
			}
			h.appendByteString(hs, data, quote)
		case *slog.Source:
			h.appendSource(hs, *v)
		case slog.Source:
			h.appendSource(hs, v)
		default:
			hs.scratch.Reset()
			_, _ = fmt.Fprintf(&hs.scratch, "%+v", v)
			h.appendByteString(hs, hs.scratch.Bytes(), quote)
		}
	}
}

func (h *Handler) appendString(hs *handleState, s string, quote bool) {
	if quote {
		if s == "null" {
			hs.buf.AppendString(`"null"`)
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
			hs.buf.AppendString(`"null"`)
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
		hs.buf.AppendString(`""`)
	case quoting.IsNeeded(v):
		hs.buf.AppendByte('"')
		h.appendEscapedString(hs, v)
		hs.buf.AppendByte('"')
	default:
		hs.buf.AppendString(v)
	}
}

func (h *Handler) appendAutoQuotedByteString(hs *handleState, v []byte) {
	switch {
	case len(v) == 0:
		hs.buf.AppendString(`""`)
	case quoting.IsNeededForBytes(v):
		hs.buf.AppendByte('"')
		h.appendEscapedByteString(hs, v)
		hs.buf.AppendByte('"')
	default:
		hs.buf.AppendBytes(v)
	}
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
				hs.buf.AppendString(`\t`)
			case '\r':
				hs.buf.AppendString(`\r`)
			case '\n':
				hs.buf.AppendString(`\n`)
			case '\\':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('\\')
			case '"':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('"')
			default:
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hex[c>>4])
				hs.buf.AppendByte(hex[c&0xf])
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
				hs.buf.AppendString(`\t`)
			case '\r':
				hs.buf.AppendString(`\r`)
			case '\n':
				hs.buf.AppendString(`\n`)
			case '\\':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('\\')
			case '"':
				hs.buf.AppendByte('\\')
				hs.buf.AppendByte('"')
			default:
				hs.buf.AppendString(`\u00`)
				hs.buf.AppendByte(hex[c>>4])
				hs.buf.AppendByte(hex[c&0xf])
			}
			i++
			p = i

		default:
			v, wd := utf8.DecodeRune(s[i:])
			if v == utf8.RuneError && wd == 1 {
				hs.buf.AppendBytes(s[p:i])
				hs.buf.AppendString(`\ufffd`)
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
	i := min(max(0, (level+4)/4), 3)
	hs.buf.AppendString(hs.h.theme.Level[i].Text)
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

// ---

type shared struct {
	options
	mu     sync.Mutex
	writer io.Writer
}

// ---

type groupKeys struct {
	head    [4]string
	tail    []string
	headLen int
}

func (g *groupKeys) len() int {
	return g.headLen + len(g.tail)
}

func (g *groupKeys) at(i int) string {
	if i < g.headLen {
		return g.head[i]
	}

	return g.tail[i-g.headLen]
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

const hex = "0123456789abcdef"

var _ slog.Handler = (*Handler)(nil)
