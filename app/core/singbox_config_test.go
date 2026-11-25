package core

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

func TestBuildRouteRules_IncludesDomainSuffix(t *testing.T) {
	cfg := models.AppConfig{
		RoutingProfile: "p1",
		RoutingProfiles: []models.RoutingProfile{
			{
				Name: "p1",
				Rules: []models.RouteRule{
					{
						Name:           "ru-tld",
						DomainSuffixes: []string{"ru", "xn--p1ai"},
						Action:         models.RouteDirect,
					},
				},
			},
		},
	}
	rules := buildRouteRules(cfg)
	if len(rules) < 1 {
		t.Fatalf("expected merged rules, got %d", len(rules))
	}
	var hit map[string]any
	for _, raw := range rules {
		m := raw.(map[string]any)
		ds, ok := m["domain_suffix"].([]string)
		if !ok || len(ds) != 2 || ds[0] != "ru" || ds[1] != "xn--p1ai" {
			continue
		}
		hit = m
		break
	}
	if hit == nil {
		t.Fatalf("expected merged rule with domain_suffix ru / xn--p1ai, got %+v", rules)
	}
	if _, ok := hit["domain_suffix"].([]string); !ok {
		t.Fatalf("expected domain_suffix, got %#v", hit["domain_suffix"])
	}
}

func TestBuildRouteRules_IncludesProcesses(t *testing.T) {
	cfg := models.AppConfig{
		RoutingProfile: "p1",
		RoutingProfiles: []models.RoutingProfile{
			{
				Name: "p1",
				Rules: []models.RouteRule{
					{
						Name:      "proc rule",
						Processes: []string{"chrome.exe", "msedge.exe"},
						Action:    models.RouteProxy,
					},
				},
			},
		},
	}

	rules := buildRouteRules(cfg)
	if len(rules) < 1 {
		t.Fatalf("expected merged rules, got %d", len(rules))
	}
	var hit map[string]any
	for _, raw := range rules {
		m := raw.(map[string]any)
		pn, ok := m["process_name"].([]string)
		if !ok || len(pn) < 2 {
			continue
		}
		if pn[0] == "chrome.exe" && pn[1] == "msedge.exe" {
			hit = m
			break
		}
	}
	if hit == nil {
		t.Fatalf("expected merged rule with chrome.exe/msedge.exe processes")
	}
	if got, _ := hit["outbound"].(string); got != "proxy" {
		t.Fatalf("expected outbound proxy, got %q", got)
	}
}

func TestRouteFinalOutbound_ByProxyMode(t *testing.T) {
	if got := routeFinalOutbound(models.AppConfig{ProxyMode: models.ProxyModeDirect}); got != "direct" {
		t.Fatalf("expected direct for ProxyModeDirect, got %q", got)
	}
	if got := routeFinalOutbound(models.AppConfig{ProxyMode: models.ProxyModeSystem}); got != "proxy" {
		t.Fatalf("expected proxy for ProxyModeSystem, got %q", got)
	}
	if got := routeFinalOutbound(models.AppConfig{ProxyMode: models.ProxyModeTun}); got != "proxy" {
		t.Fatalf("expected proxy for ProxyModeTun, got %q", got)
	}
}

func TestBuildRouteRules_DefaultProfileOrder(t *testing.T) {
	cfg := models.AppConfig{
		RoutingProfile: "default",
		UserRoutingRules: []models.RouteRule{
			{Name: "user-direct", DomainSuffixes: []string{"example.com"}, Action: models.RouteDirect},
		},
	}
	rules := buildRouteRules(cfg)
	if len(rules) < 4 {
		t.Fatalf("expected at least sniff, private-direct, process and user rules, got %d", len(rules))
	}
	// Первое правило — sniff, затем ip_is_private: true -> direct
	sniff := rules[0].(map[string]any)
	if sniff["action"] != "sniff" {
		t.Fatalf("expected first rule action=sniff, got %#v", rules[0])
	}
	priv := rules[1].(map[string]any)
	if got, _ := priv["ip_is_private"].(bool); !got {
		t.Fatalf("expected second rule ip_is_private, got %#v", rules[1])
	}
	// Ищем пользовательское правило: оно должно присутствовать и направлять в direct
	var foundUserRule bool
	for _, raw := range rules {
		m := raw.(map[string]any)
		if ds, ok := m["domain_suffix"].([]string); ok && len(ds) == 1 && ds[0] == "example.com" {
			if got, _ := m["outbound"].(string); got == "direct" {
				foundUserRule = true
				break
			}
		}
	}
	if !foundUserRule {
		t.Fatalf("expected user-direct rule in rules, got %#v", rules)
	}
}

