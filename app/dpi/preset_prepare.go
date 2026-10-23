package dpi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

// PrepareWinwsPresetArg возвращает аргумент для winws2 (@path). При активной фильтрации списков
// пишет патч в %LocalAppData%/PulseShield/runtime/pulse-zapret-active.txt и возвращает абсолютный путь.
func PrepareWinwsPresetArg(zapretDir string, profile models.DPIProfile, cfg *models.AppConfig) (atArg string, err error) {
	if profile.Tool != "zapret_winws2" {
		return "@" + profile.ConfigPath, nil
	}
	rel := filepath.ToSlash(strings.TrimSpace(profile.ConfigPath))
	if cfg == nil || cfg.DpiListEnabled == nil || len(cfg.DpiListEnabled) == 0 {
		return "@" + rel, nil
	}

	src := filepath.Join(zapretDir, filepath.FromSlash(rel))
	raw, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("чтение пресета %s: %w", src, err)
	}
	patched := FilterPresetLinesByEnabledLists(string(raw), cfg)
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return "", err
	}
	runDir := filepath.Join(dataDir, "runtime")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", err
	}
	outPath := filepath.Join(runDir, "pulse-zapret-active.txt")
	if err := os.WriteFile(outPath, []byte(patched), 0o644); err != nil {
		return "", err
	}
	return "@" + outPath, nil
}
