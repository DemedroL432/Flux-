//go:build windows

package uiwin

import (
	"os"
	"os/exec"
	"path/filepath"
)

func openBundledRoutingGuide() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	path := filepath.Join(filepath.Dir(exe), "ROUTING_GUIDE.md")
	return exec.Command("notepad.exe", path).Start()
}

