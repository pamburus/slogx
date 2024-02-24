package processing

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
	"github.com/pamburus/slogx/cmd/slogxfmt/parsing"
	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

func Run(ctx context.Context, parser parsing.Parser, input <-chan *Buffer, output chan<- *Buffer) error {
	defer close(output)

	buf := model.NewBuffer()
	writer := *bytes.NewBuffer(*buf)

	handler := slogtext.NewHandler(&writer,
		slogtext.WithLevel(slog.LevelDebug),
		slogtext.WithColor(slogtext.ColorAlways),
		slogtext.WithSource(true),
		slogtext.WithLoggerNameKey("logger"),
		slogtext.WithTheme(themes.Fancy()),
	)

	for {
		// slog.Debug("processor: reading input block")
		select {
		case <-ctx.Done():
			// slog.Debug("processor: context done")
			return ctx.Err()
		case block, ok := <-input:
			if !ok {
				// slog.Debug("processor: end of stream")
				return nil
			}

			// slog.Debug("processor: got block from input", slog.Int("len", block.Len()), slog.Int("cap", block.Cap()))
			chunk, err := parser.Parse(*block)
			if err != nil {
				return err
			}
			for _, record := range chunk.Records {
				handler.Handle(ctx, record)
			}
			chunk.Free()

			*buf = writer.Bytes()

			// slog.Debug("processor: sending block to output")
			select {
			case <-ctx.Done():
				// slog.Debug("processor: context done")
				return ctx.Err()
			case output <- buf:
				// slog.Debug("processor: sent block to output")
			}

			buf = model.NewBuffer()
			writer = *bytes.NewBuffer(*buf)
		}
	}
}
