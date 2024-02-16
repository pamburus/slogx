package slogtext

import (
	"encoding"
	"strconv"
)

// ---

// PrecisionAuto is a constant that can be used to request automatic precision selection.
const PrecisionAuto Precision = -1

// Precision can be used to specify exact or automatic precision for formatting floating point values.
type Precision int

// MarshalText implements encoding.TextMarshaler interfaces and allows precision to be converted into text.
func (p Precision) MarshalText() ([]byte, error) {
	if p < 0 {
		return []byte("auto"), nil
	}

	return strconv.AppendInt(nil, int64(p), 10), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interfaces and allows precision to be converted from text.
func (p *Precision) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "auto" {
		*p = -1
	} else {
		v, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}

		*p = Precision(v)
	}

	return nil
}

// ---

var (
	_ encoding.TextMarshaler   = Precision(0)
	_ encoding.TextUnmarshaler = &struct{ Precision }{}
)