func TestRouteFinalOutbound_FromProfileFinal(t *testing.T) {
	cfg := models.AppConfig{
		ProxyMode:      models.ProxyModeSystem,
		RoutingProfile: "selective",
		RoutingProfiles: []models.RoutingProfile{
			{
				Name:  "selective",
				Final: "direct",
				Rules: []models.RouteRule{{Name: "r", Domains: []string{"example.com"}, Action: models.RouteProxy}},
			},
		},
	}
	if got := routeFinalOutbound(cfg); got != "direct" {
		t.Fatalf("expected direct when profile.final is direct, got %q", got)
	}
	cfg.RoutingProfiles[0].Final = "proxy"
	if got := routeFinalOutbound(cfg); got != "proxy" {
		t.Fatalf("expected proxy when profile.final is proxy, got %q", got)
	}
}

// Профиль есть, но final пустой или неизвестный — решение принимает proxyMode.
// Этот случай читает UI-проверка маршрута, поэтому он должен быть зафиксирован.
func TestRouteFinalOutbound_EmptyOrUnknownFinalFallsBackToProxyMode(t *testing.T) {
	for _, final := range []string{"", "   ", "whatever"} {
		cfg := models.AppConfig{
			ProxyMode:      models.ProxyModeDirect,
			RoutingProfile: "p",
			RoutingProfiles: []models.RoutingProfile{
				{Name: "p", Final: final},
			},
		}
		if got := routeFinalOutbound(cfg); got != "direct" {
			t.Errorf("final=%q, proxyMode=direct_with_rules: got %q, want direct", final, got)
		}

		cfg.ProxyMode = models.ProxyModeSystem
		if got := routeFinalOutbound(cfg); got != "proxy" {
			t.Errorf("final=%q, proxyMode=system_proxy: got %q, want proxy", final, got)
		}
	}
}

func BenchmarkManagerApply_ConfigReuse(b *testing.B) {
	m := NewManager(NewAdapter(""))
	cfg := models.AppConfig{
		ProxyEnabled:   false,
		DPIEnabled:     false,
		ProxyMode:      models.ProxyModeSystem,
		ProxyPort:      8080,
		ActiveServerID: "s1",
		Servers: []models.Server{
			{
				ID:       "s1",
				Name:     "bench",
				Host:     "127.0.0.1",
				Port:     "443",
				UUID:     "123e4567-e89b-12d3-a456-426614174000",
				RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@127.0.0.1:443?security=tls#bench",
				Params:   map[string]string{"security": "tls"},
			},
		},
		RoutingProfile: "default",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Rules: []models.RouteRule{{Name: "catch-all", Action: models.RouteProxy}}},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Apply(cfg)
	}
}

func BenchmarkRenderSingBoxConfig_ServerSwitch(b *testing.B) {
	cfg := models.AppConfig{
		ProxyEnabled:         true,
		ProxyMode:            models.ProxyModeSystem,
		ProxyPort:            8080,
		LocalInboundPassword: "bench-secret",
		ActiveServerID:       "s1",
		Servers: []models.Server{
			{
				ID:       "s1",
				Name:     "node-a",
				Host:     "a.example.com",
				Port:     "443",
				UUID:     "123e4567-e89b-12d3-a456-426614174000",
				RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@a.example.com:443?security=tls#A",
				Params:   map[string]string{"security": "tls"},
			},
			{
				ID:       "s2",
				Name:     "node-b",
				Host:     "b.example.com",
				Port:     "443",
				UUID:     "223e4567-e89b-12d3-a456-426614174000",
				RawVLESS: "vless://223e4567-e89b-12d3-a456-426614174000@b.example.com:443?security=tls#B",
				Params:   map[string]string{"security": "tls"},
			},
		},
		RoutingProfile: "default",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Rules: []models.RouteRule{{Name: "catch-all", Action: models.RouteProxy}}},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			cfg.ActiveServerID = "s1"
		} else {
			cfg.ActiveServerID = "s2"
		}
		if _, err := RenderSingBoxConfig(cfg, b.TempDir()); err != nil {
			b.Fatal(err)
		}
	}
}

