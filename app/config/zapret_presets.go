package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pulseshield/app/models"
)

// PreferredZapretPresetName — дефолтный пресет winws2 (имя после SyncZapretPresetProfiles).
const PreferredZapretPresetName = "zapret:Gamingnew"

// SyncZapretPresetProfiles loads bundled zapret presets from ./zapret/presets/*.txt
// and makes them available as DPI profiles (Tool=zapret_winws2, ConfigPath=presets/<file>).
func SyncZapretPresetProfiles(cfg *models.AppConfig) {
	presets := listBundledZapretPresets()
	if len(presets) == 0 {
		return
	}

	// Keep non-zapret profiles, replace zapret profiles with actual preset list.
	out := make([]models.DPIProfile, 0, len(cfg.DPIProfiles)+len(presets))
	for _, p := range cfg.DPIProfiles {
		if p.Tool == "zapret_winws2" {
			continue
		}
		out = append(out, p)
	}
	out = append(out, presets...)
	cfg.DPIProfiles = out

	// Ensure active profile still exists; предпочитаем Gamingnew, если он есть.
	if cfg.ActiveDPIProfile == "" || !hasDPIProfile(*cfg, cfg.ActiveDPIProfile) {
		if hasDPIProfile(*cfg, PreferredZapretPresetName) {
			cfg.ActiveDPIProfile = PreferredZapretPresetName
		} else if len(presets) > 0 {
			cfg.ActiveDPIProfile = presets[0].Name
		}
	}
}

func listBundledZapretPresets() []models.DPIProfile {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	base := filepath.Dir(exe)
	presetsDir := filepath.Join(base, "zapret", "presets")
	var relFiles []string
	_ = filepath.WalkDir(presetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		n := d.Name()
		if !strings.HasSuffix(strings.ToLower(n), ".txt") {
			return nil
		}
		rel, err := filepath.Rel(presetsDir, path)
		if err != nil {
			return nil
		}
		relFiles = append(relFiles, rel)
		return nil
	})
	sort.Strings(relFiles)

	out := make([]models.DPIProfile, 0, len(relFiles))
	for _, rel := range relFiles {
		baseName := strings.TrimSuffix(rel, filepath.Ext(rel))
		id := strings.ReplaceAll(filepath.ToSlash(baseName), "/", ":")
		out = append(out, models.DPIProfile{
			Name:       "zapret:" + id,
			Tool:       "zapret_winws2",
			ConfigPath: filepath.ToSlash(filepath.Join("presets", rel)),
			Args:       []string{},
		})
	}
	return out
}

func hasDPIProfile(cfg models.AppConfig, name string) bool {
	for _, p := range cfg.DPIProfiles {
		if p.Name == name {
			return true
		}
	}
	return false
}

