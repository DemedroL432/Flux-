package core

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"pulseshield/app/models"
)

// protocolOf возвращает протокол сервера в терминах sing-box:
// vless | vmess | trojan | shadowsocks | hysteria2 | hysteria | tuic | anytls | socks | http | wireguard
func protocolOf(s models.Server) string {
	p := strings.ToLower(strings.TrimSpace(s.Params["_proto"]))
	rawLow := strings.ToLower(strings.TrimSpace(s.RawVLESS))
	if p == "" {
		switch {
		case strings.HasPrefix(rawLow, "hysteria2://"), strings.HasPrefix(rawLow, "hy2://"):
			p = "hysteria2"
		case strings.HasPrefix(rawLow, "hysteria://"):
			p = "hysteria"
		case strings.HasPrefix(rawLow, "trojan://"):
			p = "trojan"
		case strings.HasPrefix(rawLow, "vmess://"):
			p = "vmess"
		case strings.HasPrefix(rawLow, "ss://"):
			p = "shadowsocks"
		case strings.HasPrefix(rawLow, "tuic://"):
			p = "tuic"
		case strings.HasPrefix(rawLow, "anytls://"):
			p = "anytls"
		case strings.HasPrefix(rawLow, "socks://"), strings.HasPrefix(rawLow, "socks5://"):
			p = "socks"
		default:
			p = "vless"
		}
	}
	switch p {
	case "hy2":
		return "hysteria2"
	case "ss":
		return "shadowsocks"
	case "socks4", "socks5":
		return "socks"
	case "https":
		return "http"
	case "wg":
		return "wireguard"
	}
	return p
}

