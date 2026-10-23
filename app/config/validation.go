package config

import (
	"fmt"
	"strings"

	"pulseshield/app/models"
)

func ValidateAppConfig(cfg models.AppConfig) error {
	if cfg.ProxyPort != 0 && (cfg.ProxyPort < 1 || cfg.ProxyPort > 65535) {
		return fmt.Errorf("proxyPort must be between 1 and 65535")
	}
	if err := validateRouting(cfg.RoutingProfiles); err != nil {
		return err
	}
	return nil
}

func validateRouting(profiles []models.RoutingProfile) error {
	for _, p := range profiles {
		switch strings.ToLower(strings.TrimSpace(p.Final)) {
		case "", "direct", "proxy":
		default:
			return fmt.Errorf("routing profile %q: final must be empty, direct or proxy", p.Name)
		}
		if len(p.Rules) == 0 {
			return fmt.Errorf("routing profile %q has no rules", p.Name)
		}
		defaultProxyOrDirect := 0
		for _, r := range p.Rules {
			if len(r.Domains) == 0 && len(r.DomainSuffixes) == 0 && len(r.CIDRs) == 0 && len(r.Processes) == 0 {
				if r.Action == models.RouteProxy || r.Action == models.RouteDirect {
					defaultProxyOrDirect++
				}
			}
		}
		if defaultProxyOrDirect > 1 {
			return fmt.Errorf("routing profile %q has multiple catch-all rules", p.Name)
		}
	}
	return nil
}
