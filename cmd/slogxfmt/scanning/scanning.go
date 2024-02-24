package scanning

import (
	"bytes"
	"io"
	"slices"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
)

func NewScanner(reader io.Reader) *Scanner {
	return &Scanner{reader: reader}
}

type Scanner struct {
	reader io.Reader
	buf    *Buffer
	next   *Buffer
	err    error
}

func (s *Scanner) Next() bool {
	if s.err != nil {
		return false
	}

	if s.next != nil {
		s.buf = s.next
		s.next = nil
	} else {
		s.buf = model.NewBuffer()
	}

	return s.scan()
}

func (s *Scanner) Block() *Buffer {
	return s.buf
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}

	return s.err
}

func (s *Scanner) scan() bool {
	foundNewLine := false
	for !foundNewLine {
		if len(s.buf.Tail()) == 0 {
			*s.buf = slices.Grow(*s.buf, s.buf.Cap())
		}

		var n int
		n, s.err = s.reader.Read(s.buf.Tail())
		if n <= 0 {
			return s.buf.Len() > 0
		}

		begin := s.buf.Len()
		end := n + begin
		*s.buf = (*s.buf)[:end]
		i := bytes.LastIndexByte((*s.buf)[begin:], '\n')
		if i >= 0 {
			foundNewLine = true
			s.next = model.NewBuffer()
			*s.next = append(*s.next, (*s.buf)[begin+i+1:]...)
			*s.buf = (*s.buf)[:begin+i+1]
		}
	}

	return true
}
