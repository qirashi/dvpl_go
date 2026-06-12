//go:build !windows

package terminal

func EnableVirtualTerminal() bool {
	// Linux
	return true
}
