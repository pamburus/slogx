package pipeline

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"sync"

	"github.com/pamburus/slogx/cmd/slogxfmt/parsing"
	"github.com/pamburus/slogx/cmd/slogxfmt/processing"
	"github.com/pamburus/slogx/cmd/slogxfmt/scanning"
)

func New(handler HandlerFactory) *Pipeline {
	return &Pipeline{
		handler,
		0,
		parsing.NewDefaultJSONParser,
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
	concurrency := p.concurrency
	if concurrency == 0 {
		concurrency = runtime.NumCPU()
	}

	errs := make(chan error, concurrency+2)
	defer func() {
		errSlice := make([]error, 0, concurrency+2)
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

	stream := make(chan *Buffer, concurrency)
	in := make([]chan *Buffer, concurrency)
	out := make([]chan *Buffer, concurrency)
	done := make([]chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		in[i] = make(chan *Buffer, 1)
		out[i] = make(chan *Buffer, 1)
		done[i] = make(chan struct{})
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Scan input and send blocks to the dispatcher.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(stream)

		slog.Debug("scanner: started")
		defer slog.Debug("scanner: stopped")

		scanner := scanning.NewScanner(input)
		for scanner.Next() {
			select {
			case <-ctx.Done():
				return
			case stream <- scanner.Block():
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	// Dispatch blocks to workers.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			for i := 0; i < concurrency; i++ {
				close(in[i])
			}
		}()

		slog.Debug("dispatcher: started")
		defer slog.Debug("dispatcher: stopped")

		for i := 0; ; i = (i + 1) % concurrency {
			select {
			case <-ctx.Done():
				return
			case block, ok := <-stream:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case in[i] <- block:
					continue
				case <-done[i]:
					continue
				}
			}
		}
	}()

	// Spawn workers.
	for i := 0; i < concurrency; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(done[i])
			defer close(out[i])

			logger := slog.Default().With(slog.Int("worker", i))

			logger.Debug("worker started")
			defer logger.Debug("worker stopped")

			processor := processing.NewProcessor(p.parser, p.handler).WithLogger(logger)
			err := processor.Run(ctx, in[i], out[i])
			if err != nil {
				errs <- err
			}
		}()
	}

	// Collect results from workers and write to output.
	for i := 0; ; i = (i + 1) % concurrency {
		select {
		case <-ctx.Done():
			return nil
		case block, ok := <-out[i]:
			if !ok {
				return nil
			}
			_, err := output.Write(*block)
			if err != nil {
				return err
			}
			block.Free()
		}
	}
}
