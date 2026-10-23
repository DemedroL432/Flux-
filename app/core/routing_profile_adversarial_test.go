package core

import (
	"encoding/json"
	"testing"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

func TestSingBoxConfig_ProfileAliases_EndToEnd(t *testing.T) {
	defaultCfg := config.DefaultConfig()
	defaultCfg.ProxyEnabled = true
	defaultCfg.ActiveServerID = "srv-test"
	defaultCfg.Servers = []models.Server{
		{
			ID:       "srv-test",
			Name:     "Test Node",
			Host:     "vpn.example.com",
			Port:     "443",
			UUID:     "123e4567-e89b-12d3-a456-426614174000",
			RawVLESS: "vless://123e4567-e89b-12d3-a456-426614174000@vpn.example.com:443?security=tls&sni=vpn.example.com#node",
			Params:   map[string]string{"security": "tls", "sni": "vpn.example.com"},
		},
	}

	testCases := []struct {
		name              string
		routingProfile    string
		expectedFinal     string
		expectedMinRules  int
		mustHaveDirectOut bool
	}{
		{
			name:              "profile bypass-russia produces direct final and rich rules",
			routingProfile:    "bypass-russia",
			expectedFinal:     "direct",
			expectedMinRules:  5,
			mustHaveDirectOut: true,
		},
		{
			name:              "profile all-direct produces direct final",
			routingProfile:    "all-direct",
			expectedFinal:     "direct",
			expectedMinRules:  4,
			mustHaveDirectOut: true,
		},
		{
			name:              "profile all-proxy produces proxy final under system proxy",
			routingProfile:    "all-proxy",
			expectedFinal:     "proxy",
			expectedMinRules:  2,
			mustHaveDirectOut: false,
		},
		{
			name:              "profile rf-smart-split produces direct final and rich rules",
			routingProfile:    "rf-smart-split",
			expectedFinal:     "direct",
			expectedMinRules:  5,
			mustHaveDirectOut: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultCfg
			cfg.RoutingProfile = tc.routingProfile

			// 1. Verify routeFinalOutbound
			finalOut := routeFinalOutbound(cfg)
			if finalOut != tc.expectedFinal {
				t.Fatalf("routeFinalOutbound(%q) = %q, expected %q", tc.routingProfile, finalOut, tc.expectedFinal)
			}

			// 2. Render sing-box config
			raw, err := RenderSingBoxConfig(cfg, t.TempDir())
			if err != nil {
				t.Fatalf("RenderSingBoxConfig failed: %v", err)
			}

			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("invalid json produced: %v", err)
			}

			route, ok := doc["route"].(map[string]any)
			if !ok {
				t.Fatal("missing route section in sing-box config")
			}

			if route["final"] != tc.expectedFinal {
				t.Fatalf("route.final = %v, expected %q", route["final"], tc.expectedFinal)
			}

			rules, ok := route["rules"].([]any)
			if !ok {
				t.Fatal("missing route.rules slice")
			}

			if len(rules) < tc.expectedMinRules {
				t.Fatalf("expected at least %d rules, got %d", tc.expectedMinRules, len(rules))
			}

			if tc.mustHaveDirectOut {
				var hasDirect bool
				for _, r := range rules {
					m, ok := r.(map[string]any)
					if ok && m["outbound"] == "direct" {
						hasDirect = true
						break
					}
				}
				if !hasDirect {
					t.Fatalf("expected at least one rule with outbound: direct for profile %q", tc.routingProfile)
				}
			}
		})
	}
}
