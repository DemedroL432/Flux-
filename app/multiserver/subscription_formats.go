package multiserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"pulseshield/app/models"
)

// ---------------------------------------------------------------------------
// Разбор подписок любого формата.
//
// Поддерживаются:
//   • plain-список URI (vless/vmess/trojan/ss/ssr/hysteria/hysteria2/tuic/anytls/socks/http)
//   • base64 от такого списка (в любом варианте кодирования и с переносами строк)
//   • base64 от base64 (двойное кодирование, встречается у некоторых панелей)
//   • Clash / Clash.Meta / Stash YAML (proxies: + proxy-groups:)
//   • sing-box JSON (outbounds:)
//   • Xray / v2ray JSON (outbounds: с settings.vnext / servers)
//   • JSON-массив строк или объект {"proxies": [...]}
// ---------------------------------------------------------------------------

// subscriptionFormats — список форматов, которые умеет разбирать приложение.
var subscriptionFormats = []string{
	"plain-uri", "base64", "clash-yaml", "sing-box-json", "v2ray-json",
}

// ParseSubscriptionBody превращает тело ответа подписки в список серверов.
func ParseSubscriptionBody(body string) ([]*models.VlessServer, error) {
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		return nil, err
	}
	servers := make([]*models.VlessServer, 0, len(links))
	var lastErr error
	for _, link := range links {
		srv, err := ParseProxyLink(link)
		if err != nil {
			lastErr = err
			continue
		}
		servers = append(servers, srv)
	}
	if len(servers) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("в подписке не найдено поддерживаемых узлов: %w", lastErr)
		}
		return nil, fmt.Errorf("в подписке не найдено поддерживаемых узлов")
	}
	return servers, nil
}

// ExtractSubscriptionLinks вытаскивает из тела подписки список прокси-ссылок,
// автоматически определяя формат (plain, base64, Clash YAML, sing-box/Xray JSON).
func ExtractSubscriptionLinks(body string) ([]string, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil, fmt.Errorf("пустой ответ подписки")
	}

	// 1. JSON (sing-box / Xray / {"proxies":[...]}) — проверяем раньше YAML.
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if links, ok := parseJSONSubscription(trimmed); ok && len(links) > 0 {
			return links, nil
		}
	}

	// 2. Clash / Clash.Meta YAML (в том числе с вложенными ws-opts).
	if looksLikeClashYAML(trimmed) {
		if links, ok := parseClashYAML(trimmed); ok && len(links) > 0 {
			return links, nil
		}
	}

	// 2b. YAML, обёрнутый в base64 (часть панелей кодирует целиком весь конфиг Clash).
	if !strings.Contains(trimmed, "\n") {
		if dec, ok := decodeSubscriptionBody(trimmed); ok && looksLikeClashYAML(dec) {
			if links, ok := parseClashYAML(dec); ok && len(links) > 0 {
				return links, nil
			}
		}
	}

	// 3. plain-список URI.
	if links := splitProxyLines(trimmed); len(links) > 0 {
		return links, nil
	}

	// 4. base64 (с возможным двойным кодированием и переносами строк).
	decoded, ok := decodeSubscriptionBody(trimmed)
	if ok {
		if strings.HasPrefix(strings.TrimSpace(decoded), "{") || strings.HasPrefix(strings.TrimSpace(decoded), "[") {
			if links, ok := parseJSONSubscription(decoded); ok && len(links) > 0 {
				return links, nil
			}
		}
		if looksLikeClashYAML(decoded) {
			if links, ok := parseClashYAML(decoded); ok && len(links) > 0 {
				return links, nil
			}
		}
		if links := splitProxyLines(decoded); len(links) > 0 {
			return links, nil
		}
		// Двойное base64.
		if decoded2, ok2 := decodeSubscriptionBody(decoded); ok2 {
			if links := splitProxyLines(decoded2); len(links) > 0 {
				return links, nil
			}
		}
	}

	// 5. Последняя попытка: вытащить любые URI-подобные куски из произвольного текста
	//    (некоторые провайдеры оборачивают список в HTML или комментарии).
	if links := scavengeProxyLinks(trimmed); len(links) > 0 {
		return links, nil
	}

	return nil, fmt.Errorf("не удалось распознать формат подписки (ни base64-список, ни Clash YAML, ни sing-box/Xray JSON, ни plain-URI)")
}

