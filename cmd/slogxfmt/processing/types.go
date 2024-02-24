package processing

import (
	"io"
	"log/slog"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
	"github.com/pamburus/slogx/cmd/slogxfmt/parsing"
)

type Buffer = model.Buffer
type Parser = parsing.Parser

type ParserFactory func() Parser
type HandlerFactory func(io.Writer) slog.Handler
