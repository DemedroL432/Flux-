package multiserver

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pulseshield/app/models"
)

// Поддерживаемые URI-схемы прокси-ссылок. sing-box умеет говорить по всем из них,
// поэтому мы принимаем любую ссылку, а не только vless://.
//
//	vless://  vmess://  trojan://  ss://      ssr://     hysteria:// hysteria2://
//	hy2://    tuic://   anytls://  socks://   socks5://  http:// (proxy)  wireguard://
//
// Дополнительно принимаем алиасы, которые встречаются у панелей (Happ, v2rayN, Marzban,
// sing-box-совместимые подписки): ss2:// (то же, что ss://), socks5h:// / socks4a://
// (socks с удалённым DNS), hy:// (короткая форма hysteria://), wg:// (короткая wireguard://).
var supportedSchemes = []string{
	"vless", "vmess", "trojan", "ss", "ssr", "hysteria", "hysteria2", "hy2", "hy",
	"tuic", "anytls", "socks", "socks5", "socks5h", "socks4", "socks4a",
	"http", "https", "wireguard", "wg", "ss2",
}

// knownProxySchemes — используется для быстрой проверки «похоже ли это на прокси-ссылку».
var knownProxySchemes = func() map[string]bool {
	m := make(map[string]bool, len(supportedSchemes))
	for _, s := range supportedSchemes {
		m[s] = true
	}
	return m
}()

// schemeFromRaw возвращает схему URI в нижнем регистре или "".
func schemeFromRaw(raw string) string {
	raw = strings.TrimSpace(raw)
	i := strings.Index(raw, "://")
	if i <= 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(raw[:i]))
}

// IsProxyLink сообщает, является ли строка поддерживаемой прокси-ссылкой.
func IsProxyLink(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	low := strings.ToLower(raw)
	// ss:// может быть и ss://..., и ss2://; vmess:// бывает у некоторых панелей как vmess://base64.
	scheme := schemeFromRaw(low)
	if scheme == "" {
		return false
	}
	if !knownProxySchemes[scheme] {
		return false
	}
	// http:// / https:// в контексте «списка узлов» тоже прокси-ссылка (HTTP proxy в base64),
	// но в общем контексте импорта URL подписки это обычная ссылка. Различаем по наличию
	// base64-креденшелов после схемы.
	if scheme == "http" || scheme == "https" {
		return looksLikeHTTPProxyLink(raw)
	}
	return true
}

// looksLikeHTTPProxyLink отличает http://user:pass@host:port (прокси) от https://example.com/sub (подписка).
func looksLikeHTTPProxyLink(raw string) bool {
	if strings.Contains(raw, "@") {
		return true
	}
	// Панели часто отдают http-прокси как "http://base64(user:pass)@host:port" — уже покрыто '@'.
	// Также встречается схема "http://base64" без '@': декодируем и проверяем "user:pass".
	return false
}

// ProtocolOf возвращает протокол сервера ("vless", "vmess", ...), нормализуя hy2 → hysteria2.
func ProtocolOf(srv *models.VlessServer) string {
	if srv == nil {
		return ""
	}
	if p := strings.ToLower(strings.TrimSpace(srv.Params["_proto"])); p != "" {
		if p == "hy2" {
			return "hysteria2"
		}
		return p
	}
	scheme := schemeFromRaw(srv.RawVLESS)
	switch scheme {
	case "hy2", "hy", "hysteria2":
		return "hysteria2"
	case "hysteria":
		return "hysteria"
	case "ss", "ss2":
		return "shadowsocks"
	case "wg", "wireguard":
		return "wireguard"
	case "socks4", "socks4a", "socks5h":
		return "socks"
	case "https":
		return "http"
	}
	if scheme != "" && knownProxySchemes[scheme] {
		return scheme
	}
	return "vless"
}

// ParseProxyLink разбирает любую поддерживаемую прокси-ссылку в модель сервера.
func ParseProxyLink(raw string) (*models.VlessServer, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("пустая ссылка")
	}
	switch schemeFromRaw(raw) {
	case "vmess":
		return parseVmessLink(raw)
	case "ss", "ss2":
		return parseShadowsocksLink(raw)
	case "ssr":
		return nil, fmt.Errorf("протокол ShadowsocksR (ssr://) не поддерживается ядром sing-box — используйте ss:// или vless://")
	case "hysteria", "hy":
		return parseHysteriaV1Link(raw)
	case "tuic", "anytls", "wireguard", "wg", "vless", "trojan", "hysteria2", "hy2",
		"socks", "socks5", "socks5h", "socks4", "socks4a", "http", "https":
		return parseGenericURLLink(raw)
	default:
		if IsProxyLink(raw) {
			return parseGenericURLLink(raw)
		}
		return nil, fmt.Errorf("неизвестный формат ссылки")
	}
}

