package slogc_test

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx/slogc"
)

func TestHandlerWithNameAsAttr(tt *testing.T) {
	t := New(tt)

	someTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("EmptyAttrKey", func(t Test) {
		handler := &testHandler{}
		t.Expect(slogc.HandlerWithNameAsAttr(handler, "")).To(Equal(handler))
	})

	t.Run("Enabled", func(t Test) {
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		t.Expect(handler.Enabled(context.Background(), slog.LevelInfo)).To(BeTrue())
		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallEnabled{"0", slog.LevelInfo},
		}))
	})

	t.Run("WithEmptyAttrs", func(t Test) {
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		t.Expect(handler.WithAttrs(nil)).To(Equal(handler))
	})

	t.Run("WithEmptyGroup", func(t Test) {
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		t.Expect(handler.WithGroup("")).To(Equal(handler))
	})

	t.Run("WithGroup", func(t Test) {
		ctx := context.Background()
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}

		handler := slogc.HandlerWithNameAsAttr(base, "@logger")
		handler = handler.WithAttrs([]slog.Attr{slog.String("a1", "av1")})
		handler = handler.WithGroup("g1")
		handler = handler.WithAttrs([]slog.Attr{slog.String("a2", "av2")})

		record1 := slog.NewRecord(someTime, slog.LevelDebug, "m1", 42)
		record1.AddAttrs(
			slog.String("a3", "av3"),
			slog.String("a4", "av4"),
		)

		handler.Handle(slogc.WithName(ctx, "ab"), record1)

		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallWithAttrs{"0", []testAttr{
				{"a1", "av1"},
			}},
			handlerCallWithGroup{"0.1", "g1"},
			handlerCallWithAttrs{"0.1.2", []testAttr{
				{"a2", "av2"},
			}},
			handlerCallHandle{"0.1.2.3", testRecord{
				Time:    someTime,
				Message: "m1",
				Level:   slog.LevelDebug,
				PC:      42,
				Attrs: []testAttr{
					{"a3", "av3"},
					{"a4", "av4"},
					{"@logger", "ab"},
				},
			}},
		}))
	})

	t.Run("WithoutName", func(t Test) {
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		handler.Handle(context.Background(), record1)

		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallHandle{"0", testRecord{
				Time:    someTime,
				Message: "m1",
				Level:   slog.LevelInfo,
				PC:      0,
				Attrs:   nil,
			}},
		}))
	})

	t.Run("WithName", func(t Test) {
		ctx := slogc.WithName(context.Background(), "a")
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		handler.Handle(slogc.WithName(ctx, "b"), record1)

		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallHandle{"0", testRecord{
				Time:    someTime,
				Message: "m1",
				Level:   slog.LevelInfo,
				PC:      0,
				Attrs:   []testAttr{{"@logger", "a.b"}},
			}},
		}))

		callLog.calls = nil
	})

	t.Run("WithEmptyName", func(t Test) {
		ctx := context.Background()
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		handler.Handle(slogc.WithName(ctx, ""), record1)

		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallHandle{"0", testRecord{
				Time:    someTime,
				Message: "m1",
				Level:   slog.LevelInfo,
				PC:      0,
				Attrs:   nil,
			}},
		}))
	})

	t.Run("WithAttrsAndName", func(t Test) {
		ctx := context.Background()
		callLog := &testCallLog{}
		base := &testHandler{"0", callLog}
		handler := slogc.HandlerWithNameAsAttr(base, "@logger")

		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)
		record1.AddAttrs(
			slog.Any("a1", "av1"),
			slog.Any("a2", "av2"),
		)

		handler = handler.WithAttrs([]slog.Attr{
			slog.Any("c1", "cv1"),
			slog.Any("c2", "cv2"),
		},
		)
		handler.Handle(slogc.WithName(ctx, "aa"), record1)

		t.Expect(callLog.calls).To(Equal([]any{
			handlerCallWithAttrs{"0", []testAttr{
				{"c1", "cv1"},
				{"c2", "cv2"},
			}},
			handlerCallHandle{"0.1", testRecord{
				Time:    someTime,
				Message: "m1",
				Level:   slog.LevelInfo,
				PC:      0,
				Attrs: []testAttr{
					{"a1", "av1"},
					{"a2", "av2"},
					{"@logger", "aa"},
				},
			}},
		}))

		callLog.calls = nil
	})
}

// ---

type testHandler struct {
	instance string
	log      *testCallLog
}

func (h testHandler) Enabled(_ context.Context, level slog.Level) bool {
	h.log.append(handlerCallEnabled{h.instance, level})

	return true
}

func (h testHandler) Handle(_ context.Context, record slog.Record) error {
	h.log.append(handlerCallHandle{h.instance, newTestRecord(record)})

	return nil
}

func (h testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.instance = h.newInstance(
		h.log.append(handlerCallWithAttrs{h.instance, newTestAttrs(attrs)}),
	)

	return &h
}

func (h testHandler) WithGroup(group string) slog.Handler {
	h.instance = h.newInstance(
		h.log.append(handlerCallWithGroup{h.instance, group}),
	)

	return &h
}

func (h testHandler) newInstance(n int) string {
	return fmt.Sprintf("%s.%d", h.instance, n)
}

// ---

type testCallLog struct {
	mu    sync.Mutex
	calls []any
}

func (l *testCallLog) append(call any) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.calls = append(l.calls, call)

	return len(l.calls)
}

// ---

type handlerCallEnabled struct {
	Instance string
	Level    slog.Level
}

type handlerCallHandle struct {
	Instance string
	Record   testRecord
}

type handlerCallWithAttrs struct {
	Instance string
	Attrs    []testAttr
}

type handlerCallWithGroup struct {
	Instance string
	Key      string
}

// ---

func newTestRecord(record slog.Record) testRecord {
	return testRecord{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level,
		PC:      record.PC,
		Attrs:   newTestAttrsForRecord(record),
	}
}

func newTestAttrsForRecord(record slog.Record) []testAttr {
	if record.NumAttrs() == 0 {
		return nil
	}

	attrs := make([]testAttr, 0, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, testAttr{
			Key:   a.Key,
			Value: a.Value.Any(),
		})

		return true
	})

	return attrs
}

func newTestAttrs(attrs []slog.Attr) []testAttr {
	if len(attrs) == 0 {
		return nil
	}

	testAttrs := make([]testAttr, 0, len(attrs))
	for _, a := range attrs {
		testAttrs = append(testAttrs, testAttr{
			Key:   a.Key,
			Value: a.Value.Any(),
		})
	}

	return testAttrs
}

type testRecord struct {
	Time    time.Time
	Message string
	Level   slog.Level
	PC      uintptr
	Attrs   []testAttr
}

type testAttr struct {
	Key   string
	Value any
}
