//go:build !windows

package winproxy

func ResetWinHTTPProxy() error { return nil }

func ForceResetWindowsProxy() error { return Disable() }
