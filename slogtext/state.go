package slogtext

import (
	"context"
	"log/slog"
	"sync"
)

// ---

func newHandleState(ctx context.Context, h *Handler) *handleState {
	s := handleStatePool.Get().(*handleState)
	s.ctx = ctx
	s.groups = h.groupKeys.collect(s.groups)

	return s
}

type handleState struct {
	ctx            context.Context
	buf            buffer
	scratch        buffer
	keyPrefix      buffer
	groupPrefixLen int
	groups         []string
	attrsToExpand  []attrToExpand
	messageBegin   int
	expandingAttrs bool
	pcs            [1]uintptr
	sourceAttr     slog.Attr
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
	s.expandingAttrs = false
	s.messageBegin = 0
	s.pcs[0] = 0
	s.sourceAttr = slog.Attr{}

	handleStatePool.Put(s)
}

func (s *handleState) addAttrToExpand(attr slog.Attr) {
	s.attrsToExpand = append(s.attrsToExpand, attrToExpand{
		Attr:      attr,
		KeyPrefix: s.keyPrefix.String(),
	})
}

type attrToExpand struct {
	slog.Attr
	KeyPrefix string
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
