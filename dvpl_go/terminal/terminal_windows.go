//go:build windows

package terminal

import (
	"os"

	"golang.org/x/sys/windows"
)

func EnableVirtualTerminal() {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32

	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return
	}

	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

	_ = windows.SetConsoleMode(handle, mode)
}
