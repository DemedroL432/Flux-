package config

import (
	"testing"

	"pulseshield/app/models"
)

func TestProxyDpiXboxSuffixesPairwiseDisjoint(t *testing.T) {
	xbox := XboxDnsFriendlyDomainSuffixes()
	dpi := DPIBypassMediaDomainSuffixes()
	prox := ProxyOnlyMediaDomainSuffixes()
	ru := RuDirectDomainSuffixes()
	assertNoOverlap := func(a, b []string, an, bn string) {
		set := make(map[string]struct{}, len(b))
		for _, s := range b {
			set[s] = struct{}{}
		}
		for _, s := range a {
			if _, ok := set[s]; ok {
				t.Fatalf("%s and %s both contain suffix %q", an, bn, s)
			}
		}
	}
	assertNoOverlap(prox, dpi, "proxy-only", "dpi")
	assertNoOverlap(prox, xbox, "proxy-only", "xbox-dns")
	assertNoOverlap(xbox, dpi, "xbox-dns", "dpi")
	assertNoOverlap(prox, ru, "proxy-only", "ru-direct")
	assertNoOverlap(xbox, ru, "xbox-dns", "ru-direct")
}

// TestBuiltinRoutingPresetsAreThreeModes проверяет, что встроенных режимов три
// (global, full-proxy, manual) и у каждого корректное финальное правило.
func TestBuiltinRoutingPresetsAreThreeModes(t *testing.T) {
	presets := builtinRoutingPresets()
	if len(presets) != 3 {
		t.Fatalf("expected exactly 3 builtin modes, got %d", len(presets))
	}
	byName := map[string]models.RoutingProfile{}
	for _, p := range presets {
		byName[p.Name] = p
	}
	glob, ok := byName[RoutingModeGlobal]
	if !ok {
		t.Fatal("global mode missing")
	}
	if glob.Final != "proxy" {
		t.Fatalf("global final = %q, want proxy", glob.Final)
	}
	full, ok := byName[RoutingModeFullProxy]
	if !ok {
		t.Fatal("full-proxy mode missing")
	}
	if full.Final != "proxy" {
		t.Fatalf("full-proxy final = %q, want proxy", full.Final)
	}
	manual, ok := byName[RoutingModeManual]
	if !ok {
		t.Fatal("manual mode missing")
	}
	if manual.Final != "direct" {
		t.Fatalf("manual final = %q, want direct", manual.Final)
	}
}

// TestRuDirectAndGlobalIsolation проверяет, что full-proxy и manual оставляют РФ напрямую,
// а global содержит только локальные адреса (100% скрытие GEO).
func TestRuDirectAndGlobalIsolation(t *testing.T) {
	byName := map[string]models.RoutingProfile{}
	for _, p := range builtinRoutingPresets() {
		byName[p.Name] = p
	}

	for _, name := range []string{RoutingModeFullProxy, RoutingModeManual} {
		p := byName[name]
		var hasRuTLD, hasRuServices, hasLAN bool
		for _, r := range p.Rules {
			if r.Name == "ru-tld-direct" && r.Action == models.RouteDirect && len(r.DomainSuffixes) > 0 {
				hasRuTLD = true
			}
			if r.Name == "ru-services-direct" && r.Action == models.RouteDirect && len(r.Domains) > 0 {
				hasRuServices = true
			}
			if r.Name == "private-lan-direct" && r.Action == models.RouteDirect && len(r.CIDRs) > 0 {
				hasLAN = true
			}
		}
		if !hasRuTLD || !hasRuServices || !hasLAN {
			t.Fatalf("%s: ruTLD=%v ruServices=%v lan=%v", p.Name, hasRuTLD, hasRuServices, hasLAN)
		}
	}

	glob := byName[RoutingModeGlobal]
	for _, r := range glob.Rules {
		if r.Name == "ru-tld-direct" || r.Name == "ru-services-direct" || r.Name == "process-proxy-map" {
			t.Fatalf("global mode must not contain %q (breaks full GEO hiding)", r.Name)
		}
	}
}

