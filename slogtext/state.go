package slogtext

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"
)

// ---

func newHandleState(ctx context.Context, h *Handler) *handleState {
	s := handleStatePool.Get().(*handleState)
	s.ctx = ctx
	s.groups = h.groupKeys.collect(s.groups)

	return s
}

type handleState struct {
	ctx                context.Context
	buf                buffer
	scratch            buffer
	keyPrefix          buffer
	groupPrefixLen     int
	groups             []string
	attrsToExpand      []attrToExpand
	levelBegin         int
	messageBegin       int
	timestampWidth     int
	numFlatAttrs       int
	expandingKeysWidth int
	expandingAttrs     bool
	initializingCache  bool
	pcs                [1]uintptr
	sourceAttr         slog.Attr
}

func (s *handleState) release() {
	if s.buf.Len()+s.scratch.Len()+s.keyPrefix.Len() > 64<<10 {
		return
	}
	if len(s.groups) > 32 {
		return
	}

	s.ctx = nil
	s.buf.Reset()
	s.scratch.Reset()
	s.groups = s.groups[:0]
	s.keyPrefix.Reset()
	s.groupPrefixLen = 0
	s.attrsToExpand = s.attrsToExpand[:0]
	s.expandingKeysWidth = 0
	s.expandingAttrs = false
	s.initializingCache = false
	s.levelBegin = 0
	s.messageBegin = 0
	s.timestampWidth = 0
	s.numFlatAttrs = 0
	s.pcs[0] = 0
	s.sourceAttr = slog.Attr{}

	handleStatePool.Put(s)
}

func (s *handleState) addAttrToExpand(attr slog.Attr) {
	hasNewLines := false
	switch attr.Value.Kind() {
	case slog.KindString:
		hasNewLines = strings.IndexByte(attr.Value.String(), '\n') >= 0
	case slog.KindAny:
		switch v := attr.Value.Any().(type) {
		case error:
			hasNewLines = strings.IndexByte(v.Error(), '\n') >= 0
		}
	}

	w := utf8.RuneCountInString(attr.Key) + utf8.RuneCount(s.keyPrefix)
	s.attrsToExpand = append(s.attrsToExpand, attrToExpand{
		Attr:        attr,
		KeyPrefix:   s.keyPrefix.String(),
		HasNewLines: hasNewLines,
	})
	if w > s.expandingKeysWidth {
		s.expandingKeysWidth = w
	}

}

type attrToExpand struct {
	slog.Attr
	KeyPrefix   string
	HasNewLines bool
}

// ---

var handleStatePool = sync.Pool{
	New: func() any {
		const bufSize = 3072
		const scratchSize = 1024
		const arenaSize = bufSize + scratchSize
		arena := buffer(make([]byte, 0, arenaSize))

		s := &handleState{
			buf:           arena[0:0:bufSize],
			scratch:       arena[bufSize:bufSize],
			groups:        make([]string, 0, 8),
			attrsToExpand: make([]attrToExpand, 0, 8),
		}
		s.keyPrefix.Grow(128)

		return s
	},
}