func TestRenderSingBoxConfig_TunAddsInboundAndClashAPI(t *testing.T) {
	cfg := models.AppConfig{
		ProxyEnabled:         true,
		ProxyMode:            models.ProxyModeTun,
		ProxyPort:            8080,
		LocalInboundPassword: "tun-test-secret",
		ActiveServerID:       "s1",
		Servers: []models.Server{
			{
				ID:       "s1",
				Name:     "n",
				Host:     "127.0.0.1",
				Port:     "443",
				UUID:     "123e4567-e89b-12d3-a456-426614174000",
				RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@127.0.0.1:443?security=tls&sni=example.com#n",
				Params:   map[string]string{"security": "tls", "sni": "example.com"},
			},
		},
		RoutingProfile: "default",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Rules: []models.RouteRule{{Name: "catch-all", Action: models.RouteProxy}}},
		},
	}
	raw, err := RenderSingBoxConfig(cfg, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	inbounds, _ := doc["inbounds"].([]any)
	if len(inbounds) != 2 {
		t.Fatalf("expected 2 inbounds (mixed+tun), got %d", len(inbounds))
	}
	tun := inbounds[1].(map[string]any)
	if tun["type"] != "tun" || tun["interface_name"] != "pulse-tun" {
		t.Fatalf("unexpected tun inbound: %#v", tun)
	}
	exp, _ := doc["experimental"].(map[string]any)
	api, _ := exp["clash_api"].(map[string]any)
	if api["external_controller"] != "127.0.0.1:20991" {
		t.Fatalf("unexpected clash_api: %#v", api)
	}
	if api["secret"] != "tun-test-secret" {
		t.Fatalf("expected clash_api.secret, got %#v", api["secret"])
	}
	if _, ok := api["external_ui"]; ok {
		t.Fatalf("clash_api.external_ui must be omitted (avoids GitHub download), got %#v", api["external_ui"])
	}
	if _, ok := api["store_selected"]; ok {
		t.Fatalf("clash_api must not contain store_selected, got %#v", api["store_selected"])
	}
	if _, ok := exp["store_selected"]; ok {
		t.Fatalf("experimental must not contain store_selected (выбор сервера — AppConfig), got %#v", exp["store_selected"])
	}
	cache, _ := exp["cache_file"].(map[string]any)
	if cache["enabled"] != true {
		t.Fatalf("unexpected cache_file.enabled: %#v", cache)
	}
	if _, ok := cache["store_selected"]; ok {
		t.Fatalf("cache_file must not contain store_selected, got %#v", cache["store_selected"])
	}
	if cache["path"] != "cache.db" {
		t.Fatalf("expected cache_file.path cache.db, got %#v", cache["path"])
	}
	if _, ok := doc["dns"].(map[string]any); !ok {
		t.Fatalf("expected dns section")
	}
	route, _ := doc["route"].(map[string]any)
	if route["auto_detect_interface"] != true {
		t.Fatalf("expected route.auto_detect_interface for tun, got %#v", route["auto_detect_interface"])
	}
	if route["default_domain_resolver"] != SingBoxPrimaryDNSTag {
		t.Fatalf("expected route.default_domain_resolver %q, got %#v", SingBoxPrimaryDNSTag, route["default_domain_resolver"])
	}
	outbounds, _ := doc["outbounds"].([]any)
	proxyOB := outbounds[0].(map[string]any)
	// VLESS server hostname must resolve via the DIRECT bootstrap resolver, not the
	// DoH-via-proxy resolver — otherwise DNS loops through the proxy being established.
	if proxyOB["domain_resolver"] != SingBoxBootstrapDNSTag {
		t.Fatalf("expected proxy outbound domain_resolver %q, got %#v", SingBoxBootstrapDNSTag, proxyOB["domain_resolver"])
	}
	mixed := inbounds[0].(map[string]any)
	if _, ok := mixed["users"]; ok {
		t.Fatalf("mixed inbound must NOT have users (local 127.0.0.1, browsers via system proxy send no auth), got %#v", mixed["users"])
	}
}