// parseGenericURLLink разбирает ссылки вида scheme://user:pass@host:port?params#Name,
// которые покрывают vless/trojan/hysteria2/socks/http/tuic/anytls/wireguard.
func parseGenericURLLink(raw string) (*models.VlessServer, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("в ссылке не указан адрес сервера")
	}
	if isDummyHost(host) {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}
	port := u.Port()
	if port == "" {
		port = defaultPortForScheme(scheme)
	}
	if port == "0" || port == "1" {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}

	params := map[string]string{}
	for k, vs := range u.Query() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	proto := scheme
	switch scheme {
	case "hy2", "hy":
		proto = "hysteria2"
	case "socks5", "socks4", "socks5h", "socks4a":
		proto = "socks"
	case "https":
		proto = "http"
	case "wg":
		proto = "wireguard"
	case "ss2":
		proto = "shadowsocks"
	}
	params["_proto"] = proto

	netType := strings.ToLower(strings.TrimSpace(params["type"]))
	if netType == "xhttp" || netType == "splithttp" {
		return nil, fmt.Errorf("транспорт %s (Xray XHTTP) не поддерживается ядром sing-box", netType)
	}

	// Учётные данные: для большинства протоколов это пароль/uuid, для socks/http — логин+пароль.
	secret := ""
	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			switch proto {
			case "socks", "http":
				// Логин храним в params, пароль — в UUID (как и для остальных протоколов).
				params["username"] = u.User.Username()
				secret = pw
			case "tuic":
				// В tuic:// учётка состоит из UUID и пароля: оба нужны sing-box.
				params["tuic_uuid"] = u.User.Username()
				params["tuic_password"] = pw
				secret = pw
			default:
				secret = pw
			}
		} else {
			secret = u.User.Username()
		}
	}
	if decoded, err := url.PathUnescape(secret); err == nil && decoded != secret {
		secret = decoded
	}
	if proto != "vless" && params["proto"] == "" {
		params["proto"] = proto
	}
	if secret == "00000000-0000-0000-0000-000000000000" {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}

	name := decodeFragment(u.Fragment)
	if name == "" {
		name = host
	}

	srv := &models.VlessServer{
		ID:        serverID(raw, host, port),
		Name:      name,
		RawVLESS:  raw,
		Host:      host,
		Port:      port,
		UUID:      secret,
		Params:    params,
		CreatedAt: time.Now(),
	}
	if proto == "anytls" && srv.UUID == "" {
		return nil, fmt.Errorf("в ссылке anytls:// не указан пароль")
	}
	return srv, nil
}

func defaultPortForScheme(scheme string) string {
	switch scheme {
	case "socks", "socks4", "socks4a", "socks5", "socks5h":
		return "1080"
	case "http", "https":
		return "8080"
	case "ss", "ss2":
		return "8388"
	case "wireguard", "wg":
		return "51820"
	case "hysteria", "hy":
		return "36712"
	default:
		return "443"
	}
}

func decodeFragment(f string) string {
	f = strings.TrimSpace(f)
	if f == "" {
		return ""
	}
	if unesc, err := url.QueryUnescape(f); err == nil && unesc != "" {
		f = unesc
	}
	if unesc, err := url.PathUnescape(f); err == nil && unesc != "" {
		f = unesc
	}
	return strings.TrimSpace(f)
}

func isDummyHost(host string) bool {
	switch strings.ToLower(host) {
	case "127.0.0.1", "0.0.0.0", "localhost", "::1":
		return true
	}
	return false
}

// parseHysteriaV1Link разбирает hysteria:// (v1) ссылку.
// Формат: hysteria://host:port?auth=xxx&upmbps=100&downmbps=100&peer=sni&insecure=1&obfs=xplus&obfsParam=xxx#name
func parseHysteriaV1Link(raw string) (*models.VlessServer, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("в ссылке hysteria:// не указан адрес сервера")
	}
	if isDummyHost(host) {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}
	port := u.Port()
	if port == "" {
		port = "36712"
	}
	if port == "0" || port == "1" {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}
	params := map[string]string{}
	for k, vs := range u.Query() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	params["_proto"] = "hysteria"
	auth := params["auth"]
	if auth == "" && u.User != nil {
		auth = u.User.Username()
	}
	name := decodeFragment(u.Fragment)
	if name == "" {
		name = host
	}
	return &models.VlessServer{
		ID:        serverID(raw, host, port),
		Name:      name,
		RawVLESS:  raw,
		Host:      host,
		Port:      port,
		UUID:      auth,
		Params:    params,
		CreatedAt: time.Now(),
	}, nil
}

