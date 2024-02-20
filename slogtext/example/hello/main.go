// Package main is an example of how to use the slogtext package.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/pamburus/slogx/ansitty"
	"github.com/pamburus/slogx/slogc"
	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func main() {
	// Create handler.
	var handler slog.Handler = slogtext.NewHandler(os.Stdout,
		slogtext.WithLevel(slog.LevelDebug),
		slogtext.WithSource(true),
		slogtext.WithTheme(themes.Fancy()),
		slogtext.WithColor(ansitty.Enable),
	)

	// Create logger and set it as default.
	slog.SetDefault(slog.New(handler))

	logger := slog.New(handler).
		With(slog.String("x", "8")).
		WithGroup("main").
		With(slog.String("y", "9")).
		WithGroup("sub").
		With(slog.String("z", "10"))
	logger.Info("start")
	logger.Error("hello there",
		slog.Any("error", errors.New("failed to figure out what to do next")),
		slog.Int("int", 42),
		slog.Bool("bool", true),
		slog.Any("null", nil),
		slog.Float64("float", 3.14),
		slog.Group("aaa", slog.Attr{}),
		slog.Group("oo", slog.String("oo", ":)")),
	)

	// Do some logging.
	slog.Info("runtime info", slog.Int("cpu-count", runtime.NumCPU()))
	slog.Info("test array", slog.Any("ss", []string{"cpu-count", "abc"}))

	if info, ok := debug.ReadBuildInfo(); ok {
		slog.Debug("build info", slog.String("go-version", info.GoVersion), slog.String("path", info.Path))
		for _, setting := range info.Settings {
			slog.Debug("build setting", slog.String("key", setting.Key), slog.String("value", setting.Value))
		}
		for _, dep := range info.Deps {
			slog.Debug("dependency", slog.String("path", dep.Path), slog.String("version", dep.Version))
		}
	} else {
		slog.Warn("couldn't get build info")
	}

	doSomething(context.Background())
}

func doSomething(ctx context.Context) {
	ctx = slogc.WithName(ctx, "something")

	slogc.Info(ctx, "doing something")
	slogc.Error(ctx, "something bad happened doing something", slog.Any("error", errors.New("failed to figure out what to do next")))
}
