//go:build solaris && !appengine
// +build solaris,!appengine

package ansitty

import (
	"golang.org/x/sys/unix"
)

func setEnabled(fd uintptr, enabled bool) bool {
	if !enabled {
		return false
	}

	var termio unix.Termio

	err := unix.IoctlSetTermio(int(fd), unix.TCGETA, &termio)

	return err == nil
}
