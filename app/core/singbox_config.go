package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

type singBoxConfig map[string]any

// SingBoxPrimaryDNSTag — основной DNS-сервер в секции dns (DoH через proxy).
// sing-box 1.12+: нужен в route.default_domain_resolver.
const SingBoxPrimaryDNSTag = "google-doh"

// SingBoxBootstrapDNSTag — прямой системный резолвер (detour: direct) ТОЛЬКО для
// разрешения хостнейма самого VLESS-сервера. Без него возникает loopback: чтобы
// поднять прокси, нужно резолвить его домен, а основной DNS (DoH) сам ходит через
// этот же прокси — замкнутый круг и полное отсутствие интернета.
const SingBoxBootstrapDNSTag = "dns-direct"

func WriteSingBoxConfig(cfg models.AppConfig) (string, error) {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return "", err
	}
	runDir := filepath.Join(dataDir, "runtime")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", err
	}
	raw, err := RenderSingBoxConfig(cfg, runDir)
	if err != nil {
		return "", err
	}
	path := filepath.Join(runDir, "sing-box.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func RenderSingBoxConfig(cfg models.AppConfig, runDir string) ([]byte, error) {
	active, ok := cfg.FindActiveServer()
	if !ok {
		return nil, fmt.Errorf("не выбран активный сервер VLESS — добавьте или выберите сервер на вкладке «Серверы»")
	}
	_, lpw := cfg.LocalProxyAuth() // lpw is reused as the Clash API secret below
	lp := cfg.ListenPort()
	// No auth on the local mixed inbound: it listens only on 127.0.0.1, and the
	// Windows system proxy passes browsers just host:port (no credentials), so any
	// "users" entry causes "authentication failed, no Proxy-Authorization header"
	// and the browser falls back to a direct (unprotected) connection.
	inbounds := []any{
		map[string]any{
			"type":        "mixed",
			"tag":         "mixed-in",
			"listen":      "127.0.0.1",
			"listen_port": lp,
		},
	}
	if cfg.ProxyMode == models.ProxyModeTun {
		tunAddrs := []string{"172.19.0.1/30"}
		if hostHasGlobalIPv6() {
			tunAddrs = append(tunAddrs, "fdfe:dcba:9876::1/126")
		}
		tun := map[string]any{
			"type":           "tun",
			"tag":            "tun-in",
			"interface_name": "pulse-tun",
			"address":        tunAddrs,
			"auto_route":     true,
			"strict_route":   false,
			"stack":          "mixed",
		}
		inbounds = append(inbounds, tun)
	}

	outbounds, err := buildResilientOutbounds(cfg, active)
	if err != nil {
		return nil, err
	}

	isGlobal := models.ResolveRoutingProfileAlias(cfg.RoutingProfile) == models.RoutingModeGlobal

	route := map[string]any{
		"rules":                   buildRouteRules(cfg),
		"final":                   routeFinalOutbound(cfg),
		"default_domain_resolver": SingBoxPrimaryDNSTag,
	}
	if cfg.ProxyMode == models.ProxyModeTun {
		route["auto_detect_interface"] = true
	}

	out := singBoxConfig{
		"log": map[string]any{
			"level":     "info",
			"output":    filepath.Join(runDir, "sing-box.log"),
			"timestamp": true,
		},
		"dns":       buildDNSConfig(cfg.ProxyMode == models.ProxyModeTun, !isGlobal),
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"route":     route,
		// experimental: только cache_file + clash_api без доп. ключей — иначе sing-box 1.12+ может упасть с FATAL unknown field.
		"experimental": map[string]any{
			"cache_file": map[string]any{
				"enabled": true,
				"path":    "cache.db",
			},
			"clash_api": map[string]any{
				"external_controller": fmt.Sprintf("127.0.0.1:%d", SingBoxClashAPIPort),
				// external_ui intentionally omitted — we don't use the Yacd web UI,
				// and sing-box would try to download it from GitHub on every start.
				"secret": lpw,
			},
		},
	}
	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// buildResilientOutbounds строит outbound с тегом "proxy" строго на выбранный активный сервер.
// Группа urltest намеренно не используется, чтобы sing-box никогда не перекидывал трафик
// пользователя на другой (например, российский) узел подписки с меньшим пингом.
func buildResilientOutbounds(_ models.AppConfig, active models.Server) ([]any, error) {
	primaryOut, err := buildVLESSOutbound(active)
	if err != nil {
		return nil, err
	}
	primaryOut["tag"] = "proxy"

	outbounds := []any{
		primaryOut,
		map[string]any{"type": "direct", "tag": "direct"},
		map[string]any{"type": "block", "tag": "block"},
	}
	return outbounds, nil
}

// buildDNSConfig строит DNS-секцию: DoH/DoT через proxy + надёжный bootstrap-резолвер для
// хостнейма VLESS-сервера.
// При ruDirect=true российские домены (.ru/.su/.рф) резолвятся напрямую через bootstrapServer,
// а весь остальной трафик идёт через DoH внутри proxy-туннеля.
// При ruDirect=false (режим global) абсолютно все DNS-запросы идут через DoH внутри прокси.
func buildDNSConfig(tunMode bool, ruDirect bool) map[string]any {
	bootstrapServer := map[string]any{
		"type":        "udp",
		"tag":         SingBoxBootstrapDNSTag,
		"server":      "77.88.8.8",
		"server_port": 53,
	}
	if tunMode {
		bootstrapServer["detour"] = "direct"
	}

	dnsCfg := map[string]any{
		"independent_cache": true,
		"servers": []any{
			map[string]any{
				"type":        "https",
				"tag":         SingBoxPrimaryDNSTag,
				"server":      "8.8.8.8",
				"server_port": 443,
				"path":        "/dns-query",
				"tls":         map[string]any{"enabled": true, "server_name": "dns.google"},
				"detour":      "proxy",
			},
			map[string]any{
				"type":        "tls",
				"tag":         "dot-cloudflare",
				"server":      "1.1.1.1",
				"server_port": 853,
				"tls":         map[string]any{"enabled": true, "server_name": "cloudflare-dns.com"},
				"detour":      "proxy",
			},
			bootstrapServer,
		},
		"final":    SingBoxPrimaryDNSTag,
		"strategy": "prefer_ipv4",
	}
	if ruDirect {
		dnsCfg["rules"] = []any{
			map[string]any{
				"domain_suffix": []string{"ru", "su", "xn--p1ai"},
				"server":        SingBoxBootstrapDNSTag,
			},
		}
	}
	return dnsCfg
}

var resolvedHostCache sync.Map

// resolveServerHostIP пытается быстро разрешить домен VLESS-сервера в IPv4-адрес,
// чтобы передать в sing-box готовый IP. Это полностью исключает ошибки "lookup ...: i/o timeout"
// внутри sing-box, если у пользователя стоит мёртвый/игровой DNS (например 111.88.96.55) или Radmin VPN.
func resolveServerHostIP(host string) string {
	h := strings.TrimSpace(host)
	if h == "" || net.ParseIP(h) != nil {
		return h
	}
	lh := strings.ToLower(h)
	if strings.HasSuffix(lh, ".example.com") || strings.HasSuffix(lh, ".invalid") || strings.HasSuffix(lh, ".test") {
		return h
	}
	if cached, ok := resolvedHostCache.Load(lh); ok {
		return cached.(string)
	}

	pickIPv4 := func(ips []net.IP) string {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil && !v4.IsLoopback() && !v4.IsUnspecified() {
				return v4.String()
			}
		}
		if len(ips) > 0 {
			return ips[0].String()
		}
		return ""
	}

	// 1. Быстрый системный резолвер ОС (1.2 сек)
	ctxSys, cancelSys := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	ips, err := net.DefaultResolver.LookupIP(ctxSys, "ip4", h)
	cancelSys()
	if err == nil {
		if resolved := pickIPv4(ips); resolved != "" {
			resolvedHostCache.Store(lh, resolved)
			return resolved
		}
	}

	// 2. Fallback: прямые запросы к публичным DNS в обход сломанного системного DNS (например 111.88.96.55)
	fallbackDNS := []string{"77.88.8.8:53", "8.8.8.8:53", "1.1.1.1:53"}
	for _, dnsAddr := range fallbackDNS {
		target := dnsAddr
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 1000 * time.Millisecond}
				return d.DialContext(ctx, "udp", target)
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		ips, err := r.LookupIP(ctx, "ip4", h)
		cancel()
		if err == nil {
			if resolved := pickIPv4(ips); resolved != "" {
				resolvedHostCache.Store(lh, resolved)
				return resolved
			}
		}
	}

	return h
}

