package slogjson

import (
	"context"
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
	groupPrefixLen int
	groups         []string
}

func (s *handleState) release() {
	if s.buf.Len()+s.scratch.Len() > 64<<10 {
		return
	}
	if len(s.groups) > 32 {
		return
	}

	s.ctx = nil
	s.buf.Reset()
	s.scratch.Reset()
	s.groups = s.groups[:0]
	s.groupPrefixLen = 0

	handleStatePool.Put(s)
}

// ---

var handleStatePool = sync.Pool{
	New: func() any {
		const bufSize = 3072
		const scratchSize = 1024
		const arenaSize = bufSize + scratchSize
		arena := buffer(make([]byte, 0, arenaSize))

		s := &handleState{
			buf:     arena[0:0:bufSize],
			scratch: arena[bufSize:bufSize],
			groups:  make([]string, 0, 8),
		}

		return s
	},
}
