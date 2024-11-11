// Package mock provides a mock slog.Handler implementation and other helpers.
package mock

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
)

// ---

// NewHandler returns a new [Handler] with the given [CallLog].
func NewHandler(log *CallLog) *Handler {
	return &Handler{
		instance: "0",
		log:      log,
	}
}

// ---

// Handler is a mockable representation of [slog.Handler].
type Handler struct {
	instance string
	log      *CallLog
}

// Enabled records the call and returns true.
func (h Handler) Enabled(_ context.Context, level slog.Level) bool {
	h.log.append(HandlerEnabled{h.instance, level})

	return true
}

// Handle records the call.
func (h Handler) Handle(_ context.Context, record slog.Record) error {
	h.log.append(HandlerHandle{h.instance, NewRecord(record)})

	return nil
}

// WithAttrs records the call and returns a new [Handler] with the given attributes.
func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.instance = h.newInstance(
		h.log.append(HandlerWithAttrs{h.instance, NewAttrs(attrs)}),
	)

	return &h
}

// WithGroup records the call and returns a new [Handler] with the given group.
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

// HandlerEnabled is a representation of a call to [slog.Handler.Enabled]
// that can be used to compare against recorded calls in a [CallLog].
type HandlerEnabled struct {
	Instance string
	Level    slog.Level
}

// HandlerHandle is a representation of a call to [slog.Handler.Handle]
// that can be used to compare against recorded calls in a [CallLog].
type HandlerHandle struct {
	Instance string
	Record   Record
}

func (h HandlerHandle) cloneAny() any {
	h.Record = h.Record.clone()

	return h
}

func (h HandlerHandle) withoutTime() any {
	h.Record = h.Record.WithoutTime()

	return h
}

// HandlerWithAttrs is a representation of a call to [slog.Handler.WithAttrs]
// that can be used to compare against recorded calls in a [CallLog].
type HandlerWithAttrs struct {
	Instance string
	Attrs    []Attr
}

func (h HandlerWithAttrs) cloneAny() any {
	return h.clone()
}

func (h HandlerWithAttrs) clone() HandlerWithAttrs {
	h.Attrs = slices.Clone(h.Attrs)

	return h
}

// HandlerWithGroup is a representation of a call to [slog.Handler.WithGroup]
// that can be used to compare against recorded calls in a [CallLog].
type HandlerWithGroup struct {
	Instance string
	Key      string
}

// ---

var (
	_ slog.Handler = Handler{}
	_ anyCloner    = (*HandlerHandle)(nil)
	_ timeRemover  = (*HandlerHandle)(nil)
	_ anyCloner    = (*HandlerWithAttrs)(nil)
)
