package slogtext

import (
	"strconv"
	"time"

	"github.com/pamburus/slogx/internal/valenc"
)

// ---

// DurationAsText returns a DurationEncodeFunc that encodes time.Duration values as text with dynamic units.
// For example, it can be '1ms', '1s', '2m32s'.
func DurationAsText() DurationEncodeFunc {
	return func(buf []byte, v time.Duration) []byte {
		return append(buf, v.String()...)
	}
}

// DurationAsSeconds returns a DurationEncodeFunc that encodes time.Duration values as floating point number of seconds.
func DurationAsSeconds(options ...DurationOption) DurationEncodeFunc {
	opts := defaultDurationOptions().With(options)

	return func(buf []byte, v time.Duration) []byte {
		return strconv.AppendFloat(buf, v.Seconds(), 'f', int(opts.precision), 64)
	}
}

// DurationAsHMS returns a DurationEncodeFunc that encodes time.Duration values as 'HH:MM:SS.sss' where
//
//	HH is number of hours (minimum 2 digits),
//	MM is number of minutes (always 2 digits),
//	SS is number of seconds (always 2 digits),
//	sss is fractional part of seconds (depends on Precision option).
func DurationAsHMS(options ...DurationOption) DurationEncodeFunc {
	opts := defaultDurationOptions().With(options)

	return func(buf []byte, v time.Duration) []byte {
		return valenc.DurationAsHMS(buf, v, int(opts.precision))
	}
}

// WithDurationPrecision is a DurationOption that sets the precision for DurationAsSeconds and DurationAsHMS.
func WithDurationPrecision(p Precision) DurationOption {
	return func(o *durationOptions) {
		o.precision = p
	}
}

// ---

// DurationOption is an optional parameter for DurationAsSeconds and DurationAsHMS.
// Implemented by:
//   - Precision.
type DurationOption func(*durationOptions)

// ---

func defaultDurationOptions() durationOptions {
	return durationOptions{
		precision: PrecisionAuto,
	}
}

type durationOptions struct {
	precision Precision
}

func (o durationOptions) With(other []DurationOption) durationOptions {
	for _, oo := range other {
		oo(&o)
	}

	return o
}
