package slogx

import (
	"context"
	"errors"
	"log/slog"
)

// HandlerWithNameAsAttr returns a new handler that adds the logger name
// to the log record as an attribute with the provided key.
// If the key is empty, the handler is returned as is.
// If WithGroup is used, then the logger name attribute is added to the group.
// func HandlerWithNameAsAttr(handler slog.Handler, attrKey string) slog.Handler {
// 	if attrKey == "" {
// 		return handler
// 	}

// 	return &nameHandler{handler, attrKey, ""}
// }

// JoinHandlers returns a new handler that joins the provided handlers.
func JoinHandlers(handlers ...slog.Handler) slog.Handler {
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
	return &discardHandler{}
}

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

// ---

/*
type nameHandler struct {
	base      slog.Handler
	attrKey   string
	attrValue string
}

func (h *nameHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *nameHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.attrValue != "" {
		record = record.Clone()
		record.Add(slog.String(h.attrKey, h.attrValue))
	}

	return h.base.Handle(ctx, record)
}

func (h *nameHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &nameHandler{
		h.base.WithAttrs(attrs),
		h.attrKey,
		h.attrValue,
	}
}

func (h *nameHandler) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	return &nameHandler{
		h.base.WithGroup(key),
		h.attrKey,
		h.attrValue,
	}
}

func (h *nameHandler) WithName(name string) slog.Handler {
	return &nameHandler{
		h.base,
		h.attrKey,
		joinName(h.attrValue, name),
	}
}
*/

// ---

func joinName(base, name string) string {
	if name == "" {
		return base
	}

	if base != "" {
		name = base + "." + name
	}

	return name
}

// ---

// var (
// 	_ slog.Handler    = (*nameHandler)(nil)
// 	_ HandlerWithName = (*nameHandler)(nil)
// )