// decodeSubscriptionBody снимает base64-обёртку, игнорируя пробелы и переносы строк.
func decodeSubscriptionBody(payload string) (string, bool) {
	raw := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "").Replace(strings.TrimSpace(payload))
	if raw == "" {
		return "", false
	}
	return decodeBase64Loose(raw)
}

// splitProxyLines возвращает все строки, похожие на прокси-ссылки.
func splitProxyLines(payload string) []string {
	lines := strings.FieldsFunc(payload, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\x00'
	})
	out := make([]string, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Строка должна быть «чистой» ссылкой: обрезаем HTML-обёртку, если она есть.
		if i := strings.LastIndex(line, ">"); i >= 0 && i+1 < len(line) && strings.Contains(line[:i], "<") {
			line = strings.TrimSpace(line[i+1:])
		}
		if i := strings.Index(line, "</"); i > 0 {
			line = strings.TrimSpace(line[:i])
		}
		if !IsProxyLink(line) {
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, line)
	}
	return out
}

// scavengeProxyLinks ищет URI-подобные подстроки в произвольном тексте.
// Список схем берётся из supportedSchemes, чтобы новые алиасы подхватывались
// автоматически (раньше список был захардкожен и отставал от парсера).
func scavengeProxyLinks(text string) []string {
	needles := scavengeNeedles()
	candidates := make([]string, 0, 8)
	rest := text
	for {
		best := -1
		for _, s := range needles {
			if i := indexFold(rest, s); i >= 0 && (best < 0 || i < best) {
				best = i
			}
		}
		if best < 0 {
			break
		}
		rest = rest[best:]
		// Конец ссылки — перевод строки, кавычка, угловая скобка, пробел.
		end := len(rest)
		for i, r := range rest {
			if r == '\n' || r == '\r' || r == '"' || r == '\'' || r == '<' || r == '>' || r == ' ' || r == '\t' {
				end = i
				break
			}
		}
		cand := strings.TrimSpace(rest[:end])
		// Обрезаем хвост HTML-разметки, если тег закрывается без пробела.
		if i := strings.Index(cand, "</"); i > 0 {
			cand = cand[:i]
		}
		if IsProxyLink(cand) {
			candidates = append(candidates, cand)
		}
		rest = rest[minInt(end, len(rest)):]
	}
	return candidates
}

// scavengeNeedles строит список «схема://» из supportedSchemes, от длинных к коротким,
// чтобы hysteria2:// не перекрывался hysteria:// при поиске ближайшего совпадения.
func scavengeNeedles() []string {
	out := make([]string, 0, len(supportedSchemes))
	for _, s := range supportedSchemes {
		// ssr:// и http(s):// в свободном тексте дают ложные срабатывания на обычные URL,
		// поэтому в режиме поиска по произвольному тексту их не ищем.
		if s == "ssr" || s == "http" || s == "https" {
			continue
		}
		out = append(out, s+"://")
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func indexFold(s, sub string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(sub))
}

// ---------------------------------------------------------------------------
// Clash / Clash.Meta YAML
// ---------------------------------------------------------------------------

func looksLikeClashYAML(s string) bool {
	low := strings.ToLower(s)
	if !strings.Contains(low, "proxies:") {
		return false
	}
	return strings.Contains(low, "type:") || strings.Contains(low, "server:")
}

// parseClashYAML разбирает Clash / Clash.Meta / Stash YAML без внешних зависимостей.
//
// Поддерживаются оба стиля записи:
//
//	proxies:
//	  - name: NL                     # блочный стиль (многострочный объект)
//	    type: vless
//	    ws-opts:
//	      path: /ws
//
//	proxies:
//	  - {name: NL, type: vless}      # flow-стиль (объект в фигурных скобках)
//
// Вложенные блоки (ws-opts, grpc-opts, reality-opts, h2-opts, plugin-opts) сохраняются
// как «путь.ключ» (например ws-opts.path → p["ws-opts.path"]).
func parseClashYAML(src string) ([]string, bool) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var links []string
	inProxies := false
	proxiesIndent := -1
	var cur map[string]string
	curIndent := -1
	// Стек вложенных блоков: ключ блока → его отступ.
	type blockFrame struct {
		key    string
		indent int
	}
	var stack []blockFrame

	flush := func() {
		if cur != nil {
			if link, ok := cloakToURI(cur); ok {
				links = append(links, link)
			}
		}
		cur = nil
		curIndent = -1
		stack = stack[:0]
	}

	for _, raw := range lines {
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		trimmed := strings.TrimSpace(raw)

		if !inProxies {
			low := strings.ToLower(trimmed)
			if !strings.HasPrefix(low, "proxies:") {
				continue
			}
			inProxies = true
			proxiesIndent = indent
			rest := strings.TrimSpace(trimmed[len("proxies:"):])
			// flow-форма: proxies: [{...}, {...}]
			if strings.HasPrefix(rest, "[") {
				for _, obj := range splitFlowList(rest) {
					if m := parseFlowObject(obj); len(m) > 0 {
						if link, ok := cloakToURI(m); ok {
							links = append(links, link)
						}
					}
				}
				inProxies = false
			}
			continue
		}

		// Выход из блока proxies: (строка с отступом <= proxiesIndent и не элемент списка).
		if indent <= proxiesIndent && !strings.HasPrefix(trimmed, "-") {
			flush()
			inProxies = false
			continue
		}

		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			flush()
			cur = map[string]string{}
			curIndent = indent
			body := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
			if body == "" {
				continue
			}
			if strings.HasPrefix(body, "{") {
				// flow-стиль: - {name: A, type: vless, ...}
				for k, v := range parseFlowObject(body) {
					cur[k] = v
				}
				continue
			}
			if k, v, ok := splitYAMLKV(body); ok {
				cur[k] = v
			}
			continue
		}

		if cur == nil || indent <= curIndent && len(stack) == 0 {
			continue
		}

		k, v, ok := splitYAMLKV(trimmed)
		if !ok {
			continue
		}
		// Снимаем со стека блоки, которые уже закончились по отступу.
		for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		prefix := ""
		for _, f := range stack {
			prefix += f.key + "."
		}
		if v == "" {
			// Начало вложенного блока: следующая строка с большим отступом — его содержимое.
			// Но если такой же ключ уже есть ниже в стеке — это его повтор, а не вложенность.
			for len(stack) > 0 {
				last := stack[len(stack)-1]
				if last.key == k {
					stack = stack[:len(stack)-1]
					break
				}
				if last.indent < indent {
					break
				}
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, blockFrame{key: k, indent: indent})
			continue
		}
		cur[prefix+k] = v
	}
	flush()
	return links, len(links) > 0
}

