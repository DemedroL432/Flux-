package config

import "pulseshield/app/models"

// builtinRoutingPresets — встроенные профили маршрутизации.
// Классификация доменов: см. service_routing.go (ProxyOnly vs DPI-bypass).
func builtinRoutingPresets() []models.RoutingProfile {
	localDomains := []string{"*.local", "localhost", "*.lan"}

	// Редкие случаи, когда нужен именно выход через VLESS — допишите домены здесь.
	var vpnPreferredDomains []string

	localDirect := []models.RouteRule{
		{Name: "local-direct", Domains: localDomains, Action: models.RouteDirect},
	}
	lanCIDR := models.RouteRule{
		Name:   "private-lan-direct",
		CIDRs:  []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8"},
		Action: models.RouteDirect,
	}
	ruDirect := models.RouteRule{
		Name:           "ru-tld-direct",
		DomainSuffixes: RuDirectDomainSuffixes(),
		Action:         models.RouteDirect,
	}
	ruServicesDirect := models.RouteRule{
		Name:    "ru-services-direct",
		Domains: RuDirectDomains(),
		Action:  models.RouteDirect,
	}
	gamingProcesses := []models.RouteRule{
		{
			Name: "launchers-direct",
			Processes: []string{
				"EpicGamesLauncher.exe", "RiotClientServices.exe",
				"VALORANT.exe", "LeagueClient.exe", "Battle.net.exe", "Battle.net Launcher.exe", "cs2.exe", "dota2.exe",
			},
			Action: models.RouteDirect,
		},
	}

	// Process-based proxy rules — до любых direct/xbox правил.
	procBeforeXbox := ProcessProxyRules(nil)
	gamingExtra := append([]models.RouteRule{}, procBeforeXbox...)
	gamingExtra = append(gamingExtra, XboxDnsDirectRule())
	gamingExtra = append(gamingExtra, gamingProcesses...)
	gamingExtra = append(gamingExtra, models.RouteRule{
		Name:           "media-social-via-proxy",
		DomainSuffixes: SelectiveMediaDomainSuffixes(),
		Action:         models.RouteProxy,
	})

	rfBase := append(append([]models.RouteRule{}, localDirect...), lanCIDR, ruDirect, ruServicesDirect)
	rfBase = append(rfBase, procBeforeXbox...)
	rfBase = append(rfBase, XboxDnsDirectRule())
	rfSmartRules := append(rfBase, append(gamingProcesses,
		models.RouteRule{
			Name:           "proxy-only-messengers",
			DomainSuffixes: ProxyOnlyMediaDomainSuffixes(),
			Action:         models.RouteProxy,
		},
		models.RouteRule{
			Name:           "dpi-direct-services",
			DomainSuffixes: DPIBypassMediaDomainSuffixes(),
			Action:         models.RouteDPIBypass,
		},
	)...)
	if len(vpnPreferredDomains) > 0 {
		rfSmartRules = append(rfSmartRules, models.RouteRule{
			Name:    "vpn-preferred",
			Domains: vpnPreferredDomains,
			Action:  models.RouteProxy,
		})
	}

	return []models.RoutingProfile{
		{
			Name: "default",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost"}, Action: models.RouteDirect},
				{Name: "ru-tld-direct", DomainSuffixes: RuDirectDomainSuffixes(), Action: models.RouteDirect},
				{Name: "blocked-example", Domains: []string{"ads.example.com"}, Action: models.RouteBlock},
				{Name: "default-proxy", Action: models.RouteProxy},
			},
		},
		{
			Name:  "selective-proxy",
			Final: "direct",
			Rules: func() []models.RouteRule {
				rules := append([]models.RouteRule{}, localDirect...)
				rules = append(rules, ruDirect, ruServicesDirect)
				rules = append(rules, ProcessProxyRules(nil)...)
				rules = append(rules, XboxDnsDirectRule())
				rules = append(rules, models.RouteRule{
					Name:           "media-social-via-proxy",
					DomainSuffixes: SelectiveMediaDomainSuffixes(),
					Action:         models.RouteProxy,
				})
				return rules
			}(),
		},
		{
			Name:  "gaming-friendly",
			Final: "direct",
			Rules: append(append(append([]models.RouteRule{}, localDirect...), ruDirect, ruServicesDirect), gamingExtra...),
		},
		{
			Name:  "rf-smart-split",
			Final: "direct",
			Rules: rfSmartRules,
		},
		{
			Name:  "all-direct",
			Final: "direct",
			Rules: []models.RouteRule{
				{Name: "local-direct", Domains: localDomains, Action: models.RouteDirect},
				lanCIDR,
				ruDirect,
				ruServicesDirect,
				{
					Name:           "all-direct-services",
					DomainSuffixes: append(append(ProxyOnlyMediaDomainSuffixes(), DPIBypassMediaDomainSuffixes()...), SelectiveMediaDomainSuffixes()...),
					Action:         models.RouteDirect,
				},
				{
					Name:           "bypass-proxy-media",
					DomainSuffixes: []string{"netflix.com", "nflxvideo.net", "nflxso.net", "cdninstagram.com", "perplexity.ai", "mistral.ai", "replicate.com", "huggingface.co"},
					Action:         models.RouteDirect,
				},
			},
		},
	}
}

// dedupeRoutingProfiles оставляет первый профиль с данным именем (чинит дубликаты в сохранённом JSON).
func dedupeRoutingProfiles(cfg *models.AppConfig) bool {
	if cfg == nil || len(cfg.RoutingProfiles) == 0 {
		return false
	}
	before := len(cfg.RoutingProfiles)
	seen := make(map[string]struct{}, len(cfg.RoutingProfiles))
	out := cfg.RoutingProfiles[:0]
	for _, p := range cfg.RoutingProfiles {
		name := p.Name
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, p)
	}
	cfg.RoutingProfiles = out
	return len(cfg.RoutingProfiles) < before
}

// MergeBuiltinRoutingProfiles добавляет встроенные профили, если профиля с таким именем ещё нет.
// Возвращает true, если список профилей изменился (дубликаты убраны или добавлены встроенные).
func MergeBuiltinRoutingProfiles(cfg *models.AppConfig) bool {
	if cfg == nil {
		return false
	}
	changed := dedupeRoutingProfiles(cfg)
	have := make(map[string]struct{}, len(cfg.RoutingProfiles))
	for _, p := range cfg.RoutingProfiles {
		have[p.Name] = struct{}{}
	}
	for _, p := range builtinRoutingPresets() {
		if _, ok := have[p.Name]; !ok {
			cfg.RoutingProfiles = append(cfg.RoutingProfiles, p)
			have[p.Name] = struct{}{}
			changed = true
		}
	}
	return changed
}
