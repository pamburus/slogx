package encbuf

import (
	"slices"
	"strconv"
	"sync"
)

// DefaultBufferSize is the default size of the buffer.
const DefaultBufferSize = 1024

// New creates a new instance of Buffer or gets existing one from a pool.
func New() *Buffer {
	return bufPool.Get().(*Buffer)
}

// Buffer is a helping wrapper for byte slice.
type Buffer []byte

// Write implements io.Writer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)

	return len(p), nil
}

// String implements fmt.Stringer.
func (b Buffer) String() string {
	return string(b)
}

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements.
func (b *Buffer) Grow(n int) {
	*b = slices.Grow(*b, n)
}

// Extend increases the slice's length by n and returns the extended part.
func (b *Buffer) Extend(n int) []byte {
	*b = append(*b, make([]byte, n)...)

	return (*b)[len(*b)-n:]
}

// AppendString appends a string to the Buffer.
func (b *Buffer) AppendString(data string) {
	*b = append(*b, data...)
}

// AppendBytes appends a byte slice to the Buffer.
func (b *Buffer) AppendBytes(data []byte) {
	*b = append(*b, data...)
}

// AppendByte appends a single byte to the Buffer.
func (b *Buffer) AppendByte(data byte) {
	*b = append(*b, data)
}

// Reset resets the underlying byte slice.
func (b *Buffer) Reset() {
	*b = (*b)[:0]
}

// Back returns the last byte of the underlying byte slice. A caller is in
// charge of checking that the Buffer is not empty.
func (b Buffer) Back() byte {
	return b[len(b)-1]
}

// SetBack replaces last byte.
func (b *Buffer) SetBack(v byte) {
	(*b)[len(*b)-1] = v
}

// TrimBack removes n last bytes from the underlying byte slice.
func (b *Buffer) TrimBack(n int) {
	*b = (*b)[:len(*b)-n]
}

// TrimBackByte removes the last byte from the underlying byte slice if it
// matches the given byte.
func (b *Buffer) TrimBackByte(v byte) {
	if len(*b) > 0 && b.Back() == v {
		b.TrimBack(1)
	}
}

// Bytes returns the underlying byte slice as is.
func (b Buffer) Bytes() []byte {
	return b
}

// Len returns the length of the underlying byte slice.
func (b Buffer) Len() int {
	return len(b)
}

// Cap returns the capacity of the underlying byte slice.
func (b Buffer) Cap() int {
	return cap(b)
}

// SetLen sets the length of the underlying byte slice.
func (b *Buffer) SetLen(n int) {
	*b = (*b)[:n]
}

// AppendUint appends the string form in the base 10 of the given unsigned
// integer to the given Buffer.
func (b *Buffer) AppendUint(n uint64) {
	*b = strconv.AppendUint(*b, n, 10)
}

// AppendInt appends the string form in the base 10 of the given integer
// to the given Buffer.
func (b *Buffer) AppendInt(n int64) {
	*b = strconv.AppendInt(*b, n, 10)
}

// AppendFloat32 appends the string form of the given float32 to the given
// Buffer.
func (b *Buffer) AppendFloat32(n float32) {
	*b = strconv.AppendFloat(*b, float64(n), 'g', -1, 32)
}

// AppendFloat64 appends the string form of the given float32 to the given
// Buffer.
func (b *Buffer) AppendFloat64(n float64) {
	*b = strconv.AppendFloat(*b, n, 'g', -1, 64)
}

// AppendBool appends "true" or "false", according to the given bool to the
// given Buffer.
func (b *Buffer) AppendBool(n bool) {
	*b = strconv.AppendBool(*b, n)
}

// Free returns the Buffer to the pool.
func (b *Buffer) Free() {
	const maxBufferSize = 64 << 10
	if b.Cap() <= maxBufferSize {
		b.Reset()
		bufPool.Put(b)
	}
}

// ---

var bufPool = sync.Pool{
	New: func() any {
		buf := Buffer(make([]byte, 0, DefaultBufferSize))

		return &buf
	},
}
