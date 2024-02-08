package slogc

import (
	"context"
)

// WithName returns a new context with the logger name composed
// of the base name that is already in context as a prefix and the provided name
// using dot as a separator.
// If the name is empty, the context is returned as is.
func WithName(ctx context.Context, name string) context.Context {
	if name == "" {
		return ctx
	}

	base := Name(ctx)
	if base != "" {
		name = base + "." + name
	}

	return context.WithValue(ctx, &contextKeyName, name)
}

// Name returns the full logger name from the context.
// If the name is not set, an empty string is returned.
func Name(ctx context.Context) string {
	if name, ok := ctx.Value(&contextKeyName).(string); ok {
		return name
	}

	return ""
}

// ---

var (
	contextKeyName int
)
