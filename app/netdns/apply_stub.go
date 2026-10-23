//go:build !windows

package netdns

import "fmt"

func ApplyWindows(preset, primary, secondary string) error {
	return fmt.Errorf("DNS: только Windows")
}

func PresetServers(preset, primary, secondary string) (string, string, bool) {
	return "", "", false
}