// buildAdditionalOutbound строит outbound для протоколов, которые не покрывает buildVLESSOutbound
// (vmess / shadowsocks / tuic / anytls / hysteria v1 / socks / http / wireguard).
// Ошибка означает, что протокол не поддерживается sing-box — такой узел просто пропускается в пуле.
func buildAdditionalOutbound(s models.Server) (map[string]any, error) {
	proto := protocolOf(s)
	port := s.Port
	if port == "" {
		port = "443"
	}
	serverAddr := resolveServerHostIP(s.Host)

	base := map[string]any{
		"tag":             "proxy",
		"server":          serverAddr,
		"server_port":     mustAtoi(port, 443),
		"domain_resolver": SingBoxBootstrapDNSTag,
		"connect_timeout": "8s",
	}

	switch proto {

	case "vmess":
		security := strings.ToLower(strings.TrimSpace(s.Params["security"]))
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "vmess"
		out["uuid"] = s.UUID
		out["security"] = vmessSecurity(s.Params["method"])
		if aid := strings.TrimSpace(s.Params["alterId"]); aid != "" && aid != "0" {
			out["alter_id"] = mustAtoi(aid, 0)
		} else {
			out["alter_id"] = 0
		}
		applyStreamSettings(out, s, security)
		return out, nil

	case "shadowsocks":
		method := strings.ToLower(strings.TrimSpace(s.Params["method"]))
		if method == "" {
			return nil, fmt.Errorf("для ss:// не указан метод шифрования")
		}
		if !shadowsocksMethodSupported(method) {
			return nil, fmt.Errorf("метод шифрования Shadowsocks %q не поддерживается ядром sing-box — нужен AEAD-метод (aes-128-gcm, aes-256-gcm, chacha20-ietf-poly1305, 2022-blake3-aes-128-gcm и т.п.)", method)
		}
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "shadowsocks"
		out["method"] = method
		out["password"] = s.UUID
		if plugin := strings.ToLower(strings.TrimSpace(s.Params["plugin"])); plugin != "" {
			pluginOpts := strings.TrimSpace(s.Params["plugin-opts"])
			switch {
			case strings.HasPrefix(plugin, "obfs"):
				opts := map[string]any{"mode": "http"}
				if pluginOpts != "" {
					if mode := strings.TrimPrefix(pluginOpts, "obfs="); mode != "" {
						opts["mode"] = mode
					}
				}
				if h := strings.TrimSpace(s.Params["obfs-host"]); h != "" {
					opts["host"] = h
				}
				out["plugin"] = "obfs-local"
				out["plugin_opts"] = obfsOptsString(opts)
			case strings.HasPrefix(plugin, "v2ray"):
				out["plugin"] = "v2ray-plugin"
				out["plugin_opts"] = pluginOpts
			}
		}
		return out, nil

	case "tuic":
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "tuic"
		// В tuic:// учётка — это uuid:password (оба обязательны).
		uuid := firstNonEmpty(s.Params["tuic_uuid"], s.UUID)
		pass := firstNonEmpty(s.Params["tuic_password"], s.Params["password"], s.Params["pass"])
		out["uuid"] = uuid
		out["password"] = pass
		if cc := strings.TrimSpace(s.Params["congestion_control"]); cc != "" {
			out["congestion_control"] = cc
		} else {
			out["congestion_control"] = "bbr"
		}
		if udp := strings.TrimSpace(s.Params["udp_relay_mode"]); udp != "" {
			out["udp_relay_mode"] = udp
		}
		tls := map[string]any{"enabled": true}
		sni := firstNonEmpty(s.Params["sni"], s.Params["peer"], s.Host)
		if sni != "" && !hostLooksLikeIP(sni) {
			tls["server_name"] = sni
		}
		if v := strings.TrimSpace(s.Params["alpn"]); v != "" {
			tls["alpn"] = splitCSV(v)
		}
		if insecureParam(s.Params) {
			tls["insecure"] = true
		}
		out["tls"] = tls
		return out, nil

	case "anytls":
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "anytls"
		out["password"] = s.UUID
		tls := map[string]any{"enabled": true}
		sni := firstNonEmpty(s.Params["sni"], s.Params["peer"], s.Host)
		if sni != "" && !hostLooksLikeIP(sni) {
			tls["server_name"] = sni
		}
		if insecureParam(s.Params) {
			tls["insecure"] = true
		}
		if fp := strings.TrimSpace(s.Params["fp"]); fp != "" {
			tls["utls"] = map[string]any{"enabled": true, "fingerprint": fp}
		}
		out["tls"] = tls
		return out, nil

	case "hysteria":
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "hysteria"
		out["auth_str"] = s.UUID
		if v := strings.TrimSpace(s.Params["upmbps"]); v != "" {
			out["up_mbps"] = mustAtoi(v, 100)
		} else {
			out["up_mbps"] = 100
		}
		if v := strings.TrimSpace(s.Params["downmbps"]); v != "" {
			out["down_mbps"] = mustAtoi(v, 100)
		} else {
			out["down_mbps"] = 100
		}
		if obfs := strings.TrimSpace(s.Params["obfs"]); obfs != "" {
			out["obfs"] = obfs
		}
		tls := map[string]any{"enabled": true}
		sni := firstNonEmpty(s.Params["peer"], s.Params["sni"], s.Host)
		if sni != "" && !hostLooksLikeIP(sni) {
			tls["server_name"] = sni
		}
		if insecureParam(s.Params) {
			tls["insecure"] = true
		}
		out["tls"] = tls
		return out, nil

	case "socks", "http":
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = proto
		version := "5"
		if proto == "http" {
			version = ""
		}
		if strings.TrimSpace(s.Params["username"]) != "" || strings.TrimSpace(s.UUID) != "" {
			user := strings.TrimSpace(s.Params["username"])
			pass := s.UUID
			if user == "" {
				user = s.UUID
				pass = ""
			}
			out["username"] = user
			if pass != "" {
				out["password"] = pass
			}
		}
		if version != "" {
			out["version"] = version
		}
		sec := strings.ToLower(strings.TrimSpace(s.Params["security"]))
		if proto == "http" && sec == "tls" {
			tls := map[string]any{"enabled": true}
			if sni := firstNonEmpty(s.Params["sni"], s.Host); sni != "" && !hostLooksLikeIP(sni) {
				tls["server_name"] = sni
			}
			if insecureParam(s.Params) {
				tls["insecure"] = true
			}
			out["tls"] = tls
		}
		return out, nil

	case "wireguard":
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		out["type"] = "wireguard"
		priv := strings.TrimSpace(s.Params["private_key"])
		if priv == "" {
			priv = strings.TrimSpace(s.UUID)
		}
		pub := strings.TrimSpace(s.Params["public_key"])
		if priv == "" || pub == "" {
			return nil, fmt.Errorf("для wireguard:// нужны private_key и public_key")
		}
		out["private_key"] = priv
		peer := map[string]any{"public_key": pub}
		if psk := strings.TrimSpace(s.Params["preshared_key"]); psk != "" {
			peer["pre_shared_key"] = psk
		}
		if addr := strings.TrimSpace(s.Params["address"]); addr != "" {
			out["local_address"] = splitCSV(addr)
		}
		if mtu := strings.TrimSpace(s.Params["mtu"]); mtu != "" {
			peer["mtu"] = mustAtoi(mtu, 1420)
		}
		out["peers"] = []any{peer}
		return out, nil
	}

	return nil, fmt.Errorf("протокол %q не поддерживается", proto)
}