// parseShadowsocksLink разбирает ss:// в двух вариантах:
//  1. ss://base64(method:password)@host:port#name
//  2. ss://base64(method:password@host:port)#name  (legacy SIP002)
//  3. ss://method:password@host:port?plugin=...#name (plain, редко)
func parseShadowsocksLink(raw string) (*models.VlessServer, error) {
	// Схема может быть ss:// или ss2:// — отрезаем по фактическому разделителю.
	body := strings.TrimSpace(raw)
	if i := strings.Index(body, "://"); i >= 0 {
		body = strings.TrimSpace(body[i+3:])
	} else {
		body = strings.TrimSpace(raw[len("ss://"):])
	}
	if body == "" {
		return nil, fmt.Errorf("пустая ss:// ссылка")
	}
	// Отрезаем fragment.
	name := ""
	if i := strings.LastIndex(body, "#"); i >= 0 {
		name = decodeFragment(body[i+1:])
		body = body[:i]
	}
	// Отрезаем query (plugin и пр.).
	query := ""
	if i := strings.Index(body, "?"); i >= 0 {
		query = body[i+1:]
		body = body[:i]
	}
	params := map[string]string{}
	if query != "" {
		if q, err := url.ParseQuery(query); err == nil {
			for k, vs := range q {
				if len(vs) > 0 {
					params[k] = vs[0]
				}
			}
		}
	}
	params["_proto"] = "shadowsocks"

	var method, password, host, port string
	if at := strings.LastIndex(body, "@"); at >= 0 {
		// Варианты 1 и 3.
		cred, addr := body[:at], body[at+1:]
		if dec, ok := decodeBase64Loose(cred); ok && strings.Contains(dec, ":") {
			cred = dec
		} else if unesc, err := url.PathUnescape(cred); err == nil {
			cred = unesc
		}
		if i := strings.Index(cred, ":"); i >= 0 {
			method = strings.TrimSpace(cred[:i])
			password = strings.TrimSpace(cred[i+1:])
		}
		h, p := splitHostPortLoose(addr)
		host, port = h, p
	} else {
		// Вариант 2: целиком base64.
		dec, ok := decodeBase64Loose(body)
		if !ok {
			return nil, fmt.Errorf("не удалось декодировать ss:// ссылку")
		}
		dec = strings.TrimSpace(dec)
		at := strings.LastIndex(dec, "@")
		if at < 0 {
			return nil, fmt.Errorf("некорректная ss:// ссылка: нет разделителя @")
		}
		cred, addr := dec[:at], dec[at+1:]
		if i := strings.Index(cred, ":"); i >= 0 {
			method = strings.TrimSpace(cred[:i])
			password = strings.TrimSpace(cred[i+1:])
		}
		h, p := splitHostPortLoose(addr)
		host, port = h, p
	}
	host = strings.TrimSpace(host)
	if host == "" || isDummyHost(host) {
		return nil, fmt.Errorf("в ссылке ss:// не указан адрес сервера")
	}
	if port == "" {
		port = "8388"
	}
	if port == "0" || port == "1" {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}
	if method == "" || password == "" {
		return nil, fmt.Errorf("некорректная ss:// ссылка: нет метода или пароля")
	}
	params["method"] = normalizeShadowsocksMethod(method)
	if name == "" {
		name = host
	}
	return &models.VlessServer{
		ID:     serverID(raw, host, port),
		Name:   name,
		Host:   host,
		Port:   port,
		UUID:   password,
		Params: params,
	}, nil
}

// normalizeShadowsocksMethod приводит популярные алиасы методов к именам sing-box.
func normalizeShadowsocksMethod(m string) string {
	m = strings.ToLower(strings.TrimSpace(m))
	switch m {
	case "aes-256-cfb", "aes-192-cfb", "aes-128-cfb", "aes-256-cfb1", "aes-256-cfb8",
		"aes-192-cfb1", "aes-192-cfb8", "aes-128-cfb1", "aes-128-cfb8",
		"rc4-md5", "bf-cfb", "cast5-cfb", "des-cfb", "rc2-cfb", "idea-cfb", "seed-cfb":
		// Старые (не-AEAD) методы sing-box не поддерживает; ближайший совместимый — aes-256-gcm
		// всё равно не даст соединения, поэтому оставляем как есть и покажем ошибку на запуске.
		return m
	case "chacha20-ietf":
		return "chacha20-ietf-poly1305"
	case "xchacha20":
		return "xchacha20-ietf-poly1305"
	}
	return m
}

