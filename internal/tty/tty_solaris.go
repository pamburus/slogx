//go:build solaris && !appengine
// +build solaris,!appengine

package tty

import (
	"golang.org/x/sys/unix"
)

func enableSeqTTY(fd uintptr, flag bool) error {
	var termio unix.Termio

	return unix.IoctlSetTermio(int(fd), unix.TCGETA, &termio)
}