func TestRenderSingBoxConfig_EmitsClashAPIWhenLocalPasswordEmpty(t *testing.T) {
	cfg := models.AppConfig{
		ProxyEnabled:           true,
		ProxyMode:              models.ProxyModeSystem,
		LocalInboundPassword:   "",
		LocalInboundUser:       "",
		ActiveServerID:         "s1",
		Servers: []models.Server{
			{
				ID:       "s1",
				Name:     "n",
				Host:     "a.example.com",
				Port:     "443",
				UUID:     "123e4567-e89b-12d3-a456-426614174000",
				RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@a.example.com:443?security=tls#n",
				Params:   map[string]string{"security": "tls"},
			},
		},
		RoutingProfile: "default",
		RoutingProfiles: []models.RoutingProfile{
			{Name: "default", Rules: []models.RouteRule{{Name: "catch-all", Action: models.RouteProxy}}},
		},
	}
	runDir := t.TempDir()
	raw, err := RenderSingBoxConfig(cfg, runDir)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	exp, _ := doc["experimental"].(map[string]any)
	api, _ := exp["clash_api"].(map[string]any)
	if api["external_controller"] != "127.0.0.1:20991" {
		t.Fatalf("unexpected clash_api: %#v", api)
	}
	if api["secret"] != "" {
		t.Fatalf("expected empty clash_api.secret in this fixture, got %#v", api["secret"])
	}
	if _, ok := api["external_ui"]; ok {
		t.Fatalf("clash_api.external_ui must be omitted, got %#v", api["external_ui"])
	}
	if _, ok := api["store_selected"]; ok {
		t.Fatalf("clash_api must not contain store_selected, got %#v", api["store_selected"])
	}
	if _, ok := exp["store_selected"]; ok {
		t.Fatalf("experimental must not contain store_selected, got %#v", exp["store_selected"])
	}
	cache, _ := exp["cache_file"].(map[string]any)
	if cache["enabled"] != true {
		t.Fatalf("unexpected cache_file.enabled: %#v", cache)
	}
	if _, ok := cache["store_selected"]; ok {
		t.Fatalf("cache_file must not contain store_selected, got %#v", cache["store_selected"])
	}
	if cache["path"] != "cache.db" {
		t.Fatalf("expected cache_file.path cache.db, got %#v", cache["path"])
	}
	mixed := doc["inbounds"].([]any)[0].(map[string]any)
	if _, ok := mixed["users"]; ok {
		t.Fatalf("mixed inbound must NOT have users, got %#v", mixed["users"])
	}
}

func TestBuildResilientOutbounds_NoUrltestHijack(t *testing.T) {
	cfg := models.AppConfig{
		ProxyEnabled:   true,
		ProxyMode:      models.ProxyModeSystem,
		ActiveServerID: "nl-selected",
		Servers: []models.Server{
			{
				ID:                 "ru-fast",
				Name:               "Россия Москва",
				Host:               "ru1.example.com",
				Port:               "443",
				UUID:               "11111111-2222-3333-4444-555555555555",
				Params:             map[string]string{"security": "tls", "sni": "ru1.example.com"},
				LastCheckOK:        true,
				LastCheckLatencyMs: 12,
			},
			{
				ID:                 "nl-selected",
				Name:               "Нидерланды",
				Host:               "nl1.example.com",
				Port:               "443",
				UUID:               "22222222-2222-3333-4444-555555555555",
				Params:             map[string]string{"security": "tls", "sni": "nl1.example.com"},
				LastCheckOK:        true,
				LastCheckLatencyMs: 85,
			},
		},
	}

	outbounds, err := buildResilientOutbounds(cfg, cfg.Servers[1])
	if err != nil {
		t.Fatalf("buildResilientOutbounds: %v", err)
	}
	for _, raw := range outbounds {
		m := raw.(map[string]any)
		if m["type"] == "urltest" {
			t.Fatalf("urltest group must not be created (causes silent hijack to RU node): %#v", m)
		}
	}
	proxyOut := outbounds[0].(map[string]any)
	if proxyOut["tag"] != "proxy" || proxyOut["server"] != "nl1.example.com" {
		t.Fatalf("expected proxy outbound strictly pinned to nl1.example.com, got %#v", proxyOut)
	}
}