// splitFlowList разбивает "[{...}, {...}]" на отдельные объекты верхнего уровня.
func splitFlowList(src string) []string {
	src = strings.TrimSpace(src)
	src = strings.TrimPrefix(src, "[")
	src = strings.TrimSuffix(src, "]")
	out := make([]string, 0, 4)
	depth := 0
	start := 0
	inQuote := byte(0)
	for i := 0; i < len(src); i++ {
		c := src[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			inQuote = c
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				if part := strings.TrimSpace(src[start:i]); part != "" {
					out = append(out, part)
				}
				start = i + 1
			}
		}
	}
	if part := strings.TrimSpace(src[start:]); part != "" {
		out = append(out, part)
	}
	return out
}

// parseFlowObject разбирает Clash-объект в flow-стиле "{name: A, ws-opts: {path: /ws}}" в плоскую карту.
// Вложенные объекты сохраняются как «родитель.ключ», массивы — через запятую.
func parseFlowObject(src string) map[string]string {
	src = strings.TrimSpace(src)
	src = strings.TrimPrefix(src, "{")
	src = strings.TrimSuffix(src, "}")
	res := map[string]string{}
	for _, part := range splitFlowList(src) {
		k, v, ok := splitYAMLKV(part)
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "{") {
			for nk, nv := range parseFlowObject(v) {
				res[k+"."+nk] = nv
			}
			continue
		}
		if strings.HasPrefix(v, "[") {
			v = strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
			v = strings.TrimSpace(v)
		}
		res[k] = strings.Trim(v, `"'`)
	}
	return res
}

func splitYAMLKV(s string) (string, string, bool) {
	i := strings.Index(s, ":")
	if i <= 0 {
		return "", "", false
	}
	k := strings.TrimSpace(s[:i])
	v := strings.TrimSpace(s[i+1:])
	v = stripYAMLComment(v)
	v = strings.Trim(v, `"'`)
	return strings.ToLower(k), v, true
}

func stripYAMLComment(v string) string {
	inSingle, inDouble := false, false
	for i, r := range v {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble && i > 0 && v[i-1] == ' ' {
				return strings.TrimSpace(v[:i])
			}
		}
	}
	return v
}

