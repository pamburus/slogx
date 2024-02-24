package pipeline

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/pamburus/slogx/cmd/slogxfmt/parsing"
	"github.com/pamburus/slogx/cmd/slogxfmt/processing"
	"github.com/pamburus/slogx/cmd/slogxfmt/scanning"
)

func New(handler HandlerFactory) *Pipeline {
	return &Pipeline{
		handler,
		runtime.NumCPU(),
		parsing.NewJSONParser(parsing.JSONParserConfig{}),
	}
}

type Pipeline struct {
	handler     HandlerFactory
	concurrency int
	parser      ParserFactory
}

func (p Pipeline) WithConcurrency(concurrency int) *Pipeline {
	p.concurrency = concurrency

	return &p
}

func (p Pipeline) WithParser(parser ParserFactory) *Pipeline {
	p.parser = parser

	return &p
}

func (p *Pipeline) Run(ctx context.Context, input io.Reader, output io.Writer) (err error) {
	errs := make(chan error, p.concurrency+2)
	defer func() {
		errSlice := make([]error, 0, p.concurrency+2)
		if err != nil {
			errSlice = append(errSlice, err)
		}
		close(errs)
		for e := range errs {
			if e != nil {
				errSlice = append(errSlice, e)
			}
		}
		switch len(errSlice) {
		case 0:
		case 1:
			err = errSlice[0]
		default:
			err = errors.Join(errSlice...)
		}
	}()

	sch := make(chan *Buffer, p.concurrency)
	ichs := make([]chan *Buffer, p.concurrency)
	ochs := make([]chan *Buffer, p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		ichs[i] = make(chan *Buffer, 1)
		ochs[i] = make(chan *Buffer, 1)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Scan input and send blocks to the dispatcher.
	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Debug("scanner: started")
		defer slog.Debug("scanner: stopped")

		err := scanning.Scan(ctx, input, sch)
		if err != nil {
			errs <- err
		}
	}()

	// Dispatch blocks to workers.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			for i := 0; i < p.concurrency; i++ {
				close(ichs[i])
			}
		}()

		slog.Debug("dispatcher: started")
		defer slog.Debug("dispatcher: stopped")

		for i := 0; ; i = (i + 1) % p.concurrency {
			// slog.Debug("dispatcher: reading input block")
			select {
			case <-ctx.Done():
				// slog.Debug("dispatcher: context done")
				return
			case block, ok := <-sch:
				if !ok {
					// slog.Debug("dispatcher: end of stream")
					return
				}
				// slog.Debug("dispatcher: sending output block", slog.Int("id", i), slog.Int("len", block.Len()), slog.Int("cap", block.Cap()))
				select {
				case <-ctx.Done():
					// slog.Debug("dispatcher: context done")
					return
				case ichs[i] <- block:
					// slog.Debug("dispatcher: sent output block", slog.Int("id", i))
				}
			}
		}
	}()

	// Spawn workers.
	for i := 0; i < p.concurrency; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			// slog.Debug("worker: started", slog.Int("id", i))
			// defer slog.Debug("worker: stopped", slog.Int("id", i))
			processor := processing.NewProcessor(p.parser, p.handler)
			err := processor.Run(ctx, ichs[i], ochs[i])
			if err != nil {
				errs <- err
			}
		}()
	}

	// Collect results from workers and write to output.
	for i := 0; ; i = (i + 1) % p.concurrency {
		// slog.Debug("writer: reading block from channel", slog.Int("id", i))
		select {
		case <-ctx.Done():
			// slog.Debug("writer: context done")
			return nil
		case block, ok := <-ochs[i]:
			if !ok {
				// slog.Debug("writer: end of stream")
				return nil
			}
			// slog.Debug("writer: writing block to stdout", slog.Int("id", i), slog.Int("len", block.Len()), slog.Int("cap", block.Cap()))
			_, err := os.Stdout.Write(*block)
			if err != nil {
				return err
			}
			block.Free()
		}
	}
}
