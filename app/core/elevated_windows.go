//go:build windows

package core

import "golang.org/x/sys/windows"

// IsElevated — true, если процесс запущен с повышенными правами (UAC), нужно для TUN/Wintun.
func IsElevated() bool {
	tok, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return false
	}
	defer tok.Close()
	return tok.IsElevated()
}