// applyStreamSettings добавляет TLS/Reality/transport для протоколов с полной моделью транспорта (vmess).
func applyStreamSettings(out map[string]any, s models.Server, security string) {
	fp := strings.TrimSpace(s.Params["fp"])
	if security == "tls" || security == "reality" {
		tls := map[string]any{"enabled": true}
		sni := firstNonEmpty(s.Params["sni"], s.Params["host"], "")
		if sni == "" && !hostLooksLikeIP(s.Host) {
			sni = s.Host
		}
		if sni != "" {
			tls["server_name"] = sni
		}
		if fp == "" {
			fp = "chrome"
		}
		tls["utls"] = map[string]any{"enabled": true, "fingerprint": fp}
		if v := strings.TrimSpace(s.Params["alpn"]); v != "" {
			tls["alpn"] = splitCSV(v)
		}
		if security == "reality" {
			reality := map[string]any{"enabled": true}
			if v := strings.TrimSpace(s.Params["pbk"]); v != "" {
				reality["public_key"] = v
			}
			if v := strings.TrimSpace(s.Params["sid"]); v != "" {
				reality["short_id"] = v
			}
			tls["reality"] = reality
		}
		out["tls"] = tls
	}

	switch strings.ToLower(strings.TrimSpace(s.Params["type"])) {
	case "ws":
		tr := map[string]any{"type": "ws"}
		if p := strings.TrimSpace(s.Params["path"]); p != "" {
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			tr["path"] = p
		}
		if h := firstNonEmpty(s.Params["host"], s.Params["sni"], s.Host); h != "" {
			tr["headers"] = map[string]any{"Host": h}
		}
		out["transport"] = tr
	case "grpc":
		tr := map[string]any{"type": "grpc"}
		if sn := firstNonEmpty(s.Params["serviceName"], s.Params["path"]); sn != "" {
			tr["service_name"] = strings.TrimPrefix(sn, "/")
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
}

func vmessSecurity(method string) string {
	m := strings.ToLower(strings.TrimSpace(method))
	switch m {
	case "", "auto":
		return "auto"
	case "aes-128-gcm", "chacha20-poly1305", "none", "zero":
		return m
	default:
		return "auto"
	}
}

// shadowsocksMethodSupported перечисляет методы, которые понимает sing-box.
func shadowsocksMethodSupported(m string) bool {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "none", "plain",
		"aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
		"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305",
		"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305",
		"aes-128-ctr", "aes-192-ctr", "aes-256-ctr",
		"aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
		"rc4-md5", "chacha20-ietf":
		return true
	}
	return false
}

func obfsOptsString(opts map[string]any) string {
	parts := make([]string, 0, len(opts))
	if mode, ok := opts["mode"]; ok {
		parts = append(parts, "obfs="+fmt.Sprint(mode))
	}
	if host, ok := opts["host"]; ok {
		parts = append(parts, "obfs-host="+fmt.Sprint(host))
	}
	return strings.Join(parts, ";")
}

func insecureParam(params map[string]string) bool {
	for _, k := range []string{"allowInsecure", "allowinsecure", "insecure", "insecure-skip-verify", "skip-cert-verify"} {
		v := strings.ToLower(strings.TrimSpace(params[k]))
		if v == "1" || v == "true" || v == "yes" {
			return true
		}
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if cp := strings.TrimSpace(p); cp != "" {
			out = append(out, cp)
		}
	}
	return out
}

// decodeBase64Any пробует все распространённые варианты base64 (используется для ss-ключей в URI).
func decodeBase64Any(s string) (string, bool) {
	s = strings.TrimSpace(s)
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
			return string(b), true
		}
	}
	return "", false
}

var _ = strconv.Itoa
