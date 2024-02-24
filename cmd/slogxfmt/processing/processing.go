package processing

import (
	"bytes"
	"context"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
)

func NewProcessor(parser ParserFactory, handler HandlerFactory) *Processor {
	return &Processor{parser, handler}
}

// ---

type Processor struct {
	parser  ParserFactory
	handler HandlerFactory
}

func (p *Processor) Run(ctx context.Context, input <-chan *Buffer, output chan<- *Buffer) error {
	buf := model.NewBuffer()
	writer := *bytes.NewBuffer(*buf)

	parser := p.parser()
	handler := p.handler(&writer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case block, ok := <-input:
			if !ok {
				return nil
			}

			chunk, err := parser.Parse(*block)
			if err != nil {
				return err
			}
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