func buildVLESSOutbound(s models.Server) (map[string]any, error) {
	port := s.Port
	if port == "" {
		port = "443"
	}

	proto := protocolOf(s)
	switch proto {
	case "vless", "trojan", "hysteria2":
		// ниже — основная ветка
	default:
		// vmess / shadowsocks / tuic / anytls / hysteria v1 / socks / http / wireguard
		return buildAdditionalOutbound(s)
	}

	security := strings.ToLower(strings.TrimSpace(s.Params["security"]))
	flow := s.Params["flow"]
	fp := strings.TrimSpace(s.Params["fp"])

	// Разрешаем домен в IP заранее через системный кэш Windows (SNI/Host сохраняются отдельно в tls и ws)
	serverAddr := resolveServerHostIP(s.Host)

	if proto == "hysteria2" {
		peerName, err := tlsPeerNameForVLESS(s, "tls")
		if err != nil {
			peerName = s.Host
		}
		out := map[string]any{
			"type":            "hysteria2",
			"tag":             "proxy",
			"server":          serverAddr,
			"server_port":     mustAtoi(port, 443),
			"password":        s.UUID,
			"domain_resolver": SingBoxBootstrapDNSTag,
			"connect_timeout": "8s",
			"tls": map[string]any{
				"enabled":     true,
				"server_name": peerName,
				"alpn":        []string{"h3"},
			},
		}
		if strings.ToLower(strings.TrimSpace(s.Params["obfs"])) == "salamander" && s.Params["obfs-password"] != "" {
			out["obfs"] = map[string]any{
				"type":     "salamander",
				"password": s.Params["obfs-password"],
			}
		}
		return out, nil
	}

	out := map[string]any{
		"type":            proto,
		"tag":             "proxy",
		"server":          serverAddr,
		"server_port":     mustAtoi(port, 443),
		"domain_resolver": SingBoxBootstrapDNSTag,
		"connect_timeout": "8s",
	}
	if proto == "trojan" {
		out["password"] = s.UUID
		if security == "" {
			security = "tls"
		}
	} else {
		out["uuid"] = s.UUID
		if flow != "" {
			out["flow"] = flow
		}
	}

	network := strings.ToLower(strings.TrimSpace(s.Params["type"]))
	if network == "xhttp" || network == "splithttp" {
		return nil, fmt.Errorf("транспорт %s (Xray XHTTP) не поддерживается ядром sing-box — выберите сервер с типом TCP/Reality, WebSocket или gRPC", network)
	}

	if security == "tls" || security == "reality" {
		peerName, err := tlsPeerNameForVLESS(s, security)
		if err != nil {
			return nil, err
		}
		tls := map[string]any{
			"enabled":     true,
			"server_name": peerName,
		}
		if fp == "" {
			fp = "chrome"
		}
		tls["utls"] = map[string]any{"enabled": true, "fingerprint": fp}

		if security == "tls" {
			if network == "ws" {
				// Для WebSocket через TLS всегда используем только http/1.1 (h2 ломает Upgrade)
				tls["alpn"] = []string{"http/1.1"}
			} else if network == "grpc" {
				tls["alpn"] = []string{"h2"}
			} else if customAlpn := strings.TrimSpace(s.Params["alpn"]); customAlpn != "" {
				parts := strings.Split(customAlpn, ",")
				cleanParts := make([]string, 0, len(parts))
				for _, p := range parts {
					if cp := strings.TrimSpace(p); cp != "" {
						cleanParts = append(cleanParts, cp)
					}
				}
				if len(cleanParts) > 0 {
					tls["alpn"] = cleanParts
				}
			} else {
				tls["alpn"] = []string{"h2", "http/1.1"}
			}
		} else if security == "reality" && network == "grpc" {
			tls["alpn"] = []string{"h2"}
		}

		if security == "reality" {
			reality := map[string]any{"enabled": true}
			if v := s.Params["pbk"]; v != "" {
				reality["public_key"] = v
			}
			if v := s.Params["sid"]; v != "" {
				reality["short_id"] = v
			}
			tls["reality"] = reality
		}
		out["tls"] = tls
	}

	switch network {
	case "ws":
		tr := map[string]any{"type": "ws"}
		if p := strings.TrimSpace(s.Params["path"]); p != "" {
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			tr["path"] = p
		}
		wsHost := strings.TrimSpace(s.Params["host"])
		if wsHost == "" {
			wsHost = strings.TrimSpace(s.Params["sni"])
		}
		if wsHost == "" {
			wsHost = strings.TrimSpace(s.Host)
		}
		if wsHost != "" {
			tr["headers"] = map[string]any{"Host": wsHost}
		}
		out["transport"] = tr
	case "grpc":
		tr := map[string]any{"type": "grpc"}
		if sn := strings.TrimSpace(s.Params["serviceName"]); sn != "" {
			// ВАЖНО: в sing-box service_name не должен начинаться со слэша '/',
			// иначе формируется путь '//serviceName/Tun' и Nginx отвечает HTTP 404!
			tr["service_name"] = strings.TrimPrefix(sn, "/")
		} else if p := strings.TrimSpace(s.Params["path"]); p != "" {
			tr["service_name"] = strings.TrimPrefix(p, "/")
		}
		out["transport"] = tr
	case "http", "h2":
		tr := map[string]any{"type": "http"}
		if p := strings.TrimSpace(s.Params["path"]); p != "" {
			tr["path"] = p
		}
		if h := strings.TrimSpace(s.Params["host"]); h != "" {
			tr["host"] = []string{h}
		}
		out["transport"] = tr
	}

	return out, nil
}

