package slogc_test

import (
	"context"
	"log/slog"

	. "github.com/pamburus/go-tst/tst"
	"github.com/pamburus/slogx"
	"github.com/pamburus/slogx/slogc"
	"github.com/pamburus/slogx/slogc/internal/mock"

	"testing"
)

func TestNewGet(tt *testing.T) {
	t := New(tt)

	t.Run("New", func(t Test) {
		cl := mock.NewCallLog()
		handler := mock.NewHandler(cl)
		logger := slogx.NewContextLogger(handler).WithSource(false)
		ctx := slogc.New(context.Background(), logger)

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
}
