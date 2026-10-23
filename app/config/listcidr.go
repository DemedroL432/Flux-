package config

import (
	"bufio"
	"net"
	"os"
	"strings"
)

// ReadZapretListCIDRs читает строки ipset/hostlist: комментарии # и пустые пропускает.
func ReadZapretListCIDRs(absPath string) ([]string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "/") {
			if _, _, err := net.ParseCIDR(line); err == nil {
				out = append(out, line)
			}
			continue
		}
		if ip := net.ParseIP(line); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			out = append(out, (&net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}).String())
		}
	}
	return out, sc.Err()
}