// cloakToURI конвертирует один Clash-прокси в URI, который понимает ParseProxyLink.
func cloakToURI(p map[string]string) (string, bool) {
	typ := strings.ToLower(strings.TrimSpace(p["type"]))
	server := strings.TrimSpace(p["server"])
	port := strings.TrimSpace(p["port"])
	name := strings.TrimSpace(p["name"])
	if name == "" {
		name = server
	}
	if server == "" || port == "" {
		return "", false
	}
	q := url.Values{}
	setIf := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			q.Set(k, strings.TrimSpace(v))
		}
	}
	frag := "#" + url.PathEscape(name)

	// Clash пишет tls: true / reality-opts:{public-key:...}; часть панелей — вложенным блоком.
	tlsFlag := strings.ToLower(strings.TrimSpace(p["tls"]))
	if strings.EqualFold(q.Get("security"), "true") {
		q.Set("security", "tls")
	}
	// ws-opts / grpc-opts / h2-opts — вложенные блоки Clash.Meta.
	wsPath := firstNonEmptyStr(p["ws-path"], p["ws-opts.path"], yamlNested(p["ws-opts"], "path"))
	wsHost := firstNonEmptyStr(
		p["ws-headers"],
		p["ws-opts.headers.host"],
		yamlNested(firstNonEmptyStr(p["ws-opts.headers"], p["ws-opts"]), "host"),
		yamlNested(p["ws-opts"], "host"),
	)
	grpcSvc := firstNonEmptyStr(p["grpc-service-name"], p["grpc-opts.grpc-service-name"], yamlNested(p["grpc-opts"], "grpc-service-name"))
	h2Path := firstNonEmptyStr(p["h2-path"], p["h2-opts.path"], yamlNested(p["h2-opts"], "path"))
	net := normalizedClashNetwork(p["network"])

	// security: Clash пишет tls/reality; часть панелей — просто true.
	security := strings.ToLower(strings.TrimSpace(firstNonEmptyStr(p["security"], tlsFlag)))
	switch security {
	case "true", "1", "yes":
		security = "tls"
	case "false", "0", "no", "":
		security = ""
	}
	if security == "" && tlsFlag == "true" {
		security = "tls"
	}
	// security=reality без ключей Reality бессмыслен — откатываемся на tls.
	if security == "reality" {
		if pk := firstNonEmptyStr(p["public-key"], p["pbk"], p["reality-opts.public-key"], yamlNested(p["reality-opts"], "public-key")); pk == "" {
			security = "tls"
		}
	}

	switch typ {
	case "vless":
		setIf("type", net)
		setIf("security", security)
		setIf("sni", firstNonEmptyStr(p["servername"], p["sni"], p["reality-opts.server-name"], yamlNested(p["reality-opts"], "server-name")))
		setIf("fp", p["client-fingerprint"])
		setIf("flow", p["flow"])
		setIf("pbk", firstNonEmptyStr(p["public-key"], p["reality-opts.public-key"], yamlNested(p["reality-opts"], "public-key")))
		setIf("sid", firstNonEmptyStr(p["short-id"], p["reality-opts.short-id"], yamlNested(p["reality-opts"], "short-id")))
		setIf("alpn", p["alpn"])
		setIf("path", firstNonEmptyStr(wsPath, h2Path))
		setIf("host", wsHost)
		setIf("serviceName", grpcSvc)
		return fmt.Sprintf("vless://%s@%s:%s?%s%s", p["uuid"], server, port, q.Encode(), frag), p["uuid"] != ""

	case "vmess":
		obj := map[string]any{
			"v": "2", "ps": name, "add": server, "port": port,
			"id": p["uuid"], "aid": defStr(p["alterid"], "0"), "scy": defStr(p["cipher"], "auto"),
			"net": defStr(net, "tcp"), "type": "none", "host": wsHost,
			"path": firstNonEmptyStr(wsPath, h2Path), "tls": tlsFlag, "sni": firstNonEmptyStr(p["servername"], p["sni"]),
		}
		if a := strings.TrimSpace(p["alterid"]); a != "" {
			if n, err := strconv.Atoi(a); err == nil {
				obj["aid"] = strconv.Itoa(n)
			}
		}
		raw, err := json.Marshal(obj)
		if err != nil {
			return "", false
		}
		return "vmess://" + base64.StdEncoding.EncodeToString(raw), true

	case "trojan":
		setIf("sni", firstNonEmptyStr(p["sni"], p["servername"]))
		setIf("type", net)
		setIf("path", firstNonEmptyStr(wsPath, h2Path))
		setIf("host", wsHost)
		setIf("serviceName", grpcSvc)
		setIf("alpn", p["alpn"])
		setIf("fp", p["client-fingerprint"])
		setIf("security", defStr(p["tls"], "tls"))
		return fmt.Sprintf("trojan://%s@%s:%s?%s%s", p["password"], server, port, q.Encode(), frag), p["password"] != ""

	case "ss", "shadowsocks":
		cipher := strings.TrimSpace(p["cipher"])
		pass := strings.TrimSpace(p["password"])
		if cipher == "" || pass == "" {
			return "", false
		}
		plugin := strings.TrimSpace(p["plugin"])
		// Clash.Meta пишет плагин вложенным блоком plugin-opts: {mode: …, host: …}.
		// Значение «obfs;/» больше не годится: собираем opts отдельно и склеиваем ровно один раз.
		opts := ""
		switch {
		case strings.TrimSpace(p["plugin-opts"]) != "":
			opts = strings.TrimSpace(p["plugin-opts"])
		default:
			mode := firstNonEmptyStr(p["plugin-opts.mode"], yamlNested(p["plugin-opts"], "mode"))
			host := firstNonEmptyStr(p["plugin-opts.host"], yamlNested(p["plugin-opts"], "host"))
			if mode != "" {
				opts = "mode=" + mode
			}
			if host != "" {
				if opts != "" {
					opts += ";"
				}
				opts += "host=" + host
			}
		}
		if opts != "" {
			if plugin == "" {
				// Плагин не назван — по режиму восстанавливаем имя (obfs / v2ray-plugin).
				plugin = "obfs"
			}
			plugin = plugin + ";" + opts
		}
		setIf("plugin", plugin)
		setIf("udp-over-tcp", p["udp-over-tcp"])
		return fmt.Sprintf("ss://%s:%s@%s:%s?%s%s",
			url.PathEscape(cipher), url.PathEscape(pass), server, port, q.Encode(), frag), true

	case "hysteria2", "hy2":
		setIf("sni", firstNonEmptyStr(p["sni"], p["servername"]))
		setIf("obfs", p["obfs"])
		setIf("obfs-password", p["obfs-password"])
		if p["skip-cert-verify"] == "true" {
			q.Set("insecure", "1")
		}
		return fmt.Sprintf("hysteria2://%s@%s:%s?%s%s", p["password"], server, port, q.Encode(), frag), p["password"] != ""

	case "hysteria":
		setIf("sni", firstNonEmptyStr(p["sni"], p["servername"]))
		setIf("obfs", p["obfs"])
		setIf("protocol", p["protocol"])
		setIf("auth", p["auth-str"])
		setIf("upmbps", p["up"])
		setIf("downmbps", p["down"])
		if p["skip-cert-verify"] == "true" {
			q.Set("insecure", "1")
		}
		return fmt.Sprintf("hysteria://%s:%s?%s%s", server, port, q.Encode(), frag), true

	case "tuic":
		q.Set("sni", firstNonEmptyStr(p["sni"], p["servername"]))
		setIf("alpn", p["alpn"])
		setIf("congestion_control", p["congestion-controller"])
		setIf("udp_relay_mode", p["udp-relay-mode"])
		if p["skip-cert-verify"] == "true" {
			q.Set("insecure", "1")
		}
		u := fmt.Sprintf("tuic://%s:%s@%s:%s?%s%s", url.PathEscape(p["uuid"]), url.PathEscape(p["password"]), server, port, q.Encode(), frag)
		return u, p["uuid"] != ""

	case "anytls":
		setIf("sni", firstNonEmptyStr(p["sni"], p["servername"]))
		setIf("fp", p["client-fingerprint"])
		if p["skip-cert-verify"] == "true" {
			q.Set("insecure", "1")
		}
		return fmt.Sprintf("anytls://%s@%s:%s?%s%s", url.PathEscape(p["password"]), server, port, q.Encode(), frag), p["password"] != ""

	case "socks5", "socks":
		user := p["username"]
		pass := p["password"]
		return fmt.Sprintf("socks5://%s:%s@%s:%s%s", url.PathEscape(user), url.PathEscape(pass), server, port, frag), true

	case "http":
		scheme := "http"
		if tlsFlag == "true" {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s:%s@%s:%s%s", scheme, url.PathEscape(p["username"]), url.PathEscape(p["password"]), server, port, frag), true

	case "wireguard":
		priv := firstNonEmptyStr(p["private-key"], p["private_key"])
		pub := firstNonEmptyStr(p["public-key"], p["public_key"])
		if priv == "" || pub == "" {
			return "", false
		}
		setIf("private_key", priv)
		setIf("public_key", pub)
		setIf("preshared_key", firstNonEmptyStr(p["pre-shared-key"], p["preshared-key"]))
		setIf("address", firstNonEmptyStr(p["ip"], p["address"]))
		setIf("mtu", p["mtu"])
		return fmt.Sprintf("wireguard://%s:%s?%s%s", server, port, q.Encode(), frag), true
	}
	return "", false
}

