//go:build !windows

package dpi

func verifyProcessAlive(pid int) bool {
	return pid > 0
}
