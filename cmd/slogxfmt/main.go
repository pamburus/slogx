package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/pamburus/ansitty"
	"github.com/pamburus/slogx/cmd/slogxfmt/pipeline"
	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func main() {
	input := os.Stdin
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		input = file
	}

	color := slogtext.ColorAlways
	if ansitty.Enable(input) {
		color = slogtext.ColorAlways
	}

	handler := func(w io.Writer) slog.Handler {
		return slogtext.NewHandler(w,
			slogtext.WithLevel(slog.LevelDebug),
			slogtext.WithColor(color),
			slogtext.WithSource(true),
			slogtext.WithTheme(themes.Fancy()),
			slogtext.WithLoggerKey("logger"),
		)
	}
	slog.SetDefault(slog.New(handler(os.Stderr)))

	err := pipeline.New(handler).Run(context.Background(), input, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
