package diagnostics

import (
	"os"
	"path/filepath"

	"pulseshield/app/config"
)

type Report struct {
	DataDir      string
	ConfigPath   string
	HasCore      bool
	HasDpiFolder bool
	LogsDir      string
}

func Run() (Report, error) {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return Report{}, err
	}
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return Report{}, err
	}
	logsDir := filepath.Join(dataDir, "logs")
	_ = os.MkdirAll(logsDir, 0o755)

	exe, err := os.Executable()
	if err != nil {
		return Report{}, err
	}
	base := filepath.Dir(exe)

	_, coreErr := os.Stat(filepath.Join(base, "core", "sing-box.exe"))
	_, dpiErr := os.Stat(filepath.Join(base, "dpi"))

	return Report{
		DataDir:      dataDir,
		ConfigPath:   cfgPath,
		HasCore:      coreErr == nil,
		HasDpiFolder: dpiErr == nil,
		LogsDir:      logsDir,
	}, nil
}

