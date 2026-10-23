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

func TestBuiltinPresetsRfSmartSplitOrder(t *testing.T) {
	for _, p := range builtinRoutingPresets() {
		if p.Name != "rf-smart-split" {
			continue
		}
		var hasXbox, hasProxyList, hasDPI, hasRuTLD, hasRuServices, hasProcMap bool
		for _, r := range p.Rules {
			if r.Name == "xbox-dns-direct" && r.Action == models.RouteDPIBypass && len(r.DomainSuffixes) > 0 {
				hasXbox = true
			}
			if r.Name == "ru-tld-direct" && r.Action == models.RouteDirect && len(r.DomainSuffixes) > 0 {
				hasRuTLD = true
			}
			if r.Name == "ru-services-direct" && r.Action == models.RouteDirect && len(r.Domains) > 0 {
				hasRuServices = true
			}
			if r.Name == "proxy-only-messengers" && r.Action == models.RouteProxy && len(r.DomainSuffixes) > 0 {
				hasProxyList = true
			}
			if r.Name == "process-proxy-map" && r.Action == models.RouteProxy && len(r.Processes) > 0 {
				hasProcMap = true
			}
			if r.Name == "dpi-direct-services" && r.Action == models.RouteDPIBypass {
				hasDPI = true
			}
		}
		if !hasXbox || !hasProxyList || !hasDPI || !hasRuTLD || !hasRuServices || !hasProcMap {
			t.Fatalf("rf-smart-split: xbox=%v proxy-list=%v dpi=%v ruTLD=%v ruServices=%v procMap=%v", hasXbox, hasProxyList, hasDPI, hasRuTLD, hasRuServices, hasProcMap)
		}
		return
	}
	t.Fatal("rf-smart-split not found")
}