// firstNonEmptyStr возвращает первое непустое значение (удобно для нескольких вариантов ключа).
func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// normalizedClashNetwork приводит название транспорта Clash к привычному виду sing-box/Xray.
func normalizedClashNetwork(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return ""
	case "h2", "http":
		return "http"
	case "grpc":
		return "grpc"
	case "ws", "websocket":
		return "ws"
	case "tcp", "raw":
		return ""
	default:
		return strings.ToLower(strings.TrimSpace(v))
	}
}

// yamlNested достаёт значение вложенного блока Clash, записанного одной строкой:
//
//	ws-opts: { path: /ws, headers: { Host: example.com } }
//
// Регистр ключей не важен. Возвращает "" если ключ не найден.
func yamlNested(block, key string) string {
	block = strings.TrimSpace(block)
	if block == "" {
		return ""
	}
	block = strings.TrimPrefix(block, "{")
	block = strings.TrimSuffix(block, "}")
	// Разбираем по запятым верхнего уровня, учитывая вложенность фигурных скобок.
	depth := 0
	start := 0
	lowKey := strings.ToLower(key)
	for i := 0; i <= len(block); i++ {
		atEnd := i == len(block)
		if !atEnd {
			switch block[i] {
			case '{':
				depth++
				continue
			case '}':
				depth--
				continue
			case ',':
				if depth != 0 {
					continue
				}
			default:
				continue
			}
		}
		part := strings.TrimSpace(block[start:i])
		start = i + 1
		if part == "" {
			continue
		}
		if k, v, ok := splitYAMLKV(part); ok && k == lowKey {
			return strings.Trim(v, "{}[]\"' ")
		}
	}
	return ""
}