func tlsPeerNameForVLESS(s models.Server, security string) (string, error) {
	cand := strings.TrimSpace(s.Params["sni"])
	if cand == "" {
		cand = strings.TrimSpace(s.Params["serverName"])
	}
	if cand != "" {
		return cand, nil
	}
	for _, k := range []string{"authority", "host"} {
		if v := strings.TrimSpace(s.Params[k]); v != "" && !hostLooksLikeIP(v) {
			return v, nil
		}
	}
	serverHost := strings.TrimSpace(s.Host)
	if serverHost != "" && !hostLooksLikeIP(serverHost) {
		return serverHost, nil
	}
	return "", fmt.Errorf("для VLESS с TLS/Reality укажите SNI (поля sni / serverName или доменный host сервера); подключение по IP без имени маскирует только после добавления SNI под ваш сертификат/Reality")
}

func hostHasGlobalIPv6() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if strings.EqualFold(iface.Name, "pulse-tun") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.To4() != nil {
				continue
			}
			if ip.IsGlobalUnicast() && !ip.IsPrivate() {
				return true
			}
		}
	}
	return false
}

func hostLooksLikeIP(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") && len(host) > 2 {
		host = host[1 : len(host)-1]
	}
	return net.ParseIP(host) != nil
}

func findActiveRoutingProfile(cfg models.AppConfig) (models.RoutingProfile, bool) {
	return cfg.ActiveRoutingProfile()
}

