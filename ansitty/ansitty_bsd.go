//go:build (darwin || freebsd || openbsd || netbsd || dragonfly) && !appengine
// +build darwin freebsd openbsd netbsd dragonfly
// +build !appengine

package ansitty

import (
	"syscall"
	"unsafe"
)

func setEnabled(fd uintptr, enabled bool) bool {
	if !enabled {
		return false
	}

	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0) //nolint:gosec

	return errno == 0
}
