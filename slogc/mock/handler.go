package mock

import (
	"context"
	"fmt"
	"log/slog"
)

// ---

func NewHandler(log *CallLog) *Handler {
	return &Handler{
		instance: "0",
		log:      log,
	}
}

// ---

type Handler struct {
	instance string
	log      *CallLog
}

func (h Handler) Enabled(_ context.Context, level slog.Level) bool {
	h.log.append(HandlerEnabled{h.instance, level})

	return true
}

func (h Handler) Handle(_ context.Context, record slog.Record) error {
	h.log.append(HandlerHandle{h.instance, NewRecord(record)})

	return nil
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.instance = h.newInstance(
		h.log.append(HandlerWithAttrs{h.instance, NewAttrs(attrs)}),
	)

	return &h
}

func (h Handler) WithGroup(group string) slog.Handler {
	h.instance = h.newInstance(
		h.log.append(HandlerWithGroup{h.instance, group}),
	)

	return &h
}

func (h Handler) newInstance(n int) string {
	return fmt.Sprintf("%s.%d", h.instance, n)
}

// ---

type HandlerEnabled struct {
	Instance string
	Level    slog.Level
}

type HandlerHandle struct {
	Instance string
	Record   Record
}

type HandlerWithAttrs struct {
	Instance string
	Attrs    []Attr
}

type HandlerWithGroup struct {
	Instance string
	Key      string
}

// ---

var _ slog.Handler = Handler{}
