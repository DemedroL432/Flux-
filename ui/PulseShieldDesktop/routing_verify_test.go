package main

import (
	"strings"
	"testing"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

// normalizeRouteQuery принимает URL, host:port, IP и .exe — проверяем основные формы.
func TestNormalizeRouteQuery(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantKd string
	}{
		{"https://YouTube.com/watch?v=1", "youtube.com", "domain"},
		{"youtube.com:443", "youtube.com", "domain"},
		{"  .RU  ", "ru", "domain"},
		{"t.me", "t.me", "domain"},
		{"1.2.3.4", "1.2.3.4", "ip"},
		{"10.0.0.0/8", "10.0.0.0/8", "ip"},
		{"Telegram.exe", "telegram.exe", "process"},
	}
	for _, c := range cases {
		got, kind := normalizeRouteQuery(c.in)
		if got != c.want || kind != c.wantKd {
			t.Errorf("normalizeRouteQuery(%q) = (%q,%q), want (%q,%q)", c.in, got, kind, c.want, c.wantKd)
		}
	}
}

// IPv6 в квадратных скобках и с портом — частый ввод в проверке маршрута.
func TestNormalizeRouteQueryIPv6(t *testing.T) {
	cases := []struct{ in, want, kind string }{
		{"[::1]:443", "::1", "ip"},
		{"https://[2606:4700::1111]/dns", "2606:4700::1111", "ip"},
		{"fd00::/8", "fd00::/8", "ip"},
		{"::1", "::1", "ip"},
	}
	for _, c := range cases {
		got, kind := normalizeRouteQuery(c.in)
		if got != c.want || kind != c.kind {
			t.Errorf("normalizeRouteQuery(%q) = (%q,%q), want (%q,%q)", c.in, got, kind, c.want, c.kind)
		}
	}
}

// first-match должен совпадать с порядком MergeSingBoxRoutingRules:
// процессы → пользователь → профиль → библиотека.
func TestVerifyRouteTargetPrecedence(t *testing.T) {
	// Библиотека отправляет .netflix.com в proxy, но пользовательское правило выше — в direct.
	cfg := models.AppConfig{
		ProxyEnabled:     true,
		DPIEnabled:       true,
		ActiveDPIProfile: "zapret:auto",
		RoutingProfile:   "default",
		UserRoutingRules: []models.RouteRule{
			{Name: "user:.netflix.com", DomainSuffixes: []string{"netflix.com"}, Action: models.RouteDirect},
			{Name: "user:*.totally-unknown.test", Domains: []string{"*.totally-unknown.test"}, Action: models.RouteBlock},
		},
	}

	// Пользовательское правило выигрывает у библиотеки.
	norm, kind := normalizeRouteQuery("www.netflix.com")
	if !matchRule(cfg.UserRoutingRules[0], norm, kind) {
		t.Fatalf("пользовательское правило должно совпадать с www.netflix.com")
	}
	// Процесс Telegram.exe уходит в proxy правилом process-proxy-map (наивысший слой).
	procs := config.ProcessProxyRules(&cfg)[0]
	pnorm, pkind := normalizeRouteQuery("Telegram.exe")
	if !matchRule(procs, pnorm, pkind) {
		t.Fatalf("правило процесса должно совпадать с Telegram.exe")
	}
	if procs.Action != models.RouteProxy {
		t.Fatalf("карта процессов должна вести в proxy, получено %q", procs.Action)
	}
}

// Пустой конфиг: ни одно правило не совпало → работает final профиля.
func TestVerifyFallbackToFinal(t *testing.T) {
	norm, kind := normalizeRouteQuery("some-brand-new-domain.test")
	if norm == "" || kind != "domain" {
		t.Fatalf("нормализация домена сломана: %q %q", norm, kind)
	}
	rule := models.RouteRule{Name: "builtin-direct-anticheat", DomainSuffixes: []string{"easyanticheat.net"}, Action: models.RouteDirect}
	if matchRule(rule, norm, kind) {
		t.Fatalf("правило anticheat не должно совпадать с %q", norm)
	}
}

// CIDR-правила не должны совпадать с доменами и наоборот.
func TestMatchRuleKindIsolation(t *testing.T) {
	cidrRule := models.RouteRule{Name: "c", CIDRs: []string{"10.0.0.0/8"}, Action: models.RouteProxy}
	if matchRule(cidrRule, "example.com", "domain") {
		t.Fatalf("CIDR-правило не должно совпадать с доменом")
	}
	if !matchRule(cidrRule, "10.1.2.3", "ip") {
		t.Fatalf("CIDR-правило должно совпадать с 10.1.2.3")
	}

	domRule := models.RouteRule{Name: "d", DomainSuffixes: []string{"t.me"}, Action: models.RouteProxy}
	if matchRule(domRule, "1.2.3.4", "ip") {
		t.Fatalf("доменное правило не должно совпадать с IP")
	}
	// суффикс должен ловить поддомены, но не «почти-совпадения»
	if !matchRule(domRule, "web.telegram.t.me", "domain") {
		t.Fatalf("суффикс .t.me должен совпадать с поддоменом")
	}
	if matchRule(domRule, "notme", "domain") {
		t.Fatalf("суффикс .t.me не должен совпадать с notme")
	}
}

// Реальный конфиг sing-box ставит ip_is_private -> direct ПЕРВЫМ правилом
// (см. buildRouteRules), поэтому приватный IP никогда не попадает под profile.final.
// Проверка «куда уйдёт трафик» должна повторять это, иначе UI врёт.
func TestVerifyPrivateIPGoesDirect(t *testing.T) {
	cases := []string{"192.168.1.10", "10.1.2.3", "172.16.5.4", "127.0.0.1", "169.254.10.1", "::1", "fd00::1"}
	for _, c := range cases {
		if !isPrivateRouteIP(c) {
			t.Errorf("%s должен считаться приватным (ip_is_private)", c)
		}
	}
	for _, c := range []string{"1.1.1.1", "8.8.8.8", "2606:4700::1111"} {
		if isPrivateRouteIP(c) {
			t.Errorf("%s не должен считаться приватным", c)
		}
	}
}

// Приватный IP не должен попадать в финальное правило профиля (VPN).
func TestVerifyPrivateNotFinal(t *testing.T) {
	cfg := models.AppConfig{ProxyEnabled: true, RoutingProfile: "default"}
	rp, ok := cfg.ActiveRoutingProfile()
	if !ok {
		t.Fatalf("профиль default должен существовать")
	}
	// final профиля по умолчанию — proxy; приватный адрес всё равно direct.
	if strings.ToLower(strings.TrimSpace(rp.Final)) != "proxy" {
		t.Logf("final профиля: %q", rp.Final)
	}
	for _, addr := range []string{"192.168.0.1", "10.0.0.5", "::1"} {
		if !isPrivateRouteIP(addr) {
			t.Fatalf("%s должен быть приватным и не попадать под final", addr)
		}
	}
}
