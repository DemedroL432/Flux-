package config

import (
	"testing"

	"pulseshield/app/models"
)

func TestAddUserRoutingRule_Normalization(t *testing.T) {
	var cfg models.AppConfig

	if err := AddUserRoutingRule(&cfg, "*.Example.ORG", "proxy"); err != nil {
		t.Fatalf("wildcard rule: %v", err)
	}
	if err := AddUserRoutingRule(&cfg, "https://ChatGPT.com:443/c/123?foo=bar", "proxy"); err != nil {
		t.Fatalf("url rule: %v", err)
	}
	if err := AddUserRoutingRule(&cfg, "Discord.exe", "direct"); err != nil {
		t.Fatalf("process rule: %v", err)
	}
	if err := AddUserRoutingRule(&cfg, "full:Exact.Host.COM", "block"); err != nil {
		t.Fatalf("exact rule: %v", err)
	}
	if err := AddUserRoutingRule(&cfg, "1.2.3.4", "proxy"); err != nil {
		t.Fatalf("ip rule: %v", err)
	}
	if err := AddUserRoutingRule(&cfg, "10.0.0.0/8", "direct"); err != nil {
		t.Fatalf("cidr rule: %v", err)
	}

	if len(cfg.UserRoutingRules) != 6 {
		t.Fatalf("expected 6 rules, got %d", len(cfg.UserRoutingRules))
	}

	r0 := cfg.UserRoutingRules[0]
	if len(r0.Domains) != 0 || len(r0.DomainSuffixes) != 1 || r0.DomainSuffixes[0] != "example.org" {
		t.Fatalf("*.Example.ORG must normalize to DomainSuffixes=[example.org], got domains=%v suffixes=%v", r0.Domains, r0.DomainSuffixes)
	}

	r1 := cfg.UserRoutingRules[1]
	if len(r1.DomainSuffixes) != 1 || r1.DomainSuffixes[0] != "chatgpt.com" {
		t.Fatalf("URL must normalize to chatgpt.com suffix, got %+v", r1)
	}

	r2 := cfg.UserRoutingRules[2]
	if len(r2.Processes) != 1 || r2.Processes[0] != "discord.exe" || len(r2.DomainSuffixes) != 0 {
		t.Fatalf("Discord.exe must go to Processes=[discord.exe], got %+v", r2)
	}

	r3 := cfg.UserRoutingRules[3]
	if len(r3.Domains) != 1 || r3.Domains[0] != "exact.host.com" || len(r3.DomainSuffixes) != 0 {
		t.Fatalf("full: prefix must populate exact Domains, got %+v", r3)
	}

	r4 := cfg.UserRoutingRules[4]
	if len(r4.CIDRs) != 1 || r4.CIDRs[0] != "1.2.3.4/32" {
		t.Fatalf("single IP must normalize to /32 CIDR, got %+v", r4)
	}

	r5 := cfg.UserRoutingRules[5]
	if len(r5.CIDRs) != 1 || r5.CIDRs[0] != "10.0.0.0/8" {
		t.Fatalf("CIDR mismatch: %+v", r5)
	}
}

func TestNormalizeUserRoutingRules_MigratesBrokenWildcards(t *testing.T) {
	cfg := models.AppConfig{
		UserRoutingRules: []models.RouteRule{
			{
				Name:    "user:*.openai.com",
				Domains: []string{"*.openai.com", "https://claude.ai/chats", "Telegram.exe"},
				Action:  models.RouteProxy,
			},
			{
				Name:           "user:discord.exe",
				DomainSuffixes: []string{"Discord.exe", "https://example.net:8443/path"},
				Action:         models.RouteDirect,
			},
		},
	}

	if !NormalizeUserRoutingRules(&cfg) {
		t.Fatal("expected NormalizeUserRoutingRules to report changes")
	}

	r0 := cfg.UserRoutingRules[0]
	if len(r0.Domains) != 0 {
		t.Fatalf("expected wildcard/url/exe removed from Domains, got %v", r0.Domains)
	}
	if len(r0.DomainSuffixes) != 2 || r0.DomainSuffixes[0] != "openai.com" || r0.DomainSuffixes[1] != "claude.ai" {
		t.Fatalf("expected migrated DomainSuffixes [openai.com claude.ai], got %v", r0.DomainSuffixes)
	}
	if len(r0.Processes) != 1 || r0.Processes[0] != "telegram.exe" {
		t.Fatalf("expected migrated Process telegram.exe, got %v", r0.Processes)
	}

	r1 := cfg.UserRoutingRules[1]
	if len(r1.Processes) != 1 || r1.Processes[0] != "discord.exe" {
		t.Fatalf("expected Discord.exe moved to Processes, got %+v", r1)
	}
	if len(r1.DomainSuffixes) != 1 || r1.DomainSuffixes[0] != "example.net" {
		t.Fatalf("expected cleaned DomainSuffixes [example.net], got %v", r1.DomainSuffixes)
	}

	// Повторный вызов идемпотентен.
	if NormalizeUserRoutingRules(&cfg) {
		t.Fatal("second NormalizeUserRoutingRules call must be a no-op")
	}
}

func TestMergeSingBoxRoutingRules_PrecedenceAndCleanBuiltins(t *testing.T) {
	cfg := DefaultConfig()
	cfg.UserRoutingRules = []models.RouteRule{
		{Name: "user-first", Processes: []string{"discord.exe"}, Action: models.RouteDirect},
	}

	// Ни один встроенный режим не подмешивает навязанные process-proxy-map или DefaultRoutingLibraryRules.
	for _, mode := range []string{RoutingModeManual, RoutingModeFullProxy, RoutingModeGlobal} {
		cfg.RoutingProfile = mode
		merged := MergeSingBoxRoutingRules(cfg, "")
		if len(merged) == 0 || merged[0].Name != "user-first" {
			t.Fatalf("%s: UserRoutingRules must be first, got %+v", mode, merged)
		}
		for _, r := range merged {
			if r.Name == "process-proxy-map" || r.Name == "builtin-proxy-ai-media" {
				t.Fatalf("%s: merged rules must not include hidden rule %q", mode, r.Name)
			}
			if mode == RoutingModeGlobal && r.Name == "ru-tld-direct" {
				t.Fatalf("global mode must not include ru-tld-direct")
			}
		}
	}
}
