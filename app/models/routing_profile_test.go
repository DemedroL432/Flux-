package models_test

import (
	"strings"
	"testing"

	"pulseshield/app/models"
)

func sampleRoutingProfiles() []models.RoutingProfile {
	return []models.RoutingProfile{
		{
			Name:  "full-proxy",
			Final: "proxy",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "default-proxy", Action: models.RouteProxy},
			},
		},
		{
			Name:  "manual",
			Final: "direct",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "ru-tld-direct", DomainSuffixes: []string{"ru", "xn--p1ai", "su"}, Action: models.RouteDirect},
				{Name: "proxy-only-messengers", DomainSuffixes: []string{"t.me", "telegram.org"}, Action: models.RouteProxy},
				{Name: "dpi-direct-services", DomainSuffixes: []string{"youtube.com", "googlevideo.com"}, Action: models.RouteDPIBypass},
			},
		},
	}
}

func TestResolveRoutingProfileAlias(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", "full-proxy"},
		{"   ", "full-proxy"},
		{"bypass-russia", "manual"},
		{"BYPASS-RUSSIA", "manual"},
		{"  bypass-russia  ", "manual"},
		{"all-proxy", "full-proxy"},
		{"ALL-PROXY", "full-proxy"},
		{"  All-Proxy  ", "full-proxy"},
		{"all-direct", "manual"},
		{"ALL-DIRECT", "manual"},
		{"rf-smart-split", "manual"},
		{"selective-proxy", "manual"},
		{"default", "full-proxy"},
		{"full-proxy", "full-proxy"},
		{"global", "global"},
		{"global-proxy", "global"},
		{"manual", "manual"},
		{"gaming-friendly", "manual"},
		{"general-dpi-bypass", "manual"},
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
			name:          "empty string resolves to full-proxy",
			inputProfile:  "",
			expectedName:  "full-proxy",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "bypass-russia resolves to manual",
			inputProfile:  "bypass-russia",
			expectedName:  "manual",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "all-proxy resolves to full-proxy",
			inputProfile:  "all-proxy",
			expectedName:  "full-proxy",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "all-direct resolves to manual",
			inputProfile:  "all-direct",
			expectedName:  "manual",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "rf-smart-split resolves to manual",
			inputProfile:  "rf-smart-split",
			expectedName:  "manual",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "default direct match",
			inputProfile:  "default",
			expectedName:  "full-proxy",
			expectedFinal: "proxy",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "gaming-friendly resolves to manual",
			inputProfile:  "gaming-friendly",
			expectedName:  "manual",
			expectedFinal: "direct",
			expectOK:      true,
			minRules:      2,
		},
		{
			name:          "unknown profile falls back to first profile",
			inputProfile:  "unknown",
			expectedName:  "full-proxy",
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
	if rp.Name != "manual" {
		t.Fatalf("expected resolved name 'manual', got %q", rp.Name)
	}
	if len(rp.Rules) == 0 {
		t.Fatal("expected non-empty rules for manual mode, got 0 rules")
	}
	if rp.Final != "direct" {
		t.Fatalf("expected Final == 'direct' for manual mode, got %q", rp.Final)
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
	// «all-direct» — старое имя; теперь это режим manual с final: direct.
	profiles := sampleRoutingProfiles()
	cfg := models.AppConfig{
		RoutingProfile:  "all-direct",
		RoutingProfiles: profiles,
	}

	rp, ok := cfg.ActiveRoutingProfile()
	if !ok {
		t.Fatal("expected ok == true for all-direct")
	}
	if rp.Name != "manual" {
		t.Fatalf("expected Name == 'manual', got %q", rp.Name)
	}
	if rp.Final != "direct" {
		t.Fatalf("expected Final == 'direct', got %q", rp.Final)
	}
	if len(rp.Rules) == 0 {
		t.Fatal("expected non-empty rules for manual mode")
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
