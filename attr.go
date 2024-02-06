package slogx

import "log/slog"

// ErrorAttr returns an attribute with the error.
func ErrorAttr(err error) slog.Attr {
	return slog.Any(ErrorKey, err)
}

const (
	// ErrorKey is the key used for the error attribute.
	ErrorKey = "error"
)
