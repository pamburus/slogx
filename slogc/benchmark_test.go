package slogc_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/slogc"
)

func BenchmarkContextLogging(b *testing.B) {
	testEnabled := func(ctx context.Context, b *testing.B) {
		b.Run("Enabled", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i != b.N; i++ {
				slogc.Get(ctx).Enabled(ctx, slog.LevelInfo)
			}
		})
	}

	testLogAttrs := func(ctx context.Context, b *testing.B) {
		b.Run("Log", func(b *testing.B) {
			b.Run("NoAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.Log(ctx, slog.LevelInfo, "msg")
				}
			})
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
		})
	}

	testWith := func(ctx context.Context, b *testing.B) {
		b.Run("With", func(b *testing.B) {
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.With(ctx, slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
			b.Run("FiveAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.With(ctx, slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
		})
	}

	testWithAndLog := func(ctx context.Context, b *testing.B) {
		b.Run("LogWithAndLog", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := slogc.Get(ctx).With(slog.String("a", "av"), slog.String("b", "bv"))
					logger.Log(ctx, slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := slogc.Get(ctx).With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
					logger.Log(ctx, slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := slogc.Get(ctx).With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
					logger.Log(ctx, slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testLogAfterWith := func(ctx context.Context, b *testing.B) {
		b.Run("LogAfterWith", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				ctx := slogc.With(ctx, slog.String("a", "av"), slog.String("b", "bv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				ctx := slogc.With(ctx, slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				ctx := slogc.With(ctx, slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testAllForContext := func(ctx context.Context, b *testing.B) {
		testEnabled(ctx, b)
		testLogAttrs(ctx, b)
		testWith(ctx, b)
		testWithAndLog(ctx, b)
		testLogAfterWith(ctx, b)
	}

	testWithSource := func(ctx context.Context, b *testing.B, handler slog.Handler, enabled bool) {
		name := "WithSource"
		if !enabled {
			name = "WithoutSource"
		}

		b.Run(name, func(b *testing.B) {
			b.Run("Unwrapped", func(b *testing.B) {
				ctx := slogc.New(ctx, slogx.NewContextLogger(handler).WithSource(enabled))
				testAllForContext(ctx, b)
			})

			b.Run("3xWrapped", func(b *testing.B) {
				handler = wrapHandlerN(handler, 3)
				ctx := slogc.New(ctx, slogx.NewContextLogger(handler).WithSource(enabled))
				testAllForContext(ctx, b)
			})
		})
	}

	testAllForHandler := func(ctx context.Context, b *testing.B, handler slog.Handler) {
		testWithSource(ctx, b, handler, false)
		testWithSource(ctx, b, handler, true)
	}

	b.Run("Discard", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(context.Background(), b, slogx.Discard())
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(context.Background(), b, &enabledDiscardHandler{})
		})
	})

	b.Run("slog.JSONHandler", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(context.Background(), b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(context.Background(), b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		})
	})
}

// ---

func wrapHandlerN(handler slog.Handler, times int) slog.Handler {
	for i := 0; i != times; i++ {
		handler = wrapHandler(handler)
	}

	return handler
}

func wrapHandler(handler slog.Handler) slog.Handler {
	return &testHandlerWrapper{handler}
}

// ---

type testHandlerWrapper struct {
	base slog.Handler
}

func (h *testHandlerWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *testHandlerWrapper) Handle(ctx context.Context, record slog.Record) error {
	return h.base.Handle(ctx, record)
}

func (h *testHandlerWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &testHandlerWrapper{h.base.WithAttrs(attrs)}
}

func (h *testHandlerWrapper) WithGroup(key string) slog.Handler {
	if key == "" {
		return h
	}

	return &testHandlerWrapper{h.base.WithGroup(key)}
}

//---

type enabledDiscardHandler struct{}

func (h *enabledDiscardHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *enabledDiscardHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h *enabledDiscardHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *enabledDiscardHandler) WithGroup(string) slog.Handler {
	return h
}
