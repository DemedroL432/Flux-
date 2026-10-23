package config

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
	"pulseshield/app/models"
)

// RoutingYAMLDir — поддиректория рядом с exe, где лежат пользовательские профили *.yaml.
const RoutingYAMLDir = "routing"

// LoadExternalRoutingProfiles читает *.yaml / *.yml из <exeDir>/routing/.
// Файлы сортируются по имени; ошибки разбора отдельных файлов молча пропускаются.
func LoadExternalRoutingProfiles() []models.RoutingProfile {
	dir, err := externalRoutingDir()
	if err != nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	var out []models.RoutingProfile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var p models.RoutingProfile
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		if p.Name == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// MergeExternalRoutingProfiles загружает YAML-профили из <exeDir>/routing/ и подмешивает их в cfg.
// Профиль с тем же именем, что встроенный, заменяет его — так можно переопределить любой пресет.
// Новые имена добавляются как пользовательские профили.
// Возвращает true, если cfg был изменён.
func MergeExternalRoutingProfiles(cfg *models.AppConfig) bool {
	if cfg == nil {
		return false
	}
	profiles := LoadExternalRoutingProfiles()
	if len(profiles) == 0 {
		return false
	}
	byIdx := make(map[string]int, len(cfg.RoutingProfiles))
	for i, p := range cfg.RoutingProfiles {
		byIdx[p.Name] = i
	}
	changed := false
	for _, yp := range profiles {
		if idx, ok := byIdx[yp.Name]; ok {
			cfg.RoutingProfiles[idx] = yp
		} else {
			cfg.RoutingProfiles = append(cfg.RoutingProfiles, yp)
			byIdx[yp.Name] = len(cfg.RoutingProfiles) - 1
		}
		changed = true
	}
	return changed
}

func externalRoutingDir() (string, error) {
	exe, err := ExecutableDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(exe, RoutingYAMLDir), nil
}
