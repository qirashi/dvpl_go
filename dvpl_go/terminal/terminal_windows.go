//go:build windows

package terminal

import (
	"os"

	"golang.org/x/sys/windows"
)

func EnableVirtualTerminal() bool {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return false
	}

	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		return false
	}

	return true
}