func defStr(v, fallback string) string {
	if s := strings.TrimSpace(v); s != "" {
		return s
	}
	return fallback
}

// ---------------------------------------------------------------------------
// sing-box / Xray / generic JSON
// ---------------------------------------------------------------------------

func parseJSONSubscription(src string) ([]string, bool) {
	src = strings.TrimSpace(src)

	// Массив строк: ["vless://...", ...]
	var arr []any
	if strings.HasPrefix(src, "[") {
		if err := json.Unmarshal([]byte(src), &arr); err == nil {
			strs := make([]string, 0, len(arr))
			for _, it := range arr {
				if s, ok := it.(string); ok && IsProxyLink(s) {
					strs = append(strs, s)
				}
			}
			if len(strs) > 0 {
				return strs, true
			}
		}
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(src), &root); err != nil {
		return nil, false
	}

	// {"proxies":[{...Clash-like...}]} — некоторые панели отдают JSON вместо YAML.
	if rawProxies, ok := root["proxies"].([]any); ok {
		links := make([]string, 0, len(rawProxies))
		for _, it := range rawProxies {
			if m, ok := it.(map[string]any); ok {
				flat := map[string]string{}
				for k, v := range m {
					flat[strings.ToLower(k)] = jsonScalar(v)
				}
				if link, ok := cloakToURI(flat); ok {
					links = append(links, link)
				}
			}
		}
		if len(links) > 0 {
			return links, true
		}
	}

	// outbounds: [...] — sing-box или Xray.
	rawOutbounds, hasOutbounds := root["outbounds"].([]any)
	if !hasOutbounds {
		return nil, false
	}
	links := make([]string, 0, len(rawOutbounds))
	for _, it := range rawOutbounds {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if link, ok := outboundToURI(m); ok {
			links = append(links, link)
		}
	}
	return links, len(links) > 0
}

func jsonScalar(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case []any:
		parts := make([]string, 0, len(t))
		for _, it := range t {
			parts = append(parts, jsonScalar(it))
		}
		return strings.Join(parts, ",")
	case map[string]any:
		parts := make([]string, 0, len(t))
		for _, it := range t {
			parts = append(parts, jsonScalar(it))
		}
		return strings.Join(parts, ",")
	}
	return ""
}