func TestRenderSingBoxConfig_AllModesCheckWithBinary(t *testing.T) {
	singBox := filepath.Join("..", "..", "third_party", "core", "sing-box.exe")
	_, statErr := os.Stat(singBox)
	haveBinary := statErr == nil

	base := config.DefaultConfig()
	base.ProxyEnabled = true
	base.ActiveServerID = "nl-node"
	base.Servers = []models.Server{
		{
			ID:       "nl-node",
			Name:     "NL",
			Host:     "nl.example.com",
			Port:     "443",
			UUID:     "123e4567-e89b-12d3-a456-426614174000",
			RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@nl.example.com:443?security=tls&sni=nl.example.com#NL",
			Params:   map[string]string{"security": "tls", "sni": "nl.example.com"},
		},
	}
	base.UserRoutingRules = []models.RouteRule{
		{Name: "user:*.example.net", Domains: []string{"*.example.net"}, Action: models.RouteProxy},
		{Name: "user:chatgpt.com", DomainSuffixes: []string{"chatgpt.com"}, Action: models.RouteProxy},
	}

	for _, proxyMode := range []models.ProxyMode{models.ProxyModeSystem, models.ProxyModeTun} {
		for _, routingMode := range []string{config.RoutingModeGlobal, config.RoutingModeFullProxy, config.RoutingModeManual} {
			name := string(proxyMode) + "_" + routingMode
			t.Run(name, func(t *testing.T) {
				cfg := base
				cfg.ProxyMode = proxyMode
				cfg.RoutingProfile = routingMode

				raw, err := RenderSingBoxConfig(cfg, t.TempDir())
				if err != nil {
					t.Fatalf("RenderSingBoxConfig: %v", err)
				}

				var doc map[string]any
				if err := json.Unmarshal(raw, &doc); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}

				route := doc["route"].(map[string]any)
				rules := route["rules"].([]any)
				if rules[0].(map[string]any)["action"] != "sniff" {
					t.Fatalf("expected sniff rule first, got %#v", rules[0])
				}

				hasHijackDNS := false
				hasWildcardDomainBug := false
				hasConvertedSuffix := false
				for _, rRaw := range rules {
					rMap := rRaw.(map[string]any)
					if rMap["action"] == "hijack-dns" {
						hasHijackDNS = true
					}
					if doms, ok := rMap["domain"].([]any); ok {
						for _, d := range doms {
							if s, _ := d.(string); s == "*.example.net" || s == "*.local" {
								hasWildcardDomainBug = true
							}
						}
					}
					if sufs, ok := rMap["domain_suffix"].([]any); ok {
						for _, s := range sufs {
							if s == "example.net" {
								hasConvertedSuffix = true
							}
						}
					}
				}
				if hasWildcardDomainBug {
					t.Fatalf("wildcard pattern leaked into domain instead of domain_suffix")
				}
				if !hasConvertedSuffix {
					t.Fatalf("expected *.example.net converted to domain_suffix example.net")
				}

				if proxyMode == models.ProxyModeTun {
					if !hasHijackDNS {
						t.Fatalf("TUN mode must include hijack-dns rule")
					}
					tun := doc["inbounds"].([]any)[1].(map[string]any)
					if tun["stack"] != "mixed" {
						t.Fatalf("TUN stack = %v, want mixed", tun["stack"])
					}
					if tun["strict_route"] != false {
						t.Fatalf("TUN strict_route = %v, want false", tun["strict_route"])
					}
					addrs, _ := tun["address"].([]any)
					if len(addrs) < 1 || addrs[0] != "172.19.0.1/30" {
						t.Fatalf("TUN must specify IPv4 address 172.19.0.1/30, got %#v", tun["address"])
					}
				}

				dnsCfg := doc["dns"].(map[string]any)
				_, hasDNSRules := dnsCfg["rules"]
				if routingMode == config.RoutingModeGlobal && hasDNSRules {
					t.Fatalf("global mode must send all DNS through DoH proxy without .ru direct DNS rules")
				}
				if routingMode != config.RoutingModeGlobal && !hasDNSRules {
					t.Fatalf("%s mode must keep .ru direct DNS rules", routingMode)
				}

				if haveBinary {
					path := filepath.Join(t.TempDir(), "sing-box.json")
					if err := os.WriteFile(path, raw, 0o644); err != nil {
						t.Fatalf("write config: %v", err)
					}
					if out, err := exec.Command(singBox, "check", "-c", path).CombinedOutput(); err != nil {
						t.Fatalf("sing-box check failed: %v\n%s\n%s", err, out, raw)
					}
				}
			})
		}
	}
}
