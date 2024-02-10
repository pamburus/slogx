package slogx_test

import (
	"context"
	"io"
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
			// mock.HandlerWithAttrs{
			// 	Instance: "0",
			// 	Attrs: []mock.Attr{
			// 		{Key: "a", Value: "va"},
			// 		{Key: "b", Value: "vb"},
			// 	},
			// },
			mock.HandlerEnabled{
				Instance: "0",
				Level:    slog.LevelInfo,
			},
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Level:   slog.LevelInfo,
					Message: "msg",
					Attrs: []mock.Attr{
						{Key: "a", Value: "va"},
						{Key: "b", Value: "vb"},
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
	b.Run("slogx", benchmarkSLogXLogger)
	b.Run("slog", benchmarkSLogLogger)
}

func benchmarkSLogXLogger(b *testing.B) {
	testEnabled := func(b *testing.B, logger *slogx.Logger) {
		b.Run("Enabled", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i != b.N; i++ {
				logger.Enabled(context.Background(), slog.LevelInfo)
			}
		})
	}

	testLogAttrs := func(b *testing.B, logger *slogx.Logger) {
		b.Run("LogAttrs", func(b *testing.B) {
			b.Run("NoAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg")
				}
			})
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
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
			b.Run("FiveAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
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
					logger.Log(slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
					logger.Log(slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
					logger.Log(slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testLogAfterWith := func(b *testing.B, logger *slogx.Logger) {
		b.Run("LogAfterWith", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.Log(slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.Log(slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.Log(slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testAllForLogger := func(b *testing.B, logger *slogx.Logger) {
		testEnabled(b, logger)
		testLogAttrs(b, logger)
		testWith(b, logger)
		testWithAndLog(b, logger)
		testLogAfterWith(b, logger)
	}

	testWithSource := func(b *testing.B, handler slog.Handler, enabled bool) {
		name := "WithSource"
		if !enabled {
			name = "WithoutSource"
		}

		b.Run(name, func(b *testing.B) {
			b.Run("Unwrapped", func(b *testing.B) {
				logger := slogx.New(handler).WithSource(enabled)
				testAllForLogger(b, logger)
			})

			b.Run("3xWrapped", func(b *testing.B) {
				handler = wrapHandlerN(handler, 3)
				logger := slogx.New(handler).WithSource(enabled)
				testAllForLogger(b, logger)
			})
		})
	}

	testAllForHandler := func(b *testing.B, handler slog.Handler) {
		testWithSource(b, handler, false)
		testWithSource(b, handler, true)
	}

	b.Run("Discard", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(b, slogx.Discard())
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(b, &enabledDiscardHandler{})
		})
	})

	b.Run("JSON", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		})
	})
}

func benchmarkSLogLogger(b *testing.B) {
	testEnabled := func(b *testing.B, logger *slog.Logger) {
		b.Run("Enabled", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i != b.N; i++ {
				logger.Enabled(context.Background(), slog.LevelInfo)
			}
		})
	}

	testLogAttrs := func(b *testing.B, logger *slog.Logger) {
		b.Run("LogAttrs", func(b *testing.B) {
			b.Run("NoAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg")
				}
			})
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
		})
	}

	testWith := func(b *testing.B, logger *slog.Logger) {
		b.Run("With", func(b *testing.B) {
			b.Run("ThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				}
			})
			b.Run("FiveAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
		})
	}

	testWithAndLog := func(b *testing.B, logger *slog.Logger) {
		b.Run("WithAndLog", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"))
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testLogAfterWith := func(b *testing.B, logger *slog.Logger) {
		b.Run("LogAfterWith", func(b *testing.B) {
			b.Run("TwoAndThreeAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				}
			})
			b.Run("ThreeAndFourAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("d", "dv"), slog.String("e", "ev"), slog.String("f", "fv"), slog.String("g", "gv"))
				}
			})
			b.Run("FiveAndThreeAttrs", func(b *testing.B) {
				logger := logger.With(slog.String("a", "av"), slog.String("b", "bv"), slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
				b.ResetTimer()
				for i := 0; i != b.N; i++ {
					logger.LogAttrs(context.Background(), slog.LevelInfo, "msg", slog.String("f", "fv"), slog.String("g", "gv"), slog.String("h", "hv"))
				}
			})
		})
	}

	testAllForLogger := func(b *testing.B, logger *slog.Logger) {
		testEnabled(b, logger)
		testLogAttrs(b, logger)
		testWith(b, logger)
		testWithAndLog(b, logger)
		testLogAfterWith(b, logger)
	}

	testAllForHandler := func(b *testing.B, handler slog.Handler) {
		b.Run("WithSource", func(b *testing.B) {
			b.Run("Unwrapped", func(b *testing.B) {
				logger := slog.New(handler)
				testAllForLogger(b, logger)
			})

			b.Run("3xWrapped", func(b *testing.B) {
				handler = wrapHandlerN(handler, 3)
				logger := slog.New(handler)
				testAllForLogger(b, logger)
			})
		})
	}

	b.Run("Discard", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(b, slogx.Discard())
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(b, &enabledDiscardHandler{})
		})
	})

	b.Run("JSON", func(b *testing.B) {
		b.Run("Disabled", func(b *testing.B) {
			testAllForHandler(b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
		})
		b.Run("Enabled", func(b *testing.B) {
			testAllForHandler(b, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		})
	})
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

func (h *enabledDiscardHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *enabledDiscardHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (h *enabledDiscardHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *enabledDiscardHandler) WithGroup(key string) slog.Handler {
	return h
}
