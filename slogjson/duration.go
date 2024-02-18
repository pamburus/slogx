package slogjson

import (
	"log/slog"
	"strconv"
	"time"
)

// ---

// DurationAsText returns a DurationEncodeFunc that encodes time.Duration values as text with dynamic units.
// For example, it can be '1ms', '1s', '2m32s'.
func DurationAsText() DurationEncodeFunc {
	return func(v time.Duration) slog.Value {
		return slog.StringValue(v.String())
	}
}

// DurationAsSeconds returns a DurationEncodeFunc that encodes time.Duration values as floating point number of seconds.
func DurationAsSeconds() DurationEncodeFunc {
	return func(v time.Duration) slog.Value {
		return slog.Float64Value(v.Seconds())
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

	return func(v time.Duration) slog.Value {
		var bufStorage [12]byte
		buf := bufStorage[:0]

		if v < 0 {
			v = v.Abs()
			buf = append(buf, '-')
		}

		seconds := v % time.Minute
		minutes := int64((v % time.Hour) / time.Minute)
		hours := int64(v / time.Hour)

		if hours < 10 {
			buf = append(buf, '0')
		}
		buf = strconv.AppendInt(buf, hours, 10)

		buf = append(buf, ':')

		if minutes < 10 {
			buf = append(buf, '0')
		}
		buf = strconv.AppendInt(buf, minutes, 10)

		buf = append(buf, ':')

		if seconds < 10*time.Second {
			buf = append(buf, '0')
		}
		buf = strconv.AppendFloat(buf, seconds.Seconds(), 'f', int(opts.precision), 64)

		return slog.StringValue(string(buf))
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
