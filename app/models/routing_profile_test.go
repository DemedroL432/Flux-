package models_test

import (
	"strings"
	"testing"

	"pulseshield/app/models"
)

func sampleRoutingProfiles() []models.RoutingProfile {
	return []models.RoutingProfile{
		{
			Name:  "default",
			Final: "proxy",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "default-proxy", Action: models.RouteProxy},
			},
		},
		{
			Name:  "rf-smart-split",
			Final: "direct",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "ru-tld-direct", DomainSuffixes: []string{"ru", "xn--p1ai", "su"}, Action: models.RouteDirect},
				{Name: "proxy-only-messengers", DomainSuffixes: []string{"t.me", "telegram.org"}, Action: models.RouteProxy},
				{Name: "dpi-direct-services", DomainSuffixes: []string{"youtube.com", "googlevideo.com"}, Action: models.RouteDPIBypass},
			},
		},
		{
			Name:  "all-direct",
			Final: "direct",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "all-direct-services", DomainSuffixes: []string{"youtube.com", "discord.com"}, Action: models.RouteDirect},
			},
		},
		{
			Name:  "gaming-friendly",
			Final: "direct",
			Rules: []models.RouteRule{
				{Name: "launchers-direct", Processes: []string{"cs2.exe", "dota2.exe"}, Action: models.RouteDirect},
			},
		},
	}
}

