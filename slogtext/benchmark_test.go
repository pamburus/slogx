package slogtext_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/pamburus/slogx/slogtext"
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

	testSuite := func(b *testing.B, do func(testFunc) func(*testing.B)) {
		b.Run("HandleAfterWithAttrs", do(withAttrs(testHandle)))
		b.Run("WithAttrsFirst", do(testWithAttrs))
		b.Run("WithAttrsSecond", do(withAttrs(testWithAttrs)))
		b.Run("WithAttrsAndHandle", do(testWithAttrsAndHandle))
		b.Run("WithGroupFirst", do(testWithGroup))
		b.Run("WithGroupSecond", do(withGroup("first-group", testWithGroup)))
		b.Run("WithGroupSecondAndHandle", do(withGroup("first-group", testWithGroupAndHandle)))
	}

	b.Run("slogtext.Handler", func(b *testing.B) {
		options := []slogtext.Option{
			slogtext.WithSource(false),
		}

		withColorVariants := func(test testFunc) func(b *testing.B) {
			return func(b *testing.B) {
				b.Run("WithColor", test(
					slogtext.NewHandler(io.Discard, append(options,
						slogtext.WithColor(slogtext.ColorAlways),
					)...),
				))
				b.Run("WithoutColor", test(
					slogtext.NewHandler(io.Discard, append(options,
						slogtext.WithColor(slogtext.ColorNever),
					)...),
				))
			}
		}

		testSuite(b, withColorVariants)
	})

	b.Run("slog.TextHandler", func(b *testing.B) {
		withFormatVariants := func(test testFunc) func(b *testing.B) {
			return func(b *testing.B) {
				b.Run("JSON", test(
					slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}).WithAttrs(commonAttrs),
				))
				b.Run("Text", test(
					slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}).WithAttrs(commonAttrs),
				))
			}
		}

		testSuite(b, withFormatVariants)
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
