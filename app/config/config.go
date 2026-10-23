package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"pulseshield/app/models"
)

const defaultConfigFile = "app-config.json"

const dataDirName = "PulseShield"
const legacyDataDirName = "vpnboost"

func EnsureDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	local := filepath.Join(home, "AppData", "Local")
	newDir := filepath.Join(local, dataDirName)
	_ = migrateLegacyVpnboostIfNeeded(local, newDir)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return "", err
	}
	return newDir, nil
}

func migrateLegacyVpnboostIfNeeded(localRoot, newDir string) error {
	newCfg := filepath.Join(newDir, defaultConfigFile)
	if _, err := os.Stat(newCfg); err == nil {
		return nil
	}
	oldDir := filepath.Join(localRoot, legacyDataDirName)
	oldCfg := filepath.Join(oldDir, defaultConfigFile)
	if _, err := os.Stat(oldCfg); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return copyDirTree(oldDir, newDir)
}

func copyDirTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0o755)
		}
		target := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFileOne(path, target)
	})
}

func copyFileOne(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func ConfigPath() (string, error) {
	dataDir, err := EnsureDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, defaultConfigFile), nil
}

func DefaultConfig() models.AppConfig {
	return models.AppConfig{
		ProxyMode: models.ProxyModeSystem,
		ProxyPort: 8080,
		DnsPreset: "dhcp",
		ProcessProxyMap: map[string]bool{
			"Telegram.exe":       true,
			"telegram.exe":       true,
			"Cursor.exe":         true,
			"cursor.exe":         true,
			"Discord.exe":        true,
			"discord.exe":        true,
			"steam.exe":          true,
			"Steam.exe":          true,
			"steamwebhelper.exe": true,
		},
		// Zapret-профили подставляет SyncZapretPresetProfiles из zapret/presets/*.txt рядом с exe.
		DPIProfiles:       []models.DPIProfile{},
		ActiveDPIProfile:  PreferredZapretPresetName,
		RoutingProfiles:   builtinRoutingPresets(),
		RoutingProfile:    "default",
		RoutingBuiltinRev: CurrentRoutingBuiltinRev,
	}
}

func LoadOrCreate() (models.AppConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return models.AppConfig{}, err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		SyncZapretPresetProfiles(&cfg)
		_ = MergeBuiltinRoutingProfiles(&cfg)
		EnsureLocalInboundSecrets(&cfg)
		if err := Save(cfg); err != nil {
			return models.AppConfig{}, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return models.AppConfig{}, err
	}
	var cfg models.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Recover from partial/corrupted config files (e.g. after abrupt shutdown).
		backup := path + ".broken_" + time.Now().Format("20060102_150405")
		_ = os.WriteFile(backup, data, 0o644)
		fresh := DefaultConfig()
		SyncZapretPresetProfiles(&fresh)
		_ = MergeBuiltinRoutingProfiles(&fresh)
		EnsureLocalInboundSecrets(&fresh)
		_ = Save(fresh)
		return fresh, fmt.Errorf("config was corrupted; backed up to %s and recreated default", backup)
	}
	SyncZapretPresetProfiles(&cfg)
	changed := MergeBuiltinRoutingProfiles(&cfg)
	if ApplyRoutingBuiltinUpgrade(&cfg) {
		changed = true
	}
	if MergeExternalRoutingProfiles(&cfg) {
		changed = true
	}
	if EnsureLocalInboundSecrets(&cfg) {
		changed = true
	}
	if pruneUnsupportedVlessServers(&cfg) {
		changed = true
	}
	if changed {
		_ = Save(cfg)
	}
	return cfg, nil
}

// pruneUnsupportedVlessServers удаляет из конфига серверы с транспортами xhttp/splithttp,
// которые не поддерживаются sing-box (вызывают ошибку "unknown version: 72"), а также
// фиктивные инфо-записи (127.0.0.1 / 00000000-0000-0000-0000-000000000000) и переключает
// ActiveServerID на рабочий узел.
func pruneUnsupportedVlessServers(cfg *models.AppConfig) bool {
	if cfg == nil || len(cfg.VlessServers) == 0 {
		return false
	}
	var kept []models.VlessServer
	removed := false
	for _, s := range cfg.VlessServers {
		t := strings.ToLower(strings.TrimSpace(s.Params["type"]))
		h := strings.ToLower(strings.TrimSpace(s.Host))
		p := strings.TrimSpace(s.Port)
		if t == "xhttp" || t == "splithttp" || h == "127.0.0.1" || h == "0.0.0.0" || h == "localhost" || p == "1" || p == "0" || s.UUID == "00000000-0000-0000-0000-000000000000" {
			removed = true
			continue
		}
		kept = append(kept, s)
	}
	if !removed {
		return false
	}
	cfg.VlessServers = kept
	activeFound := false
	for _, s := range kept {
		if s.ID == cfg.ActiveServerID {
			activeFound = true
			break
		}
	}
	if !activeFound && len(kept) > 0 {
		for _, s := range kept {
			if s.Recommended {
				cfg.ActiveServerID = s.ID
				return true
			}
		}
		cfg.ActiveServerID = kept[0].ID
	}
	return true
}

func Save(cfg models.AppConfig) error {
	if err := ValidateAppConfig(cfg); err != nil {
		return err
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadDPIProfile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
