package jsonparser

import (
	"log/slog"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
	"github.com/pamburus/slogx/cmd/slogxfmt/parsing/stat"
)

type Chunk = model.Chunk
type Stat = stat.Stat
type LevelParseFunc = func(string) (slog.Level, error)