// outboundToURI конвертирует outbound sing-box/Xray в URI-строку.
func outboundToURI(m map[string]any) (string, bool) {
	typ := strings.ToLower(jsonScalar(m["type"]))
	if typ == "" {
		typ = strings.ToLower(jsonScalar(m["protocol"])) // Xray-формат
	}
	tag := jsonScalar(m["tag"])
	if tag == "" {
		tag = jsonScalar(m["remarks"])
	}
	server := jsonScalar(m["server"])
	port := jsonScalar(m["server_port"])
	if port == "" {
		port = jsonScalar(m["port"])
	}
	// Xray streamSettings → те же поля, что даёт sing-box TLS/transport.
	if ss, ok := m["streamSettings"].(map[string]any); ok {
		if m["tls"] == nil {
			if sec := strings.ToLower(jsonScalar(ss["security"])); sec == "tls" || sec == "reality" {
				tlsObj := map[string]any{"enabled": true}
				tlsSrc, _ := ss[sec+"Settings"].(map[string]any)
				if tlsSrc == nil {
					tlsSrc, _ = ss["tlsSettings"].(map[string]any)
				}
				for _, k := range []string{"serverName", "server_name"} {
					if v := jsonScalar(tlsSrc[k]); v != "" {
						tlsObj["server_name"] = v
					}
				}
				if v := jsonScalar(tlsSrc["fingerprint"]); v != "" {
					tlsObj["utls"] = map[string]any{"enabled": true, "fingerprint": v}
				}
				if v := jsonScalar(tlsSrc["publicKey"]); v != "" {
					tlsObj["reality"] = map[string]any{"enabled": true, "public_key": v, "short_id": jsonScalar(tlsSrc["shortId"])}
				}
				if v := jsonScalar(tlsSrc["allowInsecure"]); v == "true" {
					tlsObj["insecure"] = true
				}
				if alpn, ok := tlsSrc["alpn"].([]any); ok && len(alpn) > 0 {
					tlsObj["alpn"] = alpn
				}
				m["tls"] = tlsObj
			}
		}
		if m["transport"] == nil {
			if net := strings.ToLower(jsonScalar(ss["network"])); net != "" && net != "tcp" {
				tr := map[string]any{"type": net}
				if net == "ws" {
					if ws, ok := ss["wsSettings"].(map[string]any); ok {
						tr["path"] = jsonScalar(ws["path"])
						if h, ok := ws["headers"].(map[string]any); ok {
							tr["headers"] = map[string]any{"Host": jsonScalar(h["Host"])}
						}
					}
				}
				if net == "grpc" {
					if g, ok := ss["grpcSettings"].(map[string]any); ok {
						tr["service_name"] = jsonScalar(g["serviceName"])
					}
				}
				m["transport"] = tr
			}
		}
	}
	// Xray: settings.vnext[0].address / settings.servers[0].address
	if settings, ok := m["settings"].(map[string]any); ok {
		if vnext, ok := settings["vnext"].([]any); ok && len(vnext) > 0 {
			if first, ok := vnext[0].(map[string]any); ok {
				server = jsonScalar(first["address"])
				port = jsonScalar(first["port"])
				if users, ok := first["users"].([]any); ok && len(users) > 0 {
					if u, ok := users[0].(map[string]any); ok {
						if id := jsonScalar(u["id"]); id != "" {
							m["uuid"] = id
						}
						if fl := jsonScalar(u["flow"]); fl != "" {
							m["flow"] = fl
						}
					}
				}
			}
		}
		if servers, ok := settings["servers"].([]any); ok && len(servers) > 0 {
			if first, ok := servers[0].(map[string]any); ok {
				server = jsonScalar(first["address"])
				port = jsonScalar(first["port"])
				if pw := jsonScalar(first["password"]); pw != "" {
					m["password"] = pw
				}
				if mt := jsonScalar(first["method"]); mt != "" {
					m["method"] = mt
				}
				if users, ok := first["users"].([]any); ok && len(users) > 0 {
					if u, ok := users[0].(map[string]any); ok {
						if pw := jsonScalar(u["password"]); pw != "" {
							m["password"] = pw
						}
						if us := jsonScalar(u["user"]); us != "" {
							m["username"] = us
						}
					}
				}
			}
		}
	}
	if server == "" || port == "" || isDummyHost(server) {
		return "", false
	}
	q := url.Values{}
	setQ := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			q.Set(k, strings.TrimSpace(v))
		}
	}
	if tr, ok := m["transport"].(map[string]any); ok {
		setQ("type", jsonScalar(tr["type"]))
		setQ("path", jsonScalar(tr["path"]))
		setQ("serviceName", jsonScalar(tr["service_name"]))
		if h, ok := tr["headers"].(map[string]any); ok {
			setQ("host", jsonScalar(h["Host"]))
		}
		if hosts, ok := tr["host"].([]any); ok && len(hosts) > 0 {
			setQ("host", jsonScalar(hosts[0]))
		}
	}
	if tls, ok := m["tls"].(map[string]any); ok {
		enabled := jsonScalar(tls["enabled"])
		if enabled == "true" || enabled == "" {
			setQ("security", "tls")
		}
		setQ("sni", jsonScalar(tls["server_name"]))
		if alpn, ok := tls["alpn"].([]any); ok && len(alpn) > 0 {
			parts := make([]string, 0, len(alpn))
			for _, a := range alpn {
				parts = append(parts, jsonScalar(a))
			}
			setQ("alpn", strings.Join(parts, ","))
		}
		if utls, ok := tls["utls"].(map[string]any); ok {
			setQ("fp", jsonScalar(utls["fingerprint"]))
		}
		if reality, ok := tls["reality"].(map[string]any); ok {
			setQ("security", "reality")
			setQ("pbk", jsonScalar(reality["public_key"]))
			setQ("sid", jsonScalar(reality["short_id"]))
		}
		if jsonScalar(tls["insecure"]) == "true" {
			q.Set("insecure", "1")
		}
	}
	frag := ""
	if tag != "" {
		frag = "#" + url.PathEscape(tag)
	}

	switch typ {
	case "vless":
		setQ("flow", jsonScalar(m["flow"]))
		return fmt.Sprintf("vless://%s@%s:%s?%s%s", jsonScalar(m["uuid"]), server, port, q.Encode(), frag), jsonScalar(m["uuid"]) != ""
	case "vmess":
		obj := map[string]any{
			"v": "2", "ps": tag, "add": server, "port": port,
			"id": jsonScalar(m["uuid"]), "aid": defStr(jsonScalar(m["alter_id"]), "0"),
			"scy": defStr(jsonScalar(m["security"]), "auto"),
			"net": defStr(q.Get("type"), "tcp"), "path": q.Get("path"), "host": q.Get("host"),
		}
		raw, err := json.Marshal(obj)
		if err != nil {
			return "", false
		}
		return "vmess://" + base64.StdEncoding.EncodeToString(raw), obj["id"] != ""
	case "trojan":
		return fmt.Sprintf("trojan://%s@%s:%s?%s%s", jsonScalar(m["password"]), server, port, q.Encode(), frag), jsonScalar(m["password"]) != ""
	case "shadowsocks":
		method := jsonScalar(m["method"])
		pass := jsonScalar(m["password"])
		if method == "" || pass == "" {
			return "", false
		}
		return fmt.Sprintf("ss://%s:%s@%s:%s?%s%s", url.PathEscape(method), url.PathEscape(pass), server, port, q.Encode(), frag), true
	case "hysteria2":
		setQ("obfs-password", jsonScalar(func() string {
			if obfs, ok := m["obfs"].(map[string]any); ok {
				return jsonScalar(obfs["password"])
			}
			return ""
		}()))
		return fmt.Sprintf("hysteria2://%s@%s:%s?%s%s", jsonScalar(m["password"]), server, port, q.Encode(), frag), jsonScalar(m["password"]) != ""
	case "tuic":
		return fmt.Sprintf("tuic://%s:%s@%s:%s?%s%s", url.PathEscape(jsonScalar(m["uuid"])), url.PathEscape(jsonScalar(m["password"])), server, port, q.Encode(), frag), jsonScalar(m["uuid"]) != ""
	case "anytls":
		return fmt.Sprintf("anytls://%s@%s:%s?%s%s", url.PathEscape(jsonScalar(m["password"])), server, port, q.Encode(), frag), jsonScalar(m["password"]) != ""
	case "socks":
		return fmt.Sprintf("socks5://%s:%s@%s:%s%s", url.PathEscape(jsonScalar(m["username"])), url.PathEscape(jsonScalar(m["password"])), server, port, frag), true
	case "http":
		return fmt.Sprintf("http://%s:%s@%s:%s%s", url.PathEscape(jsonScalar(m["username"])), url.PathEscape(jsonScalar(m["password"])), server, port, frag), true
	}
	return "", false
}
