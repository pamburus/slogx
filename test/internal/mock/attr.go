package mock

import (
	"log/slog"
)

// NewAttrs returns a new slice of [Attr] based on the given slice of [slog.Attr].
func NewAttrs(attrs []slog.Attr) []Attr {
	return AttrsUsingFunc(len(attrs), func(fn func(slog.Attr)) {
		for _, a := range attrs {
			fn(a)
		}
	})
}

// AttrsUsingFunc returns a new slice of [Attr] based on the given number and function to get next [slog.Attr].
func AttrsUsingFunc(n int, fn func(func(slog.Attr))) []Attr {
	if n == 0 {
		return nil
	}

	testAttrs := make([]Attr, 0, n)
	fn(func(a slog.Attr) {
		testAttrs = append(testAttrs, Attr{
			Key:   a.Key,
			Value: a.Value.Any(),
		})
	})

	return testAttrs
}

// NewAttr returns a new [Attr] based on the given key and value.
func NewAttr(key string, value any) Attr {
	return Attr{Key: key, Value: value}
}

// Attr is a mockable representation of [slog.Attr].
type Attr struct {
	Key   string
	Value any
}
