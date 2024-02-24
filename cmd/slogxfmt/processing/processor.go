package processing

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
)

func NewProcessor(parser ParserFactory, handler HandlerFactory) *Processor {
	return &Processor{parser, handler, slog.Default()}
}

// ---

type Processor struct {
	parser  ParserFactory
	handler HandlerFactory
	logger  *slog.Logger
}

func (p Processor) WithLogger(logger *slog.Logger) *Processor {
	p.logger = logger

	return &p
}

func (p *Processor) Run(ctx context.Context, input <-chan *Buffer, output chan<- *Buffer) error {
	buf := model.NewBuffer()
	writer := *bytes.NewBuffer(*buf)

	parser := p.parser()
	handler := p.handler(&writer)

	defer func() {
		stat := parser.Stat()
		p.logger.Debug("parser stat", slog.Any("stat", stat))
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case block, ok := <-input:
			if !ok {
				return nil
			}

			chunk := parser.Parse(*block)
			for _, record := range chunk.Records {
				handler.Handle(ctx, record)
			}
			chunk.Free()

			*buf = writer.Bytes()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case output <- buf:
				buf = model.NewBuffer()
				writer = *bytes.NewBuffer(*buf)
			}
		}
	}
}
