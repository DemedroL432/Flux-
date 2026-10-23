package core

import (
	"errors"
	"os"
	"path/filepath"
)

// ResolveBinaryPath finds a bundled core binary near the GUI executable.
// This allows shipping a single portable folder/installer without requiring
// users to install sing-box or xray manually.
func ResolveBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	baseDir := filepath.Dir(exe)

	candidates := []string{
		filepath.Join(baseDir, "core", "sing-box.exe"),
		filepath.Join(baseDir, "core", "xray.exe"),
		filepath.Join(baseDir, "sing-box.exe"),
		filepath.Join(baseDir, "xray.exe"),
	}
	candidates = append(candidates, devCandidates(baseDir, "third_party", "core", "sing-box.exe")...)
	candidates = append(candidates, devCandidates(baseDir, "third_party", "core", "xray.exe")...)

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			return c, nil
		}
	}
	return "", errors.New("core binary not found near app executable")
}

func devCandidates(baseDir string, parts ...string) []string {
	var out []string
	cur := baseDir
	for i := 0; i < 7; i++ {
		out = append(out, filepath.Join(append([]string{cur}, parts...)...))
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
	}
	return out
}