func splitHostPortLoose(addr string) (string, string) {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimSuffix(addr, "/")
	// IPv6: [::1]:443
	if strings.HasPrefix(addr, "[") {
		if i := strings.Index(addr, "]"); i > 0 {
			host := addr[1:i]
			rest := strings.TrimPrefix(addr[i+1:], ":")
			return host, strings.TrimSpace(rest)
		}
	}
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return strings.TrimSpace(addr[:i]), strings.TrimSpace(addr[i+1:])
	}
	return addr, ""
}

func decodeBase64Loose(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	if s == "" {
		return "", false
	}
	if i := strings.Index(s, "?"); i >= 0 {
		s = s[:i]
	}
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
			return string(b), true
		}
	}
	// Некоторые панели отдают base64 без паддинга с неправильной длиной — добиваем «=».
	pad := s
	for len(pad)%4 != 0 {
		pad += "="
	}
	if b, err := base64.StdEncoding.DecodeString(pad); err == nil && len(b) > 0 {
		return string(b), true
	}
	if b, err := base64.URLEncoding.DecodeString(pad); err == nil && len(b) > 0 {
		return string(b), true
	}
	return "", false
}

// parseVmessLink разбирает vmess:// формата v2rayN: vmess://base64(json).
// Поддерживаем оба регистра ключей (ps/name, add/address, port, id/uuid, aid, scy, net, type,
// host, path, tls, sni, alpn, fp) и нестандартный «vmess://uuid@host:port?...» (v2rayNG plain).
func parseVmessLink(raw string) (*models.VlessServer, error) {
	body := strings.TrimSpace(raw[len("vmess://"):])
	if body == "" {
		return nil, fmt.Errorf("пустая vmess:// ссылка")
	}
	// Plain-вариант: vmess://uuid@host:port?type=ws&...#Name
	if strings.Contains(body, "@") && !strings.HasPrefix(body, "ey") && strings.Contains(body, ":") && strings.Contains(body, ".") && !looksBase64JSON(body) {
		u, err := url.Parse(raw)
		if err == nil && u.Hostname() != "" {
			params := map[string]string{}
			for k, vs := range u.Query() {
				if len(vs) > 0 {
					params[k] = vs[0]
				}
			}
			params["_proto"] = "vmess"
			params["vmess-plain"] = "1"
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				port = "443"
			}
			name := decodeFragment(u.Fragment)
			if name == "" {
				name = host
			}
			return &models.VlessServer{
				ID:     serverID(raw, host, port),
				Name:   name,
				Host:   host,
				Port:   port,
				UUID:   u.User.Username(),
				Params: params,
			}, nil
		}
	}

	dec, ok := decodeBase64Loose(body)
	if !ok {
		// Иногда встречается vmess:// с уже JSON внутри (без base64).
		if strings.HasPrefix(strings.TrimSpace(body), "{") {
			dec = strings.TrimSpace(body)
		} else {
			return nil, fmt.Errorf("не удалось декодировать vmess:// ссылку (ожидался base64 JSON)")
		}
	}
	m := parseFlatJSONObject(dec)
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		// Ключи регистронезависимо.
		for mk, mv := range m {
			for _, k := range keys {
				if strings.EqualFold(mk, k) && strings.TrimSpace(mv) != "" {
					return strings.TrimSpace(mv)
				}
			}
		}
		return ""
	}

	host := get("add", "address", "addr", "server", "host")
	if host == "" || isDummyHost(host) {
		return nil, fmt.Errorf("в vmess:// ссылке не указан адрес сервера")
	}
	port := get("port")
	if port == "" {
		port = "443"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("некорректный порт в vmess:// (%q)", port)
	}
	if port == "0" || port == "1" {
		return nil, fmt.Errorf("служебный (пустой) узел подписки")
	}
	uuid := get("id", "uuid", "user", "password")
	if uuid == "" {
		return nil, fmt.Errorf("в vmess:// ссылке не указан id (uuid)")
	}

	params := map[string]string{}
	params["_proto"] = "vmess"
	params["vmess-meta"] = "1"
	if v := get("aid", "alterId", "alterid"); v != "" && v != "0" {
		params["alterId"] = v
	}
	if v := get("scy", "security", "cipher"); v != "" {
		params["method"] = strings.ToLower(v)
	}
	if v := get("net", "network"); v != "" {
		params["type"] = strings.ToLower(v)
	}
	if v := get("type", "headerType"); v != "" && !strings.EqualFold(v, "none") {
		params["headerType"] = v
	}
	if v := get("host", "wsHost"); v != "" {
		params["host"] = v
	}
	if v := get("path", "wsPath"); v != "" {
		params["path"] = v
	}
	if v := get("sni", "peer"); v != "" {
		params["sni"] = v
	}
	if v := get("tls"); v != "" {
		low := strings.ToLower(v)
		if low == "tls" || low == "reality" || low == "true" {
			params["security"] = "tls"
		} else {
			params["security"] = low
		}
	}
	if v := get("alpn"); v != "" {
		params["alpn"] = v
	}
	if v := get("fp", "fingerprint"); v != "" {
		params["fp"] = v
	}
	if v := get("serviceName", "service_name"); v != "" {
		params["serviceName"] = v
	}
	if v := get("grpc-mode", "grpcMode"); v != "" {
		params["grpc-mode"] = v
	}
	for k, v := range m {
		lk := strings.ToLower(k)
		if lk == "allowinsecure" || lk == "skip-cert-verify" {
			if v == "1" || strings.EqualFold(v, "true") {
				params["allowInsecure"] = "1"
			}
		}
	}

	name := get("ps", "name", "remark", "remarks")
	if name == "" {
		name = host
	}
	return &models.VlessServer{
		ID:     serverID(raw, host, port),
		Name:   name,
		Host:   host,
		Port:   port,
		UUID:   uuid,
		Params: params,
	}, nil
}

