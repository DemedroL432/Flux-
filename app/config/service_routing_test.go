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

// TestBuiltinRoutingPresetsAreTwoModes проверяет, что встроенных режимов ровно два
// и у них правильные финальные правила.
func TestBuiltinRoutingPresetsAreTwoModes(t *testing.T) {
	presets := builtinRoutingPresets()
	if len(presets) != 2 {
		t.Fatalf("expected exactly 2 builtin modes, got %d", len(presets))
	}
	byName := map[string]models.RoutingProfile{}
	for _, p := range presets {
		byName[p.Name] = p
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

// TestBothModesKeepRuDirect проверяет, что в обоих режимах российские домены
// и приватные сети уходят напрямую — иначе РФ-сайты поедут через туннель.
func TestBothModesKeepRuDirect(t *testing.T) {
	for _, p := range builtinRoutingPresets() {
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
}

// TestProcessRulesPresent проверяет, что правила по процессам есть в обоих режимах:
// они матчатся только по процессам и не пересекаются с доменными правилами.
func TestProcessRulesPresent(t *testing.T) {
	for _, p := range builtinRoutingPresets() {
		found := false
		for _, r := range p.Rules {
			if r.Name == "process-proxy-map" && len(r.Processes) > 0 {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s: process-proxy-map rule missing", p.Name)
		}
	}
}

// TestMergeBuiltinRoutingProfilesMigratesOldNames проверяет миграцию пяти старых
// профилей на два режима без потери пользовательских профилей.
func TestMergeBuiltinRoutingProfilesMigratesOldNames(t *testing.T) {
	cfg := &models.AppConfig{
		RoutingProfile: "rf-smart-split",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Final: "proxy"},
			{Name: "selective-proxy", Final: "direct"},
			{Name: "gaming-friendly", Final: "direct"},
			{Name: "rf-smart-split", Final: "direct"},
			{Name: "all-direct", Final: "direct"},
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
	if names[RoutingModeFullProxy] != 1 {
		t.Fatalf("full-proxy count = %d, want 1", names[RoutingModeFullProxy])
	}
	if names[RoutingModeManual] != 1 {
		t.Fatalf("manual count = %d, want 1", names[RoutingModeManual])
	}
	if names["my-custom"] != 1 {
		t.Fatal("user-created profile must survive migration")
	}
	for _, old := range []string{"default", "selective-proxy", "gaming-friendly", "rf-smart-split", "all-direct"} {
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
	if len(cfg.RoutingProfiles) != 2 {
		t.Fatalf("expected both modes after merge, got %d", len(cfg.RoutingProfiles))
	}
}
