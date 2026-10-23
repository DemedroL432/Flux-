//go:build windows

package winproxy

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// ResetWinHTTPProxy выполняет «netsh winhttp reset proxy» (службы WinHTTP).
func ResetWinHTTPProxy() error {
	cmd := exec.Command("netsh", "winhttp", "reset", "proxy")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd.Run()
}

// ForceResetWindowsProxy отключает пользовательский прокси в реестре CU и сбрасывает WinHTTP.
func ForceResetWindowsProxy() error {
	_ = Disable()
	if err := clearCUProxyStrings(); err != nil {
		return err
	}
	_ = ResetWinHTTPProxy()
	return nil
}

func clearCUProxyStrings() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	_ = k.DeleteValue("ProxyServer")
	_ = k.DeleteValue("ProxyOverride")
	return nil
}
