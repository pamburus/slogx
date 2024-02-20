package ansitty

import (
	"syscall"
)

var (
	kernel32       *syscall.LazyDLL  = syscall.NewLazyDLL("Kernel32.dll")
	setConsoleMode *syscall.LazyProc = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminalProcessing enables virtual terminal sequences.
// https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences
const (
	enableVirtualTerminalProcessing uint32 = 0x4
)

// enableSeqTTY enables terminal sequence handling in Windows.
func setEnabled(fd uintptr, enabled bool) bool {
	var mode uint32
	err := syscall.GetConsoleMode(syscall.Handle(fd), &mode)
	if err != nil {
		return false
	}

	if enabled {
		mode |= enableVirtualTerminalProcessing
	} else {
		mode &= ^enableVirtualTerminalProcessing
	}

	r, _, errno := setConsoleMode.Call(fd, uintptr(mode))
	if r == 0 {
		return false
	}

	return true
}
