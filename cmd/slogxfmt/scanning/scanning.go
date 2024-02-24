package scanning

import (
	"bytes"
	"context"
	"io"
	"slices"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
)

func Scan(ctx context.Context, reader io.Reader, sink chan<- *Buffer) error {
	defer close(sink)

	buf := model.NewBuffer()

	var next *Buffer
	var err error
	var done bool

	for {
		if buf.Len() == 0 {
			select {
			case <-ctx.Done():
				// slog.Debug("scanner: context done")
				return ctx.Err()
			default:
			}
		} else {
			// slog.Debug("scanner: sending buffer to sink", slog.Int("len", buf.Len()), slog.Int("cap", buf.Cap()))
			select {
			case <-ctx.Done():
				// slog.Debug("scanner: context done")
				return ctx.Err()
			case sink <- buf:
				// slog.Debug("scanner: buffer sent to sink")
				buf = next
				if buf == nil {
					// slog.Debug("scanner: allocating new buffer")
					buf = model.NewBuffer()
				} else {
					// slog.Debug("scanner: using new buffer", slog.Int("len", buf.Len()), slog.Int("cap", buf.Cap()))
				}
				next = nil
			}
		}

		if err == io.EOF {
			// slog.Debug("scanner: eof")
			return nil
		}
		if err != nil {
			// slog.Debug("scanner: error", slog.Any("error", err))
			return err
		}
		if done {
			// slog.Debug("scanner: done")
			return nil
		}

		foundNewLine := false
		for !foundNewLine {
			if len(buf.Tail()) == 0 {
				// slog.Debug("scanner: grow buffer")
				*buf = slices.Grow(*buf, buf.Cap())
			}

			// slog.Debug("scanner: reading up to n bytes", slog.Int("n", len(buf.Tail())))
			var n int
			n, err = reader.Read(buf.Tail())
			if n <= 0 {
				// slog.Debug("scanner: end of stream")

				break
			}

			// slog.Debug("scanner: read n bytes", slog.Int("n", n))

			begin := buf.Len()
			end := n + begin
			*buf = (*buf)[:end]
			i := bytes.LastIndexByte((*buf)[begin:], '\n')
			if i >= 0 {
				foundNewLine = true
				next = model.NewBuffer()
				*next = append(*next, (*buf)[begin+i+1:]...)
				*buf = (*buf)[:begin+i+1]
			}
		}
	}
}
