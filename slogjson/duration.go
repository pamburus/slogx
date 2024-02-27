package slogjson

import (
	"log/slog"
	"time"

	"github.com/pamburus/slogx/internal/valenc"
)

// ---

// DurationAsText returns a DurationEncodeFunc that encodes time.Duration values as text with dynamic units.
// For example, it can be '1ms', '1s', '2m32s'.
func DurationAsText() DurationEncodeFunc {
	return func(_ []byte, v time.Duration) ([]byte, slog.Value) {
		return nil, slog.StringValue(v.String())
	}
}

// DurationAsSeconds returns a DurationEncodeFunc that encodes time.Duration values as floating point number of seconds.
func DurationAsSeconds() DurationEncodeFunc {
	return func(_ []byte, v time.Duration) ([]byte, slog.Value) {
		return nil, slog.Float64Value(v.Seconds())
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

	return func(buf []byte, v time.Duration) ([]byte, slog.Value) {
		return valenc.DurationAsHMS(buf, v, int(opts.precision)), slog.Value{}
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
