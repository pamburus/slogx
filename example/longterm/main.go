// Package main provides an example of using the slogx logger.
package main

import (
	"log/slog"
	"os"

	"github.com/pamburus/slogx"
)

func main() {
	logger := slogx.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	l1 := logger.With(slog.String("a", "av"), slog.String("b", "bv")).LongTerm()
	l1.Log(slog.LevelInfo, "msg", slog.String("c", "cv"), slog.String("d", "dv"), slog.String("e", "ev"))
}
