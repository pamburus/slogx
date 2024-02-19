// Package valenc contains encoding functions for various types.
package valenc

import (
	"strconv"
	"time"
)

// DurationAsHMS encodes the duration as hours, minutes, and seconds.
func DurationAsHMS(buf []byte, value time.Duration, precision int) []byte {
	if value < 0 {
		value = value.Abs()
		buf = append(buf, '-')
	}

	seconds := value % time.Minute
	minutes := int64((value % time.Hour) / time.Minute)
	hours := int64(value / time.Hour)

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
	buf = strconv.AppendFloat(buf, seconds.Seconds(), 'f', precision, 64)

	return buf
}