// buildTunServerExcludeRules генерирует правила direct-маршрута для IP-адреса VLESS-сервера.
// КРИТИЧЕСКИ ВАЖНО при auto_route=true: без этого трафик к серверу попадает в TUN-интерфейс
// и создаёт routing loop → бесконечная рекурсия → потеря соединения.
// Если сервер задан доменом — резолвим через системный DNS один раз при генерации конфига.
func buildTunServerExcludeRules(active models.Server) []any {
	host := strings.TrimSpace(active.Host)
	if host == "" {
		return nil
	}
	var addrs []string
	if net.ParseIP(host) != nil {
		addrs = append(addrs, host)
	} else {
		// Резолвим один раз при генерации конфига через системный DNS.
		// Ошибка не фатальна: sing-box имеет domain_resolver для outbound.
		if ips, err := net.LookupHost(host); err == nil {
			addrs = append(addrs, ips...)
		}
	}
	if len(addrs) == 0 {
		return nil
	}
	cidrs := make([]string, 0, len(addrs))
	for _, ip := range addrs {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}
		if parsed.To4() != nil {
			cidrs = append(cidrs, ip+"/32")
		} else {
			cidrs = append(cidrs, ip+"/128")
		}
	}
	if len(cidrs) == 0 {
		return nil
	}
	// IP сервера → direct (через физический шлюз, мимо TUN-интерфейса).
	return []any{
		map[string]any{
			"ip_cidr":  cidrs,
			"outbound": "direct",
		},
	}
}

