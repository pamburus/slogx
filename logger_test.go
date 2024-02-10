package slogx_test

import (
	"context"
	"log/slog"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/internal/mock"

	"testing"
)

func TestLogger(tt *testing.T) {
	t := New(tt)

	t.Run("Default", func(t Test) {
		cl := mock.NewCallLog()
		handler := mock.NewHandler(cl)
		slog.SetDefault(slog.New(handler))
		logger := slogx.Default().WithSource(false)

		logger.Info("msg")
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelInfo,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelInfo,
					Message: "msg",
				},
			},
		))
	})

	test := func(name string, fn func(Test, *mock.CallLog, *slogx.Logger)) {
		cl := mock.NewCallLog()
		handler := mock.NewHandler(cl)
		logger := slogx.New(handler).WithSource(false)

		t.Run(name, func(t Test) {
			fn(t, cl, logger)
		})
	}

	test("Debug", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Debug("msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelDebug,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelDebug,
					Message: "msg",
					Attrs:   []mock.Attr{{Key: "a", Value: "v"}},
				},
			},
		))
	})

	test("Info", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Info("msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelInfo,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelInfo,
					Message: "msg",
					Attrs:   []mock.Attr{{Key: "a", Value: "v"}},
				},
			},
		))
	})

	test("Warn", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Warn("msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelWarn,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelWarn,
					Message: "msg",
					Attrs:   []mock.Attr{{Key: "a", Value: "v"}},
				},
			},
		))
	})

	test("Error", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Error("msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelError,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelError,
					Message: "msg",
					Attrs:   []mock.Attr{{Key: "a", Value: "v"}},
				},
			},
		))
	})

	test("Log", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Log(slog.LevelWarn, "msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelWarn,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelWarn,
					Message: "msg",
					Attrs:   []mock.Attr{{Key: "a", Value: "v"}},
				},
			},
		))
	})

	test("With", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger = logger.With(
			slog.String("a", "va"),
			slog.String("b", "vb"),
		)
		logger.Log(slog.LevelInfo, "msg", slog.String("c", "d"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerWithAttrs{
				Instance: "0",
				Attrs: []mock.Attr{
					{Key: "a", Value: "va"},
					{Key: "b", Value: "vb"},
				},
			},
			mock.HandlerEnabled{
				Instance: "0.1",
				Level:    slog.LevelInfo,
			},
			mock.HandlerHandle{
				Instance: "0.1",
				Record: mock.Record{
					Level:   slog.LevelInfo,
					Message: "msg",
					Attrs: []mock.Attr{
						{Key: "c", Value: "d"},
					},
				},
			},
		))
	})

	test("WithGroup", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger = logger.WithGroup("g1")
		logger.Log(slog.LevelInfo, "msg", slog.String("a", "v"))
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerWithGroup{
				Instance: "0",
				Key:      "g1",
			},
			mock.HandlerEnabled{
				Instance: "0.1",
				Level:    slog.LevelInfo,
			},
			mock.HandlerHandle{
				Instance: "0.1",
				Record: mock.Record{
					Level:   slog.LevelInfo,
					Message: "msg",
					Attrs: []mock.Attr{
						{Key: "a", Value: "v"},
					},
				},
			},
		))
	})

	test("Enabled", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		logger.Enabled(context.Background(), slog.LevelDebug)
		t.Expect(cl.Calls().WithoutTime()...).To(Equal(
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelDebug,
			},
		))
	})

	test("SlogLogger", func(t Test, cl *mock.CallLog, logger *slogx.Logger) {
		t.Expect(logger.SlogLogger().Handler()).To(Equal(logger.Handler()))
	})
}

func BenchmarkLogger(b *testing.B) {
	handler := slogx.Discard()

	testEnabled := func(b *testing.B, logger *slogx.Logger) {
		b.Run("Enabled", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i != b.N; i++ {
				logger.Enabled(context.Background(), slog.LevelDebug)
			}
		})
	}

	testLogAttrs := func(b *testing.B, logger *slogx.Logger) {
		b.Run("Log", func(b *testing.B) {
			b.Run("NoAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.Log(slog.LevelDebug, "msg")
				}
			})
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.Log(slog.LevelDebug, "msg", slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
		})
	}

	testWith := func(b *testing.B, logger *slogx.Logger) {
		b.Run("With", func(b *testing.B) {
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
		})
	}

	testWithAndLog := func(b *testing.B, logger *slogx.Logger) {
		b.Run("WithAndLog", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"))
					logger.Log(slog.LevelDebug, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
		})
	}

	testAll := func(b *testing.B, logger *slogx.Logger) {
		testEnabled(b, logger)
		testLogAttrs(b, logger)
		testWith(b, logger)
		testWithAndLog(b, logger)
	}

	testWithSource := func(b *testing.B, logger *slogx.Logger, enabled bool) {
		name := "WithSource"
		if !enabled {
			name = "WithoutSource"
		}

		logger = logger.WithSource(enabled)

		b.Run(name, func(b *testing.B) {
			b.Run("Unwrapped", func(b *testing.B) {
				testAll(b, slogx.New(handler))
			})

			b.Run("3xWrapped", func(b *testing.B) {
				handler = wrapHandlerN(handler, 3)
				testAll(b, slogx.New(handler))
			})
		})
	}

	testWithSource(b, slogx.New(handler), false)
	testWithSource(b, slogx.New(handler), true)
}

// ---

func wrapHandlerN(handler slog.Handler, times int) slog.Handler {
	for i := 0; i != 3; i++ {
		handler = wrapHandler(handler)
	}

	return handler
}

func wrapHandler(handler slog.Handler) slog.Handler {
	return &testHandlerWrapper{handler}

}

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
