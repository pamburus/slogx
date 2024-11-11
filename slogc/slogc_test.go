package slogc_test

import (
	"context"
	"log/slog"
	"testing"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/internal/mock"
	"github.com/pamburus/slogx/slogc"
)

func TestLogger(tt *testing.T) {
	t := New(tt)

	t.Run("Default", func(t Test) {
		cl := mock.NewCallLog()
		handler := mock.NewHandler(cl)
		slog.SetDefault(slog.New(handler))
		ctx := context.Background()
		ctx = slogc.WithSource(ctx, false)
		logger := slogc.Get(ctx)

		logger.Info(ctx, "msg")
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

	test := func(name string, fn func(context.Context, *mock.CallLog, Test)) {
		cl := mock.NewCallLog()
		handler := mock.NewHandler(cl)
		logger := slogx.NewContextLogger(handler).WithSource(false)
		ctx := slogc.New(context.Background(), logger)

		t.Run(name, func(t Test) {
			fn(ctx, cl, t)
		})
	}

	test("Debug", func(ctx context.Context, cl *mock.CallLog, t Test) {
		slogc.Debug(ctx, "msg", slog.String("a", "v"))
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

	test("Info", func(ctx context.Context, cl *mock.CallLog, t Test) {
		slogc.Info(ctx, "msg", slog.String("a", "v"))
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

	test("Warn", func(ctx context.Context, cl *mock.CallLog, t Test) {
		slogc.Warn(ctx, "msg", slog.String("a", "v"))
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

	test("Error", func(ctx context.Context, cl *mock.CallLog, t Test) {
		slogc.Error(ctx, "msg", slog.String("a", "v"))
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

	test("Log", func(ctx context.Context, cl *mock.CallLog, t Test) {
		slogc.Log(ctx, slog.LevelWarn, "msg", slog.String("a", "v"))
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

	test("With", func(ctx context.Context, cl *mock.CallLog, t Test) {
		ctx = slogc.With(ctx,
			slog.String("a", "va"),
			slog.String("b", "vb"),
		)
		slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("c", "d"))
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

	test("WithGroup", func(ctx context.Context, cl *mock.CallLog, t Test) {
		ctx = slogc.WithGroup(ctx, "g1")
		slogc.Log(ctx, slog.LevelInfo, "msg", slog.String("a", "v"))
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
}
