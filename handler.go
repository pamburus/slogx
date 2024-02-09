package slogx

import (
	"context"
	"errors"
	"log/slog"
)

// Join returns a new handler that joins the provided handlers.
func Join(handlers ...slog.Handler) slog.Handler {
	switch len(handlers) {
	case 0:
		return Discard()
	case 1:
		return handlers[0]
	}

	return &multiHandler{handlers}
}

// Discard returns a handler that discards all log records.
func Discard() slog.Handler {
	return discardHandlerInstance
}

// ---

var discardHandlerInstance = &discardHandler{}

// ---

type discardHandler struct{}

func (h *discardHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h *discardHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h *discardHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *discardHandler) WithGroup(string) slog.Handler {
	return h
}

// ---

type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error

	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			err := handler.Handle(ctx, record)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}

	return &multiHandler{handlers}
}

func (h *multiHandler) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(key)
	}

	return &multiHandler{handlers}
}
