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

// TweakHandler returns a builder for a new handler based on existing handler.
func TweakHandler(handler slog.Handler) TweakHandlerBuilder {
	return TweakHandlerBuilder{handler, handlerTweaks{}}
}

// ---

// TweakHandlerBuilder is a builder for a new handler based on existing handler.
type TweakHandlerBuilder struct {
	handler slog.Handler
	tweaks  handlerTweaks
}

// WithDynamicAttr adds a dynamic attribute to the handler.
func (b TweakHandlerBuilder) WithDynamicAttr(attr func(context.Context) slog.Attr) TweakHandlerBuilder {
	b.tweaks.dynamicAttrs = append(b.tweaks.dynamicAttrs, attr)

	return b
}

// Result returns the new handler.
func (b TweakHandlerBuilder) Result() slog.Handler {
	return &tweakedHandler{b.handler, b.tweaks}
}

// ---

type handlerTweaks struct {
	dynamicAttrs []func(context.Context) slog.Attr
}

// ---

type tweakedHandler struct {
	base slog.Handler
	handlerTweaks
}

func (h *tweakedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *tweakedHandler) Handle(ctx context.Context, record slog.Record) error {
	if len(h.dynamicAttrs) != 0 {
		record = record.Clone()

		for _, attr := range h.dynamicAttrs {
			if attr := attr(ctx); !attr.Equal(slog.Attr{}) {
				record.AddAttrs(attr)
			}
		}
	}

	return h.base.Handle(ctx, record)
}

func (h *tweakedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &tweakedHandler{h.base.WithAttrs(attrs), h.handlerTweaks}
}

func (h *tweakedHandler) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	return &tweakedHandler{h.base.WithGroup(key), h.handlerTweaks}
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