func TestResolveRoutingProfileAlias(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", "rf-smart-split"},
		{"   ", "rf-smart-split"},
		{"bypass-russia", "rf-smart-split"},
		{"BYPASS-RUSSIA", "rf-smart-split"},
		{"  bypass-russia  ", "rf-smart-split"},
		{"all-proxy", "default"},
		{"ALL-PROXY", "default"},
		{"  All-Proxy  ", "default"},
		{"all-direct", "all-direct"},
		{"ALL-DIRECT", "all-direct"},
		{"rf-smart-split", "rf-smart-split"},
		{"default", "default"},
		{"gaming-friendly", "gaming-friendly"},
		{"unknown", "unknown"},
		{"custom-profile", "custom-profile"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("input_"+tc.input, func(t *testing.T) {
			got := models.ResolveRoutingProfileAlias(tc.input)
			if got != tc.expected {
				t.Errorf("ResolveRoutingProfileAlias(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestActiveRoutingProfile_VariousInputs(t *testing.T) {
	profiles := sampleRoutingProfiles()

	testMatrix := []struct {
		name          string
		inputProfile  string
		expectedName  string
		expectedFinal string
		expectOK      bool
		minRules      int
	}{
		{
			name:          "empty string resolves to rf-smart-split",
			inputProfile:  "",
			expectedName:  "rf-smart-split",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      4,
		},
		{
			name:          "bypass-russia resolves to rf-smart-split",
			inputProfile:  "bypass-russia",
			expectedName:  "rf-smart-split",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      4,
		},
		{
			name:          "all-proxy resolves to default",
			inputProfile:  "all-proxy",
			expectedName:  "default",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "all-direct resolves to all-direct",
			inputProfile:  "all-direct",
			expectedName:  "all-direct",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "rf-smart-split direct match",
			inputProfile:  "rf-smart-split",
			expectedName:  "rf-smart-split",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      4,
		},
		{
			name:          "default direct match",
			inputProfile:  "default",
			expectedName:  "default",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "gaming-friendly direct match",
			inputProfile:  "gaming-friendly",
			expectedName:  "gaming-friendly",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      1,
		},
		{
			name:          "unknown profile falls back to default/first",
			inputProfile:  "unknown",
			expectedName:  "default",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
	}

	for _, tc := range testMatrix {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := models.AppConfig{
				RoutingProfile:  tc.inputProfile,
				RoutingProfiles: profiles,
			}

			rp, ok := cfg.ActiveRoutingProfile()
			if ok != tc.expectOK {
				t.Fatalf("ActiveRoutingProfile() ok = %v, expected %v", ok, tc.expectOK)
			}
			if !strings.EqualFold(rp.Name, tc.expectedName) {
				t.Errorf("ActiveRoutingProfile() Name = %q, expected %q", rp.Name, tc.expectedName)
			}
			if rp.Final != tc.expectedFinal {
				t.Errorf("ActiveRoutingProfile() Final = %q, expected %q", rp.Final, tc.expectedFinal)
			}
			if len(rp.Rules) < tc.minRules {
				t.Errorf("ActiveRoutingProfile() rule count = %d, expected at least %d", len(rp.Rules), tc.minRules)
			}
		})
	}
}

func TestActiveRoutingProfile_BypassRussiaDoesNotFallBackToEmptyOrBrokenRules(t *testing.T) {
	profiles := sampleRoutingProfiles()
	cfg := models.AppConfig{
		RoutingProfile:  "bypass-russia",
		RoutingProfiles: profiles,
	}

	rp, ok := cfg.ActiveRoutingProfile()
	if !ok {
		t.Fatal("expected ok == true for bypass-russia")
	}
	if rp.Name != "rf-smart-split" {
		t.Fatalf("expected resolved name 'rf-smart-split', got %q", rp.Name)
	}
	if len(rp.Rules) == 0 {
		t.Fatal("expected non-empty rules for rf-smart-split, got 0 rules")
	}
	if rp.Final != "direct" {
		t.Fatalf("expected Final == 'direct' for rf-smart-split, got %q", rp.Final)
	}

	// Verify specific rule contents exist and are not broken
	var hasRuDirect, hasProxyMessenger, hasDpiBypass bool
	for _, r := range rp.Rules {
		if r.Name == "ru-tld-direct" && r.Action == models.RouteDirect {
			hasRuDirect = true
		}
		if r.Name == "proxy-only-messengers" && r.Action == models.RouteProxy {
			hasProxyMessenger = true
		}
		if r.Name == "dpi-direct-services" && r.Action == models.RouteDPIBypass {
			hasDpiBypass = true
		}
	}

	if !hasRuDirect {
		t.Error("missing or broken ru-tld-direct rule")
	}
	if !hasProxyMessenger {
		t.Error("missing or broken proxy-only-messengers rule")
	}
	if !hasDpiBypass {
		t.Error("missing or broken dpi-direct-services rule")
	}
}

func TestActiveRoutingProfile_AllDirect_FinalOutboundDirect(t *testing.T) {
	profiles := sampleRoutingProfiles()
	cfg := models.AppConfig{
		RoutingProfile:  "all-direct",
		RoutingProfiles: profiles,
	}

	rp, ok := cfg.ActiveRoutingProfile()
	if !ok {
		t.Fatal("expected ok == true for all-direct")
	}
	if rp.Name != "all-direct" {
		t.Fatalf("expected Name == 'all-direct', got %q", rp.Name)
	}
	if rp.Final != "direct" {
		t.Fatalf("expected Final == 'direct', got %q", rp.Final)
	}
	if len(rp.Rules) == 0 {
		t.Fatal("expected non-empty rules for all-direct")
	}
}

func TestActiveRoutingProfile_EmptyConfigFallback(t *testing.T) {
	// Empty config with no RoutingProfiles and no RoutingProfile returns false
	emptyCfg := models.AppConfig{}
	_, ok := emptyCfg.ActiveRoutingProfile()
	if ok {
		t.Error("expected ok == false for empty config without profiles")
	}

	// Config with all-direct requested but empty RoutingProfiles slice triggers synthetic fallback
	synthDirect := models.AppConfig{RoutingProfile: "all-direct"}
	rpDirect, okDirect := synthDirect.ActiveRoutingProfile()
	if !okDirect {
		t.Fatal("expected synthetic fallback ok == true for all-direct")
	}
	if rpDirect.Final != "direct" {
		t.Fatalf("expected synthetic all-direct Final == 'direct', got %q", rpDirect.Final)
	}
	if len(rpDirect.Rules) == 0 {
		t.Fatal("expected synthetic all-direct to have non-empty rules")
	}

	// Config with bypass-russia requested but empty RoutingProfiles slice triggers synthetic fallback
	synthRf := models.AppConfig{RoutingProfile: "bypass-russia"}
	rpRf, okRf := synthRf.ActiveRoutingProfile()
	if !okRf {
		t.Fatal("expected synthetic fallback ok == true for bypass-russia")
	}
	if rpRf.Final != "direct" {
		t.Fatalf("expected synthetic rf-smart-split Final == 'direct', got %q", rpRf.Final)
	}
}
