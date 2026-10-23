package dpi

import (
	"fmt"
	"strings"
	"unicode"
)

// PulseShieldGeneratedHeader — комментарий в начале сгенерированного пресета.
func PulseShieldGeneratedHeader(variantID, humanLabel string) string {
	return fmt.Sprintf(
		"# PulseShield — пресет сгенерирован автоматически\n# Вариант: %s — %s\n# Пути @lua, @bin, @windivert — относительно каталога zapret рядом с программой.\n\n",
		variantID, humanLabel,
	)
}

// TrimLeadingPresetComments убирает блок комментариев # в начале файла (из встроенного эталона), чтобы не дублировать метаданные.
func TrimLeadingPresetComments(raw []byte) []byte {
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	i := 0
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, "#") {
			i++
			continue
		}
		break
	}
	return []byte(strings.Join(lines[i:], "\n"))
}

// ReplaceWFPorts подменяет строки --wf-tcp-out и --wf-udp-out (пустая строка = не менять).
func ReplaceWFPorts(body []byte, tcpPorts, udpPorts string) []byte {
	s := strings.ReplaceAll(string(body), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if tcpPorts != "" && strings.HasPrefix(t, "--wf-tcp-out=") {
			lines[i] = "--wf-tcp-out=" + tcpPorts
		}
		if udpPorts != "" && strings.HasPrefix(t, "--wf-udp-out=") {
			lines[i] = "--wf-udp-out=" + udpPorts
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// ReplaceIPCacheLifetime подменяет --ipcache-lifetime= (пусто = не трогать).
func ReplaceIPCacheLifetime(body []byte, seconds string) []byte {
	if seconds == "" {
		return append([]byte(nil), body...)
	}
	s := strings.ReplaceAll(string(body), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "--ipcache-lifetime=") {
			lines[i] = "--ipcache-lifetime=" + seconds
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// SanitizeTuneFileName — безопасное имя для временного файла подбора.
func SanitizeTuneFileName(id string) string {
	var b strings.Builder
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s := b.String()
	if s == "" {
		return "variant"
	}
	if len(s) > 48 {
		s = s[:48]
	}
	return s
}
