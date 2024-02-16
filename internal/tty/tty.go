// Package tty provides utilities for dealing with platform-dependent virtual terminal.
package tty

import (
	"os"
)

// EnableSeqTTY enables possibility to use escape sequences in TTY if possible.
func EnableSeqTTY(f *os.File, flag bool) bool {
	return enableSeqTTY(f.Fd(), flag) == nil
}
