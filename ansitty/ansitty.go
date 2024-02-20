// Package ansitty provides a function to enable possibility to use ANSI escape sequences in TTY if possible.
package ansitty

import (
	"io"
	"os"
)

// Enable enables possibility to use ANSI escape sequences in TTY if possible.
func Enable(writer io.Writer) bool {
	if f, ok := writer.(*os.File); ok {
		return setEnabled(f.Fd(), true)
	}

	return false
}

// Disable disables possibility to use ANSI escape sequences in TTY if possible.
func Disable(writer io.Writer) bool {
	if f, ok := writer.(*os.File); ok {
		return setEnabled(f.Fd(), false)
	}

	return false
}

// Writer is a convenience function to enable possibility to use ANSI escape sequences in TTY if possible to writer.
// It calls [Enable] on the given writer and returns the same writer.
func Writer(writer io.Writer) io.Writer {
	Enable(writer)

	return writer
}

// Stdout enables possibility to use ANSI escape sequences in TTY if possible to stdout and returns os.Stdout.
func Stdout() *os.File {
	Enable(os.Stdout)

	return os.Stdout
}

// Stderr enables possibility to use ANSI escape sequences in TTY if possible to stderr and returns os.Stderr.
func Stderr() *os.File {
	Enable(os.Stderr)

	return os.Stderr
}