func buildRouteRules(cfg models.AppConfig) []any {
	zapDir, _ := config.ResolveZapretDirNearExe()
	listsDir := ""
	if zapDir != "" {
		listsDir = filepath.Join(zapDir, "lists")
	}
	merged := config.MergeSingBoxRoutingRules(cfg, listsDir)

	var rules []any
	// Сниффинг TLS/HTTP заголовков: необходим, чтобы доменные правила работали и в TUN, и при подключении по IP.
	rules = append(rules, map[string]any{"action": "sniff", "timeout": "300ms"})

	// TUN-режим: перехватываем DNS до ip_is_private, чтобы запросы к 192.168.x.1:53 роутера не утекали мимо прокси.
	if cfg.ProxyMode == models.ProxyModeTun {
		rules = append(rules, map[string]any{"protocol": "dns", "action": "hijack-dns"})
	}

	// Быстрый путь: локальные сети без тяжёлого сопоставления доменов.
	rules = append(rules, map[string]any{"ip_is_private": true, "outbound": "direct"})

	// TUN-режим: явный direct-маршрут к IP VLESS-сервера ДОЛЖЕН быть до общих правил.
	// Без него трафик к серверу попадает в TUN → routing loop → нет интернета.
	if cfg.ProxyMode == models.ProxyModeTun {
		if active, ok := cfg.FindActiveServer(); ok {
			rules = append(rules, buildTunServerExcludeRules(active)...)
		}
	}

	for _, r := range merged {
		if len(r.Domains) == 0 && len(r.DomainSuffixes) == 0 && len(r.CIDRs) == 0 && len(r.Processes) == 0 {
			// Безусловный catch-all ("default-proxy") обрабатывается через route.final,
			// иначе пустое правило {"outbound": "proxy"} в середине списка перекрывает все правила после него!
			continue
		}
		var exactDomains []string
		var suffixes []string
		suffixSeen := make(map[string]bool, len(r.DomainSuffixes)+len(r.Domains))
		for _, s := range r.DomainSuffixes {
			clean := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(s)), "*."), ".")
			if clean != "" && !suffixSeen[clean] {
				suffixSeen[clean] = true
				suffixes = append(suffixes, clean)
			}
		}
		for _, d := range r.Domains {
			t := strings.TrimSpace(d)
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, "*.") || strings.HasPrefix(t, ".") {
				clean := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(t), "*."), ".")
				if clean != "" && !suffixSeen[clean] {
					suffixSeen[clean] = true
					suffixes = append(suffixes, clean)
				}
				continue
			}
			exactDomains = append(exactDomains, t)
		}

		if len(exactDomains) == 0 && len(suffixes) == 0 && len(r.CIDRs) == 0 && len(r.Processes) == 0 {
			continue
		}
		rule := map[string]any{}
		if len(exactDomains) > 0 {
			rule["domain"] = exactDomains
		}
		if len(suffixes) > 0 {
			rule["domain_suffix"] = suffixes
		}
		if len(r.CIDRs) > 0 {
			rule["ip_cidr"] = r.CIDRs
		}
		if len(r.Processes) > 0 {
			rule["process_name"] = r.Processes
		}
		switch r.Action {
		case models.RouteProxy:
			rule["outbound"] = "proxy"
		case models.RouteDirect, models.RouteDPIBypass:
			rule["outbound"] = "direct"
		case models.RouteBlock:
			rule["outbound"] = "block"
		default:
			rule["outbound"] = "proxy"
		}
		rules = append(rules, rule)
	}
	return rules
}

func routeFinalOutbound(cfg models.AppConfig) string {
	if p, ok := findActiveRoutingProfile(cfg); ok {
		switch strings.ToLower(strings.TrimSpace(p.Final)) {
		case "direct":
			return "direct"
		case "proxy":
			return "proxy"
		}
	}
	switch cfg.ProxyMode {
	case models.ProxyModeDirect:
		return "direct"
	default:
		return "proxy"
	}
}

func mustAtoi(s string, fallback int) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
