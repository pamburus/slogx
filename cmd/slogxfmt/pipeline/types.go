package pipeline

import (
	"github.com/pamburus/slogx/cmd/slogxfmt/model"
	"github.com/pamburus/slogx/cmd/slogxfmt/parsing"
	"github.com/pamburus/slogx/cmd/slogxfmt/processing"
)

type Buffer = model.Buffer
type Parser = parsing.Parser
type ParserFactory = processing.ParserFactory
type HandlerFactory = processing.HandlerFactory
