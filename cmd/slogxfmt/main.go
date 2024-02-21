package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func main() {
	input := bufio.NewReader(os.Stdin)
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		input = bufio.NewReader(file)
	}

	handler := slogtext.NewHandler(os.Stdout,
		slogtext.WithLevel(slog.LevelDebug),
		slogtext.WithColor(slogtext.ColorAlways),
		slogtext.WithSource(true),
		slogtext.WithLoggerNameKey("logger"),
		slogtext.WithTheme(themes.Fancy()),
	)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		var line line
		err := json.Unmarshal(scanner.Bytes(), &line)
		if err != nil {
			continue
		}

		ts := get[time.Time](line, "ts")
		levelIn := get[string](line, "level")
		message := get[string](line, "msg")

		delete(line, "ts")
		delete(line, "level")
		delete(line, "msg")
		delete(line, "message")

		var level slog.Level
		switch levelIn {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}

		record := slog.NewRecord(ts, level, message, 0)

		if source, ok := line["caller"]; ok {
			record.AddAttrs(slog.String(slog.SourceKey, parse[string](source)))
			delete(line, "caller")
		}

		for k, v := range line {
			val := parse[value](v).val
			if (k == "error" || k == "error-verbose") && val.Kind() == slog.KindString {
				record.AddAttrs(slog.Any(k, errors.New(val.String())))
			} else {
				record.AddAttrs(slog.Attr{Key: k, Value: val})
			}
		}

		err = handler.Handle(context.Background(), record)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type line map[string]json.RawMessage

func get[T any](m map[string]json.RawMessage, key string) T {
	return parse[T](m[key])
}

func parse[T any](m json.RawMessage) T {
	var v T
	_ = json.Unmarshal(m, &v)

	return v
}

type value struct {
	val slog.Value
}

func (v *value) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	switch data[0] {
	case '{':
		var m map[string]json.RawMessage
		err := json.Unmarshal(data, &m)
		if err != nil {
			return err
		}

		var items []slog.Attr
		for k, v := range m {
			items = append(items, slog.Attr{
				Key:   k,
				Value: parse[value](v).val,
			})
		}

		v.val = slog.GroupValue(items...)

	case '[':
		var items []json.RawMessage
		err := json.Unmarshal(data, &items)
		if err != nil {
			return err
		}

		var values []slog.Value
		for _, item := range items {
			var val value
			err := json.Unmarshal(item, &val)
			if err != nil {
				return err
			}

			values = append(values, val.val)
		}

		v.val = slog.AnyValue(values)

	case '"':
		var s string
		err := json.Unmarshal(data, &s)
		if err != nil {
			return err
		}

		v.val = slog.StringValue(s)

	case 't', 'f':
		var b bool
		err := json.Unmarshal(data, &b)
		if err != nil {
			return err
		}

		v.val = slog.BoolValue(b)

	case 'n':
		v.val = slog.AnyValue(nil)

	default:
		var n json.Number
		err := json.Unmarshal(data, &n)
		if err != nil {
			return err
		}

		if nv, err := strconv.ParseUint(n.String(), 10, 64); err == nil {
			v.val = slog.Uint64Value(nv)
		} else if nv, err := n.Int64(); err == nil {
			v.val = slog.Int64Value(nv)
		} else if nv, err := n.Float64(); err == nil {
			v.val = slog.Float64Value(nv)
		} else {
			v.val = slog.StringValue(n.String())
		}
	}

	return nil
}
