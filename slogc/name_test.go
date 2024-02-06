package slogc_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/internal/mock"
	"github.com/pamburus/slogx/slogc"
)

func TestName(tt *testing.T) {
	t := New(tt)

	someTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("EmptyAttrKey", func(t Test) {
		handler := &mock.Handler{}
		t.Expect(newHandlerWithName(handler, "")).To(Equal(handler))
	})

	t.Run("Enabled", func(t Test) {
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")

		t.Expect(handler.Enabled(context.Background(), slog.LevelInfo)).To(BeTrue())
		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerEnabled{Instance: "0", Level: slog.LevelInfo},
		))
	})

	t.Run("WithEmptyAttrs", func(t Test) {
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")

		t.Expect(handler.WithAttrs(nil)).To(Equal(handler))
	})

	t.Run("WithEmptyGroup", func(t Test) {
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")

		t.Expect(handler.WithGroup("")).To(Equal(handler))
	})

	t.Run("WithGroup", func(t Test) {
		ctx := context.Background()
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)

		handler := newHandlerWithName(base, "@logger")
		handler = handler.WithAttrs([]slog.Attr{slog.String("a1", "av1")})
		handler = handler.WithGroup("g1")
		handler = handler.WithAttrs([]slog.Attr{slog.String("a2", "av2")})

		record1 := slog.NewRecord(someTime, slog.LevelDebug, "m1", 42)
		record1.AddAttrs(
			slog.String("a3", "av3"),
			slog.String("a4", "av4"),
		)

		err := handler.Handle(slogc.WithName(ctx, "ab"), record1)
		t.Expect(err).ToNot(HaveOccurred())

		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerWithAttrs{
				Instance: "0",
				Attrs: []mock.Attr{
					{Key: "a1", Value: "av1"},
				},
			},
			mock.HandlerWithGroup{
				Instance: "0.1",
				Key:      "g1",
			},
			mock.HandlerWithAttrs{
				Instance: "0.1.2",
				Attrs: []mock.Attr{
					{Key: "a2", Value: "av2"},
				},
			},
			mock.HandlerHandle{
				Instance: "0.1.2.3",
				Record: mock.Record{
					Time:    someTime,
					Message: "m1",
					Level:   slog.LevelDebug,
					PC:      42,
					Attrs: []mock.Attr{
						{Key: "a3", Value: "av3"},
						{Key: "a4", Value: "av4"},
						{Key: "@logger", Value: "ab"},
					},
				},
			},
		))
	})

	t.Run("WithoutName", func(t Test) {
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")
		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		err := handler.Handle(context.Background(), record1)
		t.Expect(err).ToNot(HaveOccurred())
		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Time:    someTime,
					Message: "m1",
					Level:   slog.LevelInfo,
					PC:      0,
					Attrs:   nil,
				},
			},
		))
	})

	t.Run("WithName", func(t Test) {
		ctx := slogc.WithName(context.Background(), "a")
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")
		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		err := handler.Handle(slogc.WithName(ctx, "b"), record1)
		t.Expect(err).ToNot(HaveOccurred())
		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Time:    someTime,
					Message: "m1",
					Level:   slog.LevelInfo,
					PC:      0,
					Attrs: []mock.Attr{
						{Key: "@logger", Value: "a.b"},
					},
				},
			},
		))
	})

	t.Run("WithEmptyName", func(t Test) {
		ctx := context.Background()
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")
		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)

		err := handler.Handle(slogc.WithName(ctx, ""), record1)
		t.Expect(err).ToNot(HaveOccurred())
		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerHandle{
				Instance: "0",
				Record: mock.Record{
					Time:    someTime,
					Message: "m1",
					Level:   slog.LevelInfo,
					PC:      0,
					Attrs:   nil,
				},
			},
		))
	})

	t.Run("WithAttrsAndName", func(t Test) {
		ctx := context.Background()
		cl := mock.NewCallLog()
		base := mock.NewHandler(cl)
		handler := newHandlerWithName(base, "@logger")

		record1 := slog.NewRecord(someTime, slog.LevelInfo, "m1", 0)
		record1.AddAttrs(
			slog.Any("a1", "av1"),
			slog.Any("a2", "av2"),
		)

		handler = handler.WithAttrs([]slog.Attr{
			slog.Any("c1", "cv1"),
			slog.Any("c2", "cv2"),
		})
		err := handler.Handle(slogc.WithName(ctx, "aa"), record1)
		t.Expect(err).ToNot(HaveOccurred())
		t.Expect(cl.Calls()...).To(Equal(
			mock.HandlerWithAttrs{
				Instance: "0",
				Attrs: []mock.Attr{
					{Key: "c1", Value: "cv1"},
					{Key: "c2", Value: "cv2"},
				},
			},
			mock.HandlerHandle{
				Instance: "0.1",
				Record: mock.Record{
					Time:    someTime,
					Message: "m1",
					Level:   slog.LevelInfo,
					PC:      0,
					Attrs: []mock.Attr{
						{Key: "a1", Value: "av1"},
						{Key: "a2", Value: "av2"},
						{Key: "@logger", Value: "aa"},
					},
				},
			},
		))
	})
}

func newHandlerWithName(handler slog.Handler, attrKey string) slog.Handler {
	if attrKey == "" {
		return handler
	}

	return slogx.TweakHandler(handler).
		WithDynamicAttr(slogc.NameAttr(attrKey)).
		Result()
}
