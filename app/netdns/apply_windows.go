//go:build windows

package netdns

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// ApplyWindows sets DNS on all active adapters.
// preset: dhcp | cloudflare | google | quad9 | yandex | adguard | comss | xbox | opendns | custom
func ApplyWindows(preset, primary, secondary string) error {
	script := buildScript(strings.ToLower(strings.TrimSpace(preset)), primary, secondary)
	if script == "" {
		return fmt.Errorf("неверный пресет DNS")
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// PresetServers возвращает пару IPv4 DNS-серверов для выбранного пресета (или пустые строки для dhcp).
func PresetServers(preset, primary, secondary string) (string, string, bool) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "cloudflare":
		return "1.1.1.1", "1.0.0.1", true
	case "google":
		return "8.8.8.8", "8.8.4.4", true
	case "quad9":
		return "9.9.9.9", "149.112.112.112", true
	case "yandex":
		return "77.88.8.8", "77.88.8.1", true
	case "adguard":
		return "94.140.14.14", "94.140.15.15", true
	case "comss":
		return "83.220.169.155", "212.109.195.93", true
	case "xbox":
		return "111.88.96.50", "111.88.96.51", true
	case "opendns":
		return "208.67.222.222", "208.67.220.220", true
	case "custom":
		p := strings.TrimSpace(primary)
		s := strings.TrimSpace(secondary)
		if p == "" {
			return "", "", false
		}
		return p, s, true
	case "", "dhcp":
		return "", "", true
	default:
		return "", "", false
	}
}

func buildScript(preset, primary, secondary string) string {
	// Без -AddressFamily: на части систем старый NetTCPIP не знает этот параметр (ошибка ParameterBindingException).
	base := `$ErrorActionPreference='Stop'; Get-NetAdapter | Where-Object { $_.Status -eq 'Up' } | ForEach-Object { `
	p, s, ok := PresetServers(preset, primary, secondary)
	if !ok {
		return ""
	}
	if p == "" {
		return base + `Set-DnsClientServerAddress -InterfaceAlias $_.InterfaceAlias -ResetServerAddresses }`
	}
	if s == "" {
		return base + fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias $_.InterfaceAlias -ServerAddresses @('%s') }`, escapePS(p))
	}
	return base + fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias $_.InterfaceAlias -ServerAddresses @('%s','%s') }`, escapePS(p), escapePS(s))
}

func escapePS(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
