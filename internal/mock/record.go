package mock

import (
	"log/slog"
	"slices"
	"time"
)

// NewRecord returns a new [Record] based on the given [slog.Record].
func NewRecord(record slog.Record) Record {
	return Record{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level,
		PC:      record.PC,
		Attrs:   recordAttrs(record),
	}
}

// Record is a mockable representation of [slog.Record].
type Record struct {
	Time    time.Time
	Message string
	Level   slog.Level
	PC      uintptr
	Attrs   []Attr
}

// WithoutTime returns a copy of [Record] with [Record.Time] set to zero.
func (r Record) WithoutTime() Record {
	r.Time = time.Time{}

	return r
}

func (r Record) cloneAny() any {
	return r.clone()
}

func (r Record) clone() Record {
	r.Attrs = slices.Clone(r.Attrs)

	return r
}

// ---

func recordAttrs(record slog.Record) []Attr {
	return AttrsUsingFunc(record.NumAttrs(), func(fn func(slog.Attr)) {
		record.Attrs(func(a slog.Attr) bool {
			fn(a)

			return true
		})
	})
}

// ---

var _ anyCloner = (*Record)(nil)
