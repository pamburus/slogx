package mock

import (
	"log/slog"
)

func NewAttrsForRecord(record slog.Record) []Attr {
	if record.NumAttrs() == 0 {
		return nil
	}

	attrs := make([]Attr, 0, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, Attr{
			Key:   a.Key,
			Value: a.Value.Any(),
		})

		return true
	})

	return attrs
}

func NewAttrs(attrs []slog.Attr) []Attr {
	if len(attrs) == 0 {
		return nil
	}

	testAttrs := make([]Attr, 0, len(attrs))
	for _, a := range attrs {
		testAttrs = append(testAttrs, Attr{
			Key:   a.Key,
			Value: a.Value.Any(),
		})
	}

	return testAttrs
}

func NewAttr(key string, value any) Attr {
	return Attr{Key: key, Value: value}
}

type Attr struct {
	Key   string
	Value any
}
