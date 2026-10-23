//go:build windows

package winproxy

import (
	"syscall"
)

var (
	wininet             = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption = wininet.NewProc("InternetSetOptionW")
)

const (
	internetOptionSettingsChanged = 39
	internetOptionRefresh         = 37
)

// NotifyProxyChange asks WinINet-aware apps to reload proxy settings from the registry.
func NotifyProxyChange() {
	h := uintptr(0)
	_, _, _ = procInternetSetOption.Call(h, uintptr(internetOptionSettingsChanged), 0, 0)
	_, _, _ = procInternetSetOption.Call(h, uintptr(internetOptionRefresh), 0, 0)
}
