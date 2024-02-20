// Package ansitty provides a function to enable possibility to use ANSI escape sequences in TTY if possible.
package ansitty

import (
	"io"
	"os"
)

// Enable enables possibility to use ANSI escape sequences in TTY if possible.
func Enable(stream io.Writer) bool {
	if f, ok := stream.(*os.File); ok {
		return setEnabled(f.Fd(), true)
	}

	return false
}

// Disable disables possibility to use ANSI escape sequences in TTY if possible.
func Disable(stream io.Writer) bool {
	if f, ok := stream.(*os.File); ok {
		return setEnabled(f.Fd(), false)
	}

	return false
}
