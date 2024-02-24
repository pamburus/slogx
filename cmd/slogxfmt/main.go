package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/pamburus/slogx/cmd/slogxfmt/pipeline"
	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func main() {
	handler := slogtext.NewHandler(os.Stderr,
		slogtext.WithLevel(slog.LevelInfo),
		slogtext.WithColor(slogtext.ColorAlways),
		slogtext.WithSource(true),
		slogtext.WithTheme(themes.Fancy()),
	)
	slog.SetDefault(slog.New(handler))

	input := os.Stdin
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		input = file
	}

	err := pipeline.New().Run(context.Background(), input, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
