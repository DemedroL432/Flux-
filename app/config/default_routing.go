package config

import (
	"path/filepath"

	"pulseshield/app/models"
)

// EffectiveProfileRules возвращает правила активного профиля с учётом состояния DPI:
// правило dpi-bypass-media (YouTube, Discord и др.) при включённом DPI идёт в обход прокси (winws2),
// а при выключенном DPI направляется через прокси, чтобы сервисы оставались доступны.
func EffectiveProfileRules(cfg models.AppConfig) []models.RouteRule {
	rp, ok := cfg.ActiveRoutingProfile()
	if !ok {
		return nil
	}
	out := make([]models.RouteRule, 0, len(rp.Rules))
	for _, r := range rp.Rules {
		if r.Name == "dpi-bypass-media" && r.Action == models.RouteDPIBypass && !cfg.DPIEnabled {
			r.Action = models.RouteProxy
		}
		out = append(out, r)
	}
	return out
}

// MergeSingBoxRoutingRules: для встроенных режимов (manual, full-proxy, global) применяются строго
// пользовательские правила (UserRoutingRules) и правила активного профиля — без навязанных процессов и библиотек.
func MergeSingBoxRoutingRules(cfg models.AppConfig, zapretListsDir string) []models.RouteRule {
	resolved := models.ResolveRoutingProfileAlias(cfg.RoutingProfile)
	isBuiltinMode := resolved == models.RoutingModeManual || resolved == models.RoutingModeFullProxy || resolved == models.RoutingModeGlobal

	var merged []models.RouteRule
	merged = append(merged, cfg.UserRoutingRules...)
	if !isBuiltinMode {
		merged = append(merged, ProcessProxyRules(&cfg)...)
	}
	merged = append(merged, EffectiveProfileRules(cfg)...)
	if !cfg.StrictRoutingMode && !isBuiltinMode {
		merged = append(merged, DefaultRoutingLibraryRules(zapretListsDir)...)
	}
	return merged
}

// DefaultRoutingLibraryRules — домены/IP «умного сплита» (используется только для внешних пользовательских профилей без StrictRoutingMode).
func DefaultRoutingLibraryRules(zapretListsDir string) []models.RouteRule {
	var rules []models.RouteRule

	rules = append(rules, models.RouteRule{
		Name:           "builtin-proxy-ai-media",
		DomainSuffixes: builtinProxyDomainSuffixes(),
		Action:         models.RouteProxy,
	})

	if zapretListsDir != "" {
		if cidrs, err := ReadZapretListCIDRs(filepath.Join(zapretListsDir, "ipset-cloudflare.txt")); err == nil && len(cidrs) > 0 {
			rules = append(rules, models.RouteRule{
				Name:   "builtin-proxy-cloudflare-ipset-file",
				CIDRs:  cidrs,
				Action: models.RouteProxy,
			})
		}
		for _, pair := range []struct {
			name string
			file string
		}{
			{"builtin-direct-riot-ipset", "ipset-riot-games.txt"},
		} {
			if cidrs, err := ReadZapretListCIDRs(filepath.Join(zapretListsDir, pair.file)); err == nil && len(cidrs) > 0 {
				rules = append(rules, models.RouteRule{
					Name:    pair.name,
					CIDRs:   cidrs,
					Action:  models.RouteDirect,
				})
			}
		}
	}

	if len(builtinCloudflareCIDRs()) > 0 {
		rules = append(rules, models.RouteRule{
			Name:    "builtin-proxy-cloudflare-ipv4-core",
			CIDRs:   builtinCloudflareCIDRs(),
			Action:  models.RouteProxy,
		})
	}

	rules = append(rules, models.RouteRule{
		Name:           "builtin-direct-anticheat",
		DomainSuffixes: builtinAnticheatDomainSuffixes(),
		Action:         models.RouteDirect,
	})

	return rules
}

func builtinProxyDomainSuffixes() []string {
	return mergeUniqueStringSlices(
		AIServiceDomainSuffixes(),
		[]string{
			"netflix.com",
			"nflxvideo.net",
			"nflxso.net",
			"cdninstagram.com",
		},
	)
}

func builtinAnticheatDomainSuffixes() []string {
	return []string{
		"easyanticheat.net",
		"easy.ac",
		"battleye.com",
		"faceit.com",
	}
}

// Усечённый публичный список IPv4 Cloudflare (дополнение к ipset-cloudflare.txt из Zapret).
func builtinCloudflareCIDRs() []string {
	return []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}
}
