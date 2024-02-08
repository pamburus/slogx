package slogc

import (
	"context"
	"log/slog"
)

// HandlerWithNameAsAttr returns a new handler that adds the logger name
// to the log record as an attribute with the provided key.
// If the key is empty, the handler is returned as is.
// If WithGroup is used, then the logger name attribute is added to the group.
func HandlerWithNameAsAttr(handler slog.Handler, attrKey string) slog.Handler {
	if attrKey == "" {
		return handler
	}

	return &nameHandler{handler, attrKey}
}

// ---

type nameHandler struct {
	base    slog.Handler
	attrKey string
}

func (h *nameHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *nameHandler) Handle(ctx context.Context, record slog.Record) error {
	name := Name(ctx)
	if name != "" {
		record = record.Clone()
		record.Add(slog.String(h.attrKey, name))
	}

	return h.base.Handle(ctx, record)
}

func (h *nameHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &nameHandler{h.base.WithAttrs(attrs), h.attrKey}
}

func (h *nameHandler) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	return &nameHandler{h.base.WithGroup(key), h.attrKey}
}
