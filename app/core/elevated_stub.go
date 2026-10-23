//go:build !windows

package core

// IsElevated — не-Windows: не блокируем запуск (TUN на этих платформах не упакован в этот билд).
func IsElevated() bool { return true }
