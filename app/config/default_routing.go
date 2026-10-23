package config

import (
	"path/filepath"

	"pulseshield/app/models"
)

// MergeSingBoxRoutingRules: процессы → пользователь → активный профиль → [опционально] библиотека умного сплита.
// При StrictRoutingMode библиотека не добавляется (режим «чистый гейминг»).
func MergeSingBoxRoutingRules(cfg models.AppConfig, zapretListsDir string) []models.RouteRule {
	var merged []models.RouteRule
	merged = append(merged, ProcessProxyRules(&cfg)...)
	merged = append(merged, cfg.UserRoutingRules...)
	if rp, ok := cfg.ActiveRoutingProfile(); ok {
		merged = append(merged, rp.Rules...)
	}
	if !cfg.StrictRoutingMode {
		merged = append(merged, DefaultRoutingLibraryRules(zapretListsDir)...)
	}
	return merged
}

// DefaultRoutingLibraryRules — домены/IP «умного сплита» (последний приоритет среди явных правил).
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
				Name:    "builtin-proxy-cloudflare-ipset-file",
				CIDRs:   cidrs,
				Action:  models.RouteProxy,
			})
		}
		if cidrs, err := ReadZapretListCIDRs(filepath.Join(zapretListsDir, "ipset-steam.txt")); err == nil && len(cidrs) > 0 {
			rules = append(rules, models.RouteRule{
				Name:    "builtin-proxy-steam-ipset",
				CIDRs:   cidrs,
				Action:  models.RouteProxy,
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
	// Только домены, которых НЕТ в XboxDnsFriendlyDomainSuffixes().
	// Домены из Xbox-списка (openai.com, anthropic.com, claude.ai, cursor.*, jetbrains.com,
	// gemini.google.com и др.) убраны: при профиле rf-smart-split они идут через XboxDnsDirectRule
	// (dpi-bypass/direct), которое стоит первым и выигрывает в first-match.
	return []string{
		"netflix.com",
		"nflxvideo.net",
		"nflxso.net",
		"cdninstagram.com", // только CDN; instagram.com — в Xbox-списке
		"perplexity.ai",
		"mistral.ai",
		"replicate.com",
		"huggingface.co",
	}
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