func looksBase64JSON(s string) bool {
	dec, ok := decodeBase64Loose(s)
	return ok && strings.HasPrefix(strings.TrimSpace(dec), "{")
}

// parseFlatJSONObject превращает плоский JSON-объект {"k":"v","n":1} в map[string]string.
// Реализация намеренно простая и не требует импорта encoding/json в горячих путях UI,
// но полностью корректно обрабатывает экранирование и любые скалярные значения.
func parseFlatJSONObject(src string) map[string]string {
	out := map[string]string{}
	src = strings.TrimSpace(src)
	if !strings.HasPrefix(src, "{") {
		if i := strings.Index(src, "{"); i >= 0 {
			src = src[i:]
		} else {
			return out
		}
	}
	i := 1
	n := len(src)
	skipWS := func() {
		for i < n {
			switch src[i] {
			case ' ', '\t', '\r', '\n', ',':
				i++
			default:
				return
			}
		}
	}
	readString := func() (string, bool) {
		if i >= n || src[i] != '"' {
			return "", false
		}
		i++
		var b strings.Builder
		for i < n {
			c := src[i]
			if c == '\\' && i+1 < n {
				i++
				esc := src[i]
				switch esc {
				case 'n':
					b.WriteByte('\n')
				case 'r':
					b.WriteByte('\r')
				case 't':
					b.WriteByte('\t')
				case 'u':
					if i+4 < n {
						if r, err := strconv.ParseUint(src[i+1:i+5], 16, 32); err == nil {
							b.WriteRune(rune(r))
							i += 4
						}
					}
				default:
					b.WriteByte(esc)
				}
				i++
				continue
			}
			if c == '"' {
				i++
				return b.String(), true
			}
			b.WriteByte(c)
			i++
		}
		return b.String(), true
	}
	readScalar := func() string {
		start := i
		for i < n {
			c := src[i]
			if c == ',' || c == '}' || c == '\n' {
				break
			}
			i++
		}
		return strings.TrimSpace(src[start:i])
	}
	for i < n {
		skipWS()
		if i >= n || src[i] == '}' {
			break
		}
		key, ok := readString()
		if !ok {
			// не строка — прекращаем разбор
			break
		}
		skipWS()
		if i < n && src[i] == ':' {
			i++
		}
		skipWS()
		var val string
		if i < n && src[i] == '"' {
			val, _ = readString()
		} else if i < n && src[i] == '{' {
			// Вложенный объект пропускаем целиком.
			depth := 0
			for i < n {
				if src[i] == '{' {
					depth++
				} else if src[i] == '}' {
					depth--
					if depth == 0 {
						i++
						break
					}
				}
				i++
			}
			continue
		} else if i < n && src[i] == '[' {
			depth := 0
			start := i
			for i < n {
				if src[i] == '[' {
					depth++
				} else if src[i] == ']' {
					depth--
					if depth == 0 {
						i++
						break
					}
				}
				i++
			}
			val = strings.TrimSpace(src[start:i])
			val = strings.Trim(val, "[]")
			val = strings.ReplaceAll(val, `"`, "")
		} else {
			val = readScalar()
		}
		out[key] = val
	}
	return out
}
