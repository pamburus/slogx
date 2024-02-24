package model

import "sync"

func NewBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

// ---

type Buffer []byte

func (b *Buffer) Len() int {
	return len(*b)
}

func (b *Buffer) Cap() int {
	return cap(*b)
}

// ExtendToCap increases the slice's length to its capacity and returns the extended part.
func (b *Buffer) Tail() []byte {
	return (*b)[len(*b):cap(*b)]
}

func (b *Buffer) Free() {
	if cap(*b) <= bufSize {
		*b = (*b)[:0]
		bufferPool.Put(b)
	}
}

// ---

var bufferPool = sync.Pool{
	New: func() any {
		buf := Buffer(make([]byte, 0, bufSize))

		return &buf
	},
}

const bufSize = 1 << 20
