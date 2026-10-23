package config

import (
	"os"
	"path/filepath"
)

// ExecutableDir returns the absolute directory containing the running executable (not cwd).
func ExecutableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil && resolved != "" {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	out, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return filepath.Clean(dir), nil
	}
	return out, nil
}

// AbsPath normalizes p as an absolute filesystem path (best effort).
func AbsPath(p string) string {
	out, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return filepath.Clean(p)
	}
	return out
}
