package bootlog

import (
	"os"
	"path/filepath"
	"time"

	"pulseshield/app/config"
)

func Append(line string) {
	dir, err := config.EnsureDataDir()
	if err != nil {
		return
	}
	path := filepath.Join(dir, "bootstrap.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format(time.RFC3339) + " " + line + "\n")
}
