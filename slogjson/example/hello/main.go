// Package main is an example of how to use the slogjson package.
package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/slogjson"
)

func main() {
	// Create handler.
	var handler slog.Handler = slogjson.NewHandler(os.Stdout,
		slogjson.WithLevel(slog.LevelDebug),
		slogjson.WithSource(true),
		slogjson.WithBytesFormat(slogjson.BytesFormatString),
		slogjson.WithLevelOffset(false),
		slogjson.WithTimeFormat(time.RFC3339Nano),
		slogjson.WithSourceEncodeFunc(slogjson.SourceShortObject),
	)

	// Create logger and set it as default.
	slog.SetDefault(slog.New(handler))

	// Log some messages.
	slog.Info("hello",
		slog.Group("types",
			slog.Any("level", slog.LevelWarn+1),
			slog.String("string", "value"),
			slog.Group("numbers",
				slog.Int("int", 42),
				slog.Float64("float", 3.14),
			),
			slog.Bool("bool", true),
			slog.Group("empty", slog.Attr{}),
			slog.Any("nil", nil),
		),
		slog.Group("arrays",
			slog.Any("empty", []any{}),
			slog.Any("nil", []any(nil)),
			slog.Any("ints", []any{1, 2, 3}),
			slog.Any("values", []slog.Value{
				slog.IntValue(1),
				slog.StringValue("two"),
				slog.AnyValue(nil),
				slog.BoolValue(true),
			}),
			slog.Any("bytes", []byte("test-me")),
		),
		slog.Group("maps",
			slog.Any("empty", map[string]any{}),
			slog.Any("nil", map[string]any(nil)),
			slog.Any("values", map[string]any{"one": 1, "two": "two", "nil": nil, "true": true}),
			slog.Any("int-keys", map[int]any{1: "one", 2: "two"}),
			slog.Any("any-keys", map[any]any{1: "one", 2: "two"}),
		),
		slogx.ErrorAttr(errors.New("short error message")),
	)
}
