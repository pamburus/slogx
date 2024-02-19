package slogtext_test

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func BenchmarkHandler(b *testing.B) {
	ctx := context.Background()

	type testFunc func(handler slog.Handler) func(*testing.B)

	withAttrs := func(fn testFunc) testFunc {
		return func(handler slog.Handler) func(*testing.B) {
			return fn(handler.WithAttrs(commonAttrs))
		}
	}

	withGroup := func(key string, fn testFunc) testFunc {
		return func(handler slog.Handler) func(*testing.B) {
			return fn(handler.WithGroup(key))
		}
	}

	testHandle := func(handler slog.Handler) func(*testing.B) {
		return func(b *testing.B) {
			record := testRecordTemplate

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				err := handler.Handle(ctx, record)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}

	testWithAttrs := func(handler slog.Handler) func(*testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				handler.WithAttrs(perCallAttrs)
			}
		}
	}

	testWithAttrsAndHandle := func(handler slog.Handler) func(b *testing.B) {
		return func(b *testing.B) {
			record := slog.Record{
				Level:   slog.LevelInfo,
				Time:    testRecordTemplate.Time,
				Message: testRecordTemplate.Message,
			}
			handler := handler.WithAttrs(commonAttrs)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				err := handler.WithAttrs(perCallAttrs).Handle(ctx, record)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}

	testWithGroup := func(handler slog.Handler) func(*testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				handler.WithGroup("group")
			}
		}
	}

	testWithGroupAndHandle := func(handler slog.Handler) func(*testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			record := testRecordTemplate

			for i := 0; i != b.N; i++ {
				err := handler.WithGroup("group").Handle(ctx, record)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}

	testSuite := func(do func(testFunc) func(*testing.B)) func(*testing.B) {
		return func(b *testing.B) {
			b.Run("HandleAfterWithAttrs", do(withAttrs(testHandle)))
			b.Run("WithAttrsFirst", do(testWithAttrs))
			b.Run("WithAttrsSecond", do(withAttrs(testWithAttrs)))
			b.Run("WithAttrsAndHandle", do(testWithAttrsAndHandle))
			b.Run("WithGroupFirst", do(testWithGroup))
			b.Run("WithGroupSecond", do(withGroup("first-group", testWithGroup)))
			b.Run("WithGroupSecondAndHandle", do(withGroup("first-group", testWithGroupAndHandle)))
		}
	}

	b.Run("slogtext/Handler", func(b *testing.B) {
		test := func(options ...slogtext.Option) func(test testFunc) func(b *testing.B) {
			return func(test testFunc) func(b *testing.B) {
				newHandler := func(color slogtext.ColorSetting) slog.Handler {
					return slogtext.NewHandler(io.Discard, append(slices.Clip(options),
						slogtext.WithColor(color),
						slogtext.WithTheme(themes.Tint()),
					)...)
				}

				return func(b *testing.B) {
					b.Run("WithoutColor", test(newHandler(slogtext.ColorNever)))
					b.Run("WithColor", test(newHandler(slogtext.ColorAlways)))
				}
			}
		}

		b.Run("WithoutSource", testSuite(test(slogtext.WithSource(false))))
		b.Run("WithSource", testSuite(test(slogtext.WithSource(true))))
	})

	b.Run("slog", func(b *testing.B) {
		type handlerNewFunc func(addSource bool) slog.Handler

		newJSONHandler := func(addSource bool) slog.Handler {
			return slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
				AddSource: addSource,
			}).WithAttrs(commonAttrs)
		}

		newTextHandler := func(addSource bool) slog.Handler {
			return slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
				AddSource: addSource,
			}).WithAttrs(commonAttrs)
		}

		test := func(fn handlerNewFunc, addSource bool) func(test testFunc) func(b *testing.B) {
			return func(test testFunc) func(b *testing.B) {
				return test(fn(addSource))
			}
		}

		withSourceVariants := func(fn handlerNewFunc) func(b *testing.B) {
			return func(b *testing.B) {
				b.Run("WithoutSource", testSuite(test(fn, false)))
				b.Run("WithSource", testSuite(test(fn, true)))
			}
		}

		b.Run("JSONHandler", withSourceVariants(newJSONHandler))
		b.Run("TextHandler", withSourceVariants(newTextHandler))
	})
}

// ---

var testRecordTemplate = newRecordTemplate()

func newRecordTemplate() slog.Record {
	record := slog.Record{
		Level:   slog.LevelInfo,
		Time:    time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
		Message: "The quick brown fox jumps over a lazy dog",
	}

	record.AddAttrs(perCallAttrs...)

	return record
}

var commonAttrs = []slog.Attr{
	slog.String("derived-string-field-1", "string-value-1"),
	slog.Int("derived-int-field", 420),
	slog.Group("group-1", slog.String("attr-1.1", "value-1"), slog.Int("attr-1.1", 42)),
	slog.String("derived-string-field-2", "string-value-2"),
	slog.Int("derived-int-field", 840),
}

var perCallAttrs = []slog.Attr{
	slog.String("string-field", "string-value"),
	slog.Int("int-field", 42),
	slog.Group("group-2", slog.String("attr-2.1", "value-1"), slog.Int("attr-2.2", 44)),
}
