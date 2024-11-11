package slogx

import (
	"log/slog"
	"slices"
)

// ErrorAttr returns an attribute with the error.
func ErrorAttr(err error) slog.Attr {
	return slog.Any(ErrorKey, err)
}

const (
	// ErrorKey is the key used for the error attribute.
	ErrorKey = "error"
)

// AttrPack is a an optimized pack of attributes.
type AttrPack struct {
	front  [4]slog.Attr
	nFront int
	back   []slog.Attr
}

// Clone returns a copy of the AttrPack without a shared state.
func (r AttrPack) Clone() AttrPack {
	r.back = slices.Clip(r.back)

	return r
}

// Len returns the number of attributes in the AttrPack.
func (r AttrPack) Len() int {
	return r.nFront + len(r.back)
}

// Enumerate calls f on each Attr in the AttrPack.
func (r AttrPack) Enumerate(f func(slog.Attr) bool) {
	for i := range r.nFront {
		if !f(r.front[i]) {
			return
		}
	}

	for _, a := range r.back {
		if !f(a) {
			return
		}
	}
}

// Add appends the given Attrs to the AttrPack's list of Attrs.
func (r *AttrPack) Add(attrs ...slog.Attr) {
	var i int
	for i = 0; i < len(attrs) && r.nFront < len(r.front); i++ {
		a := attrs[i]
		if isEmptyGroup(a.Value) {
			continue
		}

		r.front[r.nFront] = a
		r.nFront++
	}

	if cap(r.back) > len(r.back) {
		end := r.back[:len(r.back)+1][len(r.back)]
		if !end.Equal(slog.Attr{}) {
			panic("multiple copies of a slogx.AttrPack modified the shared state simultaneously, use Clone to avoid this")
		}
	}

	ne := countEmptyGroups(attrs[i:])
	r.back = slices.Grow(r.back, len(attrs[i:])-ne)

	for _, a := range attrs[i:] {
		if !isEmptyGroup(a.Value) {
			r.back = append(r.back, a)
		}
	}
}

// Collect returns all the attributes in the AttrPack as a slice.
func (r *AttrPack) Collect() []slog.Attr {
	attrs := make([]slog.Attr, 0, r.Len())
	r.Enumerate(func(a slog.Attr) bool {
		attrs = append(attrs, a)

		return true
	})

	return attrs
}

// ---

func countEmptyGroups(as []slog.Attr) int {
	n := 0

	for _, a := range as {
		if isEmptyGroup(a.Value) {
			n++
		}
	}

	return n
}

func isEmptyGroup(v slog.Value) bool {
	if v.Kind() != slog.KindGroup {
		return false
	}

	return len(v.Group()) == 0
}