// TestManualHasNoImposedProxyLists проверяет, что ни один встроенный режим не навязывает
// скрытое проксирование Steam, Discord, YouTube, нейросетей или process-proxy-map:
// пользователь сам выбирает правила или включает full-proxy/global.
func TestManualHasNoImposedProxyLists(t *testing.T) {
	byName := map[string]models.RoutingProfile{}
	for _, p := range builtinRoutingPresets() {
		byName[p.Name] = p
	}

	for _, name := range []string{RoutingModeGlobal, RoutingModeFullProxy, RoutingModeManual} {
		p := byName[name]
		for _, r := range p.Rules {
			if r.Name == "process-proxy-map" || r.Name == "ai-services-via-proxy" ||
				r.Name == "proxy-only-messengers-steam" || r.Name == "dpi-bypass-media" {
				t.Fatalf("%s must not impose rule %q", name, r.Name)
			}
			if r.Action == models.RouteProxy {
				t.Fatalf("%s must not contain built-in proxy rules (got %q)", name, r.Name)
			}
		}
	}
}

func TestApplyRoutingBuiltinUpgradeStripsLegacySteam(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RoutingBuiltinRev = 8
	cfg.ProcessProxyMap = map[string]bool{
		"steam.exe":          true,
		"Steam.exe":          true,
		"steamwebhelper.exe": true,
		"custom.exe":         true,
	}

	if !ApplyRoutingBuiltinUpgrade(&cfg) {
		t.Fatal("expected upgrade from rev 8 to rev 11")
	}
	if cfg.RoutingBuiltinRev != CurrentRoutingBuiltinRev {
		t.Fatalf("RoutingBuiltinRev = %d, want %d", cfg.RoutingBuiltinRev, CurrentRoutingBuiltinRev)
	}
	for _, k := range []string{"steam.exe", "Steam.exe", "steamwebhelper.exe"} {
		if _, ok := cfg.ProcessProxyMap[k]; ok {
			t.Fatalf("expected legacy %q stripped from ProcessProxyMap", k)
		}
	}
	if !cfg.ProcessProxyMap["custom.exe"] {
		t.Fatal("custom process entry must be preserved")
	}
}

// TestMergeBuiltinRoutingProfilesMigratesOldNames проверяет миграцию пяти старых
// профилей на три режима без потери пользовательских профилей.
func TestMergeBuiltinRoutingProfilesMigratesOldNames(t *testing.T) {
	cfg := &models.AppConfig{
		RoutingProfile: "rf-smart-split",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Final: "proxy"},
			{Name: "selective-proxy", Final: "direct"},
			{Name: "gaming-friendly", Final: "direct"},
			{Name: "rf-smart-split", Final: "direct"},
			{Name: "all-direct", Final: "direct"},
			{Name: "general-dpi-bypass", Final: "direct"},
			{Name: "my-custom", Final: "direct"},
		},
	}
	if !MergeBuiltinRoutingProfiles(cfg) {
		t.Fatal("expected migration to report change")
	}
	if cfg.RoutingProfile != RoutingModeManual {
		t.Fatalf("active profile = %q, want %q", cfg.RoutingProfile, RoutingModeManual)
	}
	names := map[string]int{}
	for _, p := range cfg.RoutingProfiles {
		names[p.Name]++
	}
	if names[RoutingModeGlobal] != 1 {
		t.Fatalf("global count = %d, want 1", names[RoutingModeGlobal])
	}
	if names[RoutingModeFullProxy] != 1 {
		t.Fatalf("full-proxy count = %d, want 1", names[RoutingModeFullProxy])
	}
	if names[RoutingModeManual] != 1 {
		t.Fatalf("manual count = %d, want 1", names[RoutingModeManual])
	}
	if names["my-custom"] != 1 {
		t.Fatal("user-created profile must survive migration")
	}
	for _, old := range []string{"default", "selective-proxy", "gaming-friendly", "rf-smart-split", "all-direct", "general-dpi-bypass"} {
		if names[old] != 0 {
			t.Fatalf("legacy profile %q must be gone, found %d", old, names[old])
		}
	}
}

// TestMergeBuiltinRoutingProfilesResetsUnknownActive проверяет, что активный профиль,
// которого нет в списке, сбрасывается на manual.
func TestMergeBuiltinRoutingProfilesResetsUnknownActive(t *testing.T) {
	cfg := &models.AppConfig{
		RoutingProfile:  "ghost-profile",
		RoutingProfiles: []models.RoutingProfile{{Name: "manual"}},
	}
	MergeBuiltinRoutingProfiles(cfg)
	if cfg.RoutingProfile != RoutingModeManual {
		t.Fatalf("active = %q, want manual", cfg.RoutingProfile)
	}
	if len(cfg.RoutingProfiles) != 3 {
		t.Fatalf("expected all 3 modes after merge, got %d", len(cfg.RoutingProfiles))
	}
}
