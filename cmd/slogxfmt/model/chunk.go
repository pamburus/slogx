package model

import (
	"log/slog"
	"sync"
)

func NewChunk() *Chunk {
	return chunkPool.Get().(*Chunk)
}

// ---

type Chunk struct {
	Records []slog.Record
}

func (p *Chunk) Free() {
	if len(p.Records) <= 8192 {
		p.Records = p.Records[:0]
		chunkPool.Put(p)
	}
}

// ---

var chunkPool = sync.Pool{
	New: func() any {
		return &Chunk{
			Records: make([]slog.Record, 0, 2048),
		}
	},
}
