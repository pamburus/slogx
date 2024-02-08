package mock

import (
	"log/slog"
	"time"
)

func NewRecord(record slog.Record) Record {
	return Record{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level,
		PC:      record.PC,
		Attrs:   NewAttrsForRecord(record),
	}
}

type Record struct {
	Time    time.Time
	Message string
	Level   slog.Level
	PC      uintptr
	Attrs   []Attr
}
