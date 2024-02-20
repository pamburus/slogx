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

	test("SlogLogger", func(t Test, _ *mock.CallLog, logger *slogx.Logger) {
		t.Expect(logger.SlogLogger().Handler()).To(Equal(logger.Handler()))
	})
}
