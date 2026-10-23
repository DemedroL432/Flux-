package core

import (
	"os"
	"path/filepath"
	"time"

	"pulseshield/app/config"
)

func OpenLogFile(name string) (*os.File, string, error) {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return nil, "", err
	}
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, "", err
	}
	path := filepath.Join(logDir, time.Now().Format("20060102_150405")+"_"+name+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, "", err
	}
	return f, path, nil
}

